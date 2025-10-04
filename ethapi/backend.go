// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package ethapi implements the general Ethereum API functions.
package ethapi

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
)

// PeerProgress is synchronization status of a peer
type PeerProgress struct {
	CurrentEpoch     idx.Epoch
	CurrentBlock     idx.Block
	CurrentBlockHash hash.Event
	CurrentBlockTime inter.Timestamp
	HighestBlock     idx.Block
	HighestEpoch     idx.Epoch
}

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
//
//go:generate mockgen -source=backend.go -destination=backend_mock.go -package=ethapi
type Backend interface {
	// General Ethereum API
	Progress() PeerProgress
	SuggestGasTipCap(ctx context.Context, certainty uint64) *big.Int
	AccountManager() *accounts.Manager
	ExtRPCEnabled() bool
	RPCGasCap() uint64            // global gas cap for eth_call over rpc: DoS protection
	RPCEVMTimeout() time.Duration // global timeout for eth_call over rpc: DoS protection
	RPCTxFeeCap() float64         // global tx fee cap for all transaction related APIs
	UnprotectedAllowed() bool     // allows only for EIP155 transactions.
	CalcBlockExtApi() bool
	HistoryPruningCutoff() uint64 // block height at which pruning was done

	// Blockchain API
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmHeader, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmHeader, error)
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error)
	StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (state.StateDB, *evmcore.EvmHeader, error)
	ResolveRpcBlockNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (idx.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmBlock, error)
	GetReceiptsByNumber(ctx context.Context, number rpc.BlockNumber) (types.Receipts, error)
	GetEVM(ctx context.Context, state vm.StateDB, header *evmcore.EvmHeader, vmConfig *vm.Config, blockContext *vm.BlockContext) (*vm.EVM, func() error, error)
	MinGasPrice() *big.Int
	MaxGasLimit() uint64

	// Transaction pool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, uint64, uint64, error)
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	TxPoolContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription

	ChainConfig(blockHeight idx.Block) *params.ChainConfig
	ChainID() *big.Int
	CurrentBlock() *evmcore.EvmBlock

	GetNetworkRules(ctx context.Context, blockHeight idx.Block) (*opera.Rules, error)

	// Lachesis DAG API
	GetEventPayload(ctx context.Context, shortEventID string) (*inter.EventPayload, error)
	GetEvent(ctx context.Context, shortEventID string) (*inter.Event, error)
	GetHeads(ctx context.Context, epoch rpc.BlockNumber) (hash.Events, error)
	CurrentEpoch(ctx context.Context) idx.Epoch
	SealedEpochTiming(ctx context.Context) (start inter.Timestamp, end inter.Timestamp)

	// Lachesis aBFT API
	GetEpochBlockState(ctx context.Context, epoch rpc.BlockNumber) (*iblockproc.BlockState, *iblockproc.EpochState, error)
	GetDowntime(ctx context.Context, vid idx.ValidatorID) (idx.Block, inter.Timestamp, error)
	GetUptime(ctx context.Context, vid idx.ValidatorID) (*big.Int, error)
	GetOriginatedFee(ctx context.Context, vid idx.ValidatorID) (*big.Int, error)

	SccApiBackend
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "dag",
			Version:   "1.0",
			Service:   NewPublicDAGChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		}, {
			Namespace: "abft",
			Version:   "1.0",
			Service:   NewPublicAbftAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "pano",
			Version:   "1.0",
			Service:   NewPublicSccApi(apiBackend),
			Public:    true,
		},
	}
}

// GetVmConfig is a utility function resolving the VM configuration for a block
// height based on the network rules.
func GetVmConfig(
	ctx context.Context,
	backend Backend,
	blockHeight idx.Block,
) (vm.Config, error) {
	rules, err := backend.GetNetworkRules(ctx, blockHeight)
	if err != nil {
		return vm.Config{}, err
	}
	if rules == nil {
		return vm.Config{}, fmt.Errorf("no network rules found for block height %d", blockHeight)
	}
	return opera.GetVmConfig(*rules), nil
}
