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

package blockproc

import (
	"math/big"

	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
)

//go:generate mockgen -source=interface.go -package=blockproc -destination=interface_mock.go

type TxListener interface {
	OnNewLog(*types.Log)
	OnNewReceipt(tx *types.Transaction, r *types.Receipt, originator idx.ValidatorID, baseFee *big.Int, blobBaseFee *big.Int)
	Finalize() iblockproc.BlockState
	Update(bs iblockproc.BlockState, es iblockproc.EpochState)
}

type TxListenerModule interface {
	Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, statedb state.StateDB) TxListener
}

type TxTransactor interface {
	PopInternalTxs(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState, sealing bool, statedb state.StateDB) types.Transactions
}

type SealerProcessor interface {
	EpochSealing() bool
	SealEpoch() (iblockproc.BlockState, iblockproc.EpochState)
	Update(bs iblockproc.BlockState, es iblockproc.EpochState)
}

type SealerModule interface {
	Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState) SealerProcessor
}

type ConfirmedEventsProcessor interface {
	ProcessConfirmedEvent(inter.EventI)
	Finalize(block iblockproc.BlockCtx, blockSkipped bool) iblockproc.BlockState
}

type ConfirmedEventsModule interface {
	Start(bs iblockproc.BlockState, es iblockproc.EpochState) ConfirmedEventsProcessor
}

type EVMProcessor interface {
	Execute(txs types.Transactions, gasLimit uint64) []evmcore.ProcessedTransaction
	Finalize() (evmBlock *evmcore.EvmBlock, numSkipped int, receipts types.Receipts)
}

type EVM interface {
	Start(
		block iblockproc.BlockCtx,
		statedb state.StateDB,
		reader evmcore.DummyChain,
		onNewLog func(*types.Log),
		net opera.Rules,
		evmCfg *params.ChainConfig,
		prevrandao common.Hash,
	) EVMProcessor
}
