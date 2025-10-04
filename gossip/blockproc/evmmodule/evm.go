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

package evmmodule

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/gossip/blockproc"
	"github.com/panoptisDev/pano/gossip/gasprice"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
)

//go:generate mockgen -source=evm.go -destination=evm_mock.go -package=evmmodule

type EVMModule struct{}

func New() *EVMModule {
	return &EVMModule{}
}

func (p *EVMModule) Start(
	block iblockproc.BlockCtx,
	statedb state.StateDB,
	reader evmcore.DummyChain,
	onNewLog func(*types.Log),
	rules opera.Rules,
	evmCfg *params.ChainConfig,
	prevrandao common.Hash,
) blockproc.EVMProcessor {
	var prevBlockHash common.Hash
	var baseFee *big.Int
	if block.Idx == 0 {
		baseFee = gasprice.GetInitialBaseFee(rules.Economy)
	} else {
		header := reader.GetHeader(common.Hash{}, uint64(block.Idx-1))
		prevBlockHash = header.Hash
		baseFee = gasprice.GetBaseFeeForNextBlock(gasprice.ParentBlockInfo{
			BaseFee:  header.BaseFee,
			Duration: header.Duration,
			GasUsed:  header.GasUsed,
		}, rules.Economy)
	}

	// Start block
	statedb.BeginBlock(uint64(block.Idx))

	return &OperaEVMProcessor{
		block:            block,
		reader:           reader,
		statedb:          statedb,
		onNewLog:         onNewLog,
		rules:            rules,
		evmCfg:           evmCfg,
		blockIdx:         uint64(block.Idx),
		prevBlockHash:    prevBlockHash,
		prevRandao:       prevrandao,
		gasBaseFee:       baseFee,
		processorFactory: stateProcessorFactory{},
	}
}

type OperaEVMProcessor struct {
	block    iblockproc.BlockCtx
	reader   evmcore.DummyChain
	statedb  state.StateDB
	onNewLog func(*types.Log)
	rules    opera.Rules
	evmCfg   *params.ChainConfig

	blockIdx      uint64
	prevBlockHash common.Hash
	gasBaseFee    *big.Int

	gasUsed uint64

	processedTxs []evmcore.ProcessedTransaction
	prevRandao   common.Hash

	processorFactory _stateProcessorFactory
}

func (p *OperaEVMProcessor) evmBlockWith(txs types.Transactions) *evmcore.EvmBlock {
	baseFee := p.rules.Economy.MinGasPrice
	if !p.rules.Upgrades.London {
		baseFee = nil
	} else if p.rules.Upgrades.Pano {
		baseFee = p.gasBaseFee
	}

	prevRandao := common.Hash{}
	// This condition must be kept, otherwise Pano will not be able to synchronize
	if p.rules.Upgrades.Pano {
		prevRandao = p.prevRandao
	}

	var withdrawalsHash *common.Hash = nil
	if p.rules.Upgrades.Pano {
		withdrawalsHash = &types.EmptyWithdrawalsHash
	}

	blobBaseFee := evmcore.GetBlobBaseFee()
	h := &evmcore.EvmHeader{
		Number:          new(big.Int).SetUint64(p.blockIdx),
		ParentHash:      p.prevBlockHash,
		Root:            common.Hash{}, // state root is added later
		Time:            p.block.Time,
		Coinbase:        evmcore.GetCoinbase(),
		GasLimit:        p.rules.Blocks.MaxBlockGas,
		GasUsed:         p.gasUsed,
		BaseFee:         baseFee,
		BlobBaseFee:     blobBaseFee.ToBig(),
		PrevRandao:      prevRandao,
		WithdrawalsHash: withdrawalsHash,
		Epoch:           p.block.Atropos.Epoch(),
	}

	return evmcore.NewEvmBlock(h, txs)
}

func (p *OperaEVMProcessor) Execute(txs types.Transactions, gasLimit uint64) []evmcore.ProcessedTransaction {
	evmProcessor := p.processorFactory.NewStateProcessor(p.evmCfg, p.reader, p.rules.Upgrades)
	txsOffset := uint(len(p.processedTxs))

	vmConfig := opera.GetVmConfig(p.rules)

	// Process txs
	evmBlock := p.evmBlockWith(txs)
	processed := evmProcessor.Process(evmBlock, p.statedb, vmConfig, gasLimit, &p.gasUsed, func(l *types.Log) {
		// Note: l.Index is properly set before
		l.TxIndex += txsOffset
		p.onNewLog(l)
	})

	if txsOffset > 0 {
		for _, p := range processed {
			if p.Receipt != nil {
				p.Receipt.TransactionIndex += txsOffset
			}
		}
	}

	p.processedTxs = append(p.processedTxs, processed...)

	return processed
}

func (p *OperaEVMProcessor) Finalize() (evmBlock *evmcore.EvmBlock, numSkipped int, receipts types.Receipts) {
	transactions := make(types.Transactions, 0, len(p.processedTxs))
	receipts = make(types.Receipts, 0, len(p.processedTxs))
	for _, tx := range p.processedTxs {
		if tx.Receipt != nil {
			transactions = append(transactions, tx.Transaction)
			receipts = append(receipts, tx.Receipt)
		} else {
			numSkipped++
		}
	}

	evmBlock = p.evmBlockWith(transactions)

	// Commit block
	p.statedb.EndBlock(evmBlock.Number.Uint64())

	// Get state root
	evmBlock.Root = p.statedb.GetStateHash()

	return
}

// _stateProcessorFactory is an internal interface to allow introducing mocked
// state processors in tests.
type _stateProcessorFactory interface {
	NewStateProcessor(
		evmCfg *params.ChainConfig,
		reader evmcore.DummyChain,
		upgrades opera.Upgrades,
	) _stateProcessor
}

// _stateProcessor is an internal interface to allow introducing mocked
// state processors in tests.
type _stateProcessor interface {
	Process(
		block *evmcore.EvmBlock,
		statedb state.StateDB,
		vmCfg vm.Config,
		gasLimit uint64,
		gasUsed *uint64,
		onNewLog func(*types.Log),
	) []evmcore.ProcessedTransaction
}

// stateProcessorFactory is the production implementation of the
// _stateProcessorFactory using the real evmcore.StateProcessor.
type stateProcessorFactory struct{}

func (stateProcessorFactory) NewStateProcessor(
	evmCfg *params.ChainConfig,
	reader evmcore.DummyChain,
	upgrades opera.Upgrades,
) _stateProcessor {
	return evmcore.NewStateProcessor(evmCfg, reader, upgrades)
}
