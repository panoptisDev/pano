// Copyright 2025 Pano Operations Ltd
// This file is part of the Pano Client
//
// Pano is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Pano is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Pano. If not, see <http://www.gnu.org/licenses/>.

package makegenesis

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/pano/utils/objstream"
	"github.com/ethereum/go-ethereum/core/tracing"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/gossip/blockproc"
	"github.com/panoptisDev/pano/gossip/blockproc/drivermodule"
	"github.com/panoptisDev/pano/gossip/blockproc/eventmodule"
	"github.com/panoptisDev/pano/gossip/blockproc/evmmodule"
	"github.com/panoptisDev/pano/gossip/blockproc/sealmodule"
	"github.com/panoptisDev/pano/gossip/evmstore"
	"github.com/panoptisDev/pano/gossip/gasprice"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/inter/ibr"
	"github.com/panoptisDev/pano/inter/ier"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/opera/genesis"
	"github.com/panoptisDev/pano/opera/genesisstore"
	"github.com/panoptisDev/pano/utils"

	mptIo "github.com/panoptisDev/carmen/go/database/mpt/io"
	carmen "github.com/panoptisDev/carmen/go/state"
)

type GenesisBuilder struct {
	tmpStateDB    state.StateDB
	carmenDir     string
	carmenStateDb carmen.StateDB

	totalSupply *big.Int

	blocks       []ibr.LlrIdxFullBlockRecord
	epochs       []ier.LlrIdxFullEpochRecord
	currentEpoch ier.LlrIdxFullEpochRecord

	genesisCommitteeCertificate cert.CommitteeCertificate
	genesisBlockCertificates    []cert.BlockCertificate
}

type BlockProc struct {
	SealerModule     blockproc.SealerModule
	TxListenerModule blockproc.TxListenerModule
	PreTxTransactor  blockproc.TxTransactor
	PostTxTransactor blockproc.TxTransactor
	EventsModule     blockproc.ConfirmedEventsModule
	EVMModule        blockproc.EVM
}

func DefaultBlockProc() BlockProc {
	return BlockProc{
		SealerModule:     sealmodule.New(),
		TxListenerModule: drivermodule.NewDriverTxListenerModule(),
		PreTxTransactor:  drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor: drivermodule.NewDriverTxTransactor(),
		EventsModule:     eventmodule.New(),
		EVMModule:        evmmodule.New(),
	}
}

func (b *GenesisBuilder) AddBalance(acc common.Address, balance *big.Int) {
	if len(b.blocks) > 0 {
		panic("cannot add balance after block zero is finalized")
	}
	b.tmpStateDB.AddBalance(acc, utils.BigIntToUint256(balance), tracing.BalanceIncreaseGenesisBalance)
	b.totalSupply.Add(b.totalSupply, balance)
}

func (b *GenesisBuilder) SetCode(acc common.Address, code []byte) {
	if len(b.blocks) > 0 {
		panic("cannot add code after block zero is finalized")
	}
	b.tmpStateDB.SetCode(acc, code)
}

func (b *GenesisBuilder) SetNonce(acc common.Address, nonce uint64) {
	if len(b.blocks) > 0 {
		panic("cannot add nonce after block zero is finalized")
	}
	b.tmpStateDB.SetNonce(acc, nonce, tracing.NonceChangeGenesis)
}

func (b *GenesisBuilder) SetStorage(acc common.Address, key, val common.Hash) {
	if len(b.blocks) > 0 {
		panic("cannot set storage after block zero is finalized")
	}
	b.tmpStateDB.SetState(acc, key, val)
}

func (b *GenesisBuilder) SetCurrentEpoch(er ier.LlrIdxFullEpochRecord) {
	b.currentEpoch = er
}

func (b *GenesisBuilder) TotalSupply() *big.Int {
	return b.totalSupply
}

func (b *GenesisBuilder) CurrentHash() hash.Hash {
	er := b.epochs[len(b.epochs)-1]
	return er.Hash()
}

func NewGenesisBuilder() *GenesisBuilder {
	carmenDir, err := os.MkdirTemp("", "opera-tmp-genesis")
	if err != nil {
		panic(fmt.Errorf("failed to create temporary dir for GenesisBuilder: %v", err))
	}
	carmenState, err := carmen.NewState(carmen.Parameters{
		Variant:      "go-file",
		Schema:       carmen.Schema(5),
		Archive:      carmen.S5Archive,
		Directory:    carmenDir,
		LiveCache:    1, // use minimum cache (not default)
		ArchiveCache: 1, // use minimum cache (not default)
	})
	if err != nil {
		panic(fmt.Errorf("failed to create carmen state; %s", err))
	}
	// Set cache size to lowest value possible
	carmenStateDb := carmen.CreateCustomStateDBUsing(carmenState, 1024)
	tmpStateDB := evmstore.CreateCarmenStateDb(carmenStateDb)
	return &GenesisBuilder{
		tmpStateDB:    tmpStateDB,
		carmenDir:     carmenDir,
		carmenStateDb: carmenStateDb,
		totalSupply:   new(big.Int),
	}
}

type dummyHeaderReturner struct {
	blocks []ibr.LlrIdxFullBlockRecord
}

func (d dummyHeaderReturner) GetHeader(_ common.Hash, position uint64) *evmcore.EvmHeader {
	if position < uint64(len(d.blocks)) {
		return &evmcore.EvmHeader{
			BaseFee: d.blocks[position].BaseFee,
		}
	}
	return &evmcore.EvmHeader{
		BaseFee: big.NewInt(0),
	}
}

// FinalizeBlockZero finalizes the genesis block 0 by computing the state root hash of
// the initial block and filling in other block header information. This function must
// be called before ExecuteGenesisTxs.
func (b *GenesisBuilder) FinalizeBlockZero(
	rules opera.Rules,
	genesisTime inter.Timestamp,
) (
	blockHash common.Hash,
	stateRoot common.Hash,
	err error,
) {
	if len(b.blocks) > 0 {
		return common.Hash{}, common.Hash{}, errors.New("block zero already finalized")
	}

	if rules.Upgrades.Allegro {
		if err := rules.Validate(opera.Rules{}); err != nil {
			return common.Hash{}, common.Hash{}, fmt.Errorf("invalid rules: %w", err)
		}
	}

	// construct state root of initial state
	b.tmpStateDB.EndBlock(0)
	genesisStateRoot := b.tmpStateDB.GetStateHash()

	// construct the block record for the genesis block
	blockBuilder := inter.NewBlockBuilder().
		WithEpoch(1).
		WithNumber(0).
		WithParentHash(common.Hash{}).
		WithStateRoot(genesisStateRoot).
		WithTime(genesisTime).
		WithDuration(0).
		WithGasLimit(rules.Blocks.MaxBlockGas).
		WithGasUsed(0).
		WithBaseFee(gasprice.GetInitialBaseFee(rules.Economy)).
		WithPrevRandao(common.Hash{31: 1})

	block := blockBuilder.Build()
	llrBlock := ibr.FullBlockRecordFor(block, nil, nil)
	b.blocks = append(b.blocks, ibr.LlrIdxFullBlockRecord{
		LlrFullBlockRecord: *llrBlock,
		Idx:                0,
	})

	// register an empty certificate for block zero
	b.AddBlockCertificate(cert.NewCertificate(cert.NewBlockStatement(
		rules.NetworkID,
		idx.Block(0),
		block.Hash(),
		block.StateRoot,
	)))

	return common.Hash(b.blocks[0].BlockHash), genesisStateRoot, nil
}

func (b *GenesisBuilder) ExecuteGenesisTxs(blockProc BlockProc, genesisTxs types.Transactions) error {
	if len(b.blocks) == 0 {
		return errors.New("no block zero - run FinalizeBlockZero first")
	}

	bs, es := b.currentEpoch.BlockState.Copy(), b.currentEpoch.EpochState.Copy()
	es.Rules.Economy.MinGasPrice = big.NewInt(0) // < needed since genesis transactions have gas price 0

	blockCtx := iblockproc.BlockCtx{
		Idx:     bs.LastBlock.Idx + 1,
		Time:    bs.LastBlock.Time + 1,
		Atropos: hash.Event{},
	}

	sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
	txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, b.tmpStateDB)
	chainConfig := opera.CreateTransientEvmChainConfig(
		es.Rules.NetworkID,
		// apply upgrades described in genesis rules, effect immediately
		[]opera.UpgradeHeight{{
			Upgrades: es.Rules.Upgrades,
			Height:   blockCtx.Idx,
		}},
		blockCtx.Idx,
	)
	evmProcessor := blockProc.EVMModule.Start(
		blockCtx, b.tmpStateDB, dummyHeaderReturner{b.blocks},
		func(l *types.Log) { txListener.OnNewLog(l) },
		es.Rules,
		chainConfig,
		common.Hash{0x01}, // non-zero PrevRandao necessary to enable Cancun
	)

	// Execute genesis transactions
	evmProcessor.Execute(genesisTxs, es.Rules.Blocks.MaxBlockGas)
	bs = txListener.Finalize()

	// Execute pre-internal transactions
	preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, true, b.tmpStateDB)
	evmProcessor.Execute(preInternalTxs, es.Rules.Blocks.MaxBlockGas)
	bs = txListener.Finalize()

	// Seal epoch
	sealer.Update(bs, es)
	bs, es = sealer.SealEpoch()
	txListener.Update(bs, es)

	// Execute post-internal transactions
	internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, true, b.tmpStateDB)
	evmProcessor.Execute(internalTxs, es.Rules.Blocks.MaxBlockGas)

	evmBlock, numSkippedTxs, receipts := evmProcessor.Finalize()
	for i, r := range receipts {
		if r.Status == 0 {
			return fmt.Errorf("genesis transaction %d of %d reverted", i, len(receipts))
		}
	}
	if numSkippedTxs != 0 {
		return fmt.Errorf("genesis transaction is skipped (num=%d)", numSkippedTxs)
	}
	bs = txListener.Finalize()
	bs.FinalizedStateRoot = hash.Hash(evmBlock.Root)

	bs.LastBlock = blockCtx

	receiptsStorage := make([]*types.ReceiptForStorage, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = (*types.ReceiptForStorage)(r)
	}

	// construct the block record for the genesis block
	blockBuilder := inter.NewBlockBuilder().
		WithEpoch(1).
		WithNumber(uint64(blockCtx.Idx)).
		WithParentHash(common.Hash(b.blocks[len(b.blocks)-1].BlockHash)).
		WithStateRoot(common.Hash(bs.FinalizedStateRoot)).
		WithTime(evmBlock.Time).
		WithDuration(1).
		WithGasLimit(evmBlock.GasLimit).
		WithGasUsed(evmBlock.GasUsed).
		WithBaseFee(evmBlock.BaseFee).
		WithPrevRandao(evmBlock.PrevRandao)

	for txIndex, transaction := range evmBlock.Transactions {
		if !bytes.Equal(transaction.Hash().Bytes(), receipts[txIndex].TxHash.Bytes()) {
			return fmt.Errorf("genesis transaction hash %d of %d does not match with the receipt",
				txIndex, len(receipts))
		}
		blockBuilder.AddTransaction(transaction, receipts[txIndex])
	}

	block := blockBuilder.Build()
	llrBlock := ibr.FullBlockRecordFor(block, evmBlock.Transactions, receiptsStorage)
	b.blocks = append(b.blocks, ibr.LlrIdxFullBlockRecord{
		LlrFullBlockRecord: *llrBlock,
		Idx:                blockCtx.Idx,
	})

	// add epochs
	b.epochs = append(b.epochs, b.currentEpoch) // safe epoch 1
	b.currentEpoch = ier.LlrIdxFullEpochRecord{ // create epoch 2
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: bs,
			EpochState: es,
		},
		Idx: es.Epoch,
	}
	b.epochs = append(b.epochs, b.currentEpoch)

	// add a block certificate for the created block
	b.AddBlockCertificate(cert.NewCertificate(cert.NewBlockStatement(
		es.Rules.NetworkID,
		idx.Block(block.Number),
		block.Hash(),
		block.StateRoot,
	)))

	return nil
}

func (b *GenesisBuilder) SetGenesisCommitteeCertificate(
	committeeCertificate cert.CommitteeCertificate,
) {
	b.genesisCommitteeCertificate = committeeCertificate
}

func (b *GenesisBuilder) AddBlockCertificate(
	blockCertificate cert.BlockCertificate,
) {
	b.genesisBlockCertificates = append(
		b.genesisBlockCertificates,
		blockCertificate,
	)
}

type memFile struct {
	*bytes.Buffer
}

func (f *memFile) Close() error {
	*f = memFile{}
	return nil
}

func (b *GenesisBuilder) Build(head genesis.Header) *genesisstore.Store {
	err := b.carmenStateDb.Close()
	if err != nil {
		panic(fmt.Errorf("failed to close genesis carmen state; %s", err))
	}
	return genesisstore.NewStore(func(name string) (io.Reader, error) {
		buf := &memFile{bytes.NewBuffer(nil)}
		if name == genesisstore.BlocksSection(0) {
			for i := len(b.blocks) - 1; i >= 0; i-- {
				_ = rlp.Encode(buf, b.blocks[i])
			}
			return buf, nil
		}
		if name == genesisstore.EpochsSection(0) {
			for i := len(b.epochs) - 1; i >= 0; i-- {
				_ = rlp.Encode(buf, b.epochs[i])
			}
			return buf, nil
		}
		if name == genesisstore.FwsLiveSection(0) {
			err := mptIo.Export(context.Background(), mptIo.NewLog(), filepath.Join(b.carmenDir, "live"), buf)
			if err != nil {
				return nil, err
			}
		}
		if name == genesisstore.FwsArchiveSection(0) {
			err := mptIo.ExportArchive(context.Background(), mptIo.NewLog(), filepath.Join(b.carmenDir, "archive"), buf)
			if err != nil {
				return nil, err
			}
		}
		if name == genesisstore.SccCommitteeSection(0) {
			out := objstream.NewWriter[cert.Certificate[cert.CommitteeStatement]](buf)
			err := out.Write(b.genesisCommitteeCertificate)
			if err != nil {
				return nil, err
			}
		}
		if name == genesisstore.SccBlockSection(0) {
			out := objstream.NewWriter[cert.Certificate[cert.BlockStatement]](buf)
			for _, bc := range b.genesisBlockCertificates {
				if err := out.Write(bc); err != nil {
					return nil, err
				}
			}
			return buf, nil
		}
		if buf.Len() == 0 {
			return nil, errors.New("not found")
		}
		return buf, nil
	}, head, func() error {
		err := os.RemoveAll(b.carmenDir)
		*b = GenesisBuilder{}
		return err
	})
}
