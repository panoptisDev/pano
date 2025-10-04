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

package scheduler

import (
	"context"
	"math/big"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

//go:generate mockgen -source=scheduler.go -destination=scheduler_mock.go -package=scheduler

// Scheduler implements a scheduling algorithm for transactions facilitating
// the selection of transactions for inclusion in a block. The scheduler thereby
// solves the dynamic scheduling problem defined by the pending transactions in
// the transaction pool constraint by the current chain state.
type Scheduler struct {
	factory processorFactory
}

// NewScheduler creates a new scheduler to be used in the the block emitter. The
// provided chain is used to obtain the current state of the chain whenever
// transactions need to be scheduled.
func NewScheduler(chain Chain) *Scheduler {
	return newScheduler(&evmProcessorFactory{chain: chain})
}

// newScheduler is an internal factory with customizable processor
// implementations. Its is mainly intended for testing purposes.
func newScheduler(factory processorFactory) *Scheduler {
	return &Scheduler{factory: factory}
}

// Schedule determines an executable sequence of transactions from the given
// candidates collection. It does so by trail-running the transaction in order,
// accepting all transactions that can be successfully processed within the
// given gas limit.
//
// The provided collection of candidates is expected to only enumerate
// transactions with unique sender/nonce pairs. If there are duplicates, the
// scheduler will only accept the first that can be successfully processed,
// ignoring the rest.
//
// The transactions in the resulting list of transactions are ordered according
// to the priority defined by the candidates collection.
//
// Scheduling stops as soon as a schedule reaching the gas limit is found, all
// candidates have been consider, or the scheduling process is interrupted due
// to a cancellation of the provided context. In case of a cancellation, the
// transactions that have been accepted so far are returned.
func (s *Scheduler) Schedule(
	context context.Context,
	blockInfo *BlockInfo,
	candidates PrioritizedTransactions,
	limits Limits,
) []*types.Transaction {
	processor := s.factory.beginBlock(blockInfo.toEvmBlock())
	defer processor.release()

	remainingGas := limits.Gas
	remainingSize := limits.Size
	var res []*types.Transaction
	for context.Err() == nil {
		candidate := candidates.Current()
		if candidate == nil {
			break
		}

		if candidate.Gas() > remainingGas {
			candidates.Skip()
			continue
		}

		size := candidate.Size()
		if size > remainingSize {
			candidates.Skip()
			continue
		}

		success, gasUsed := processor.run(candidate)
		if !success || gasUsed > remainingGas {
			candidates.Skip()
			continue
		}
		candidates.Accept()
		res = append(res, candidate)
		remainingGas -= gasUsed
		remainingSize -= size
		if remainingGas < params.TxGas {
			break
		}
	}

	return res
}

// PrioritizedTransactions is an interface for a collection of transactions to
// be scheduled in a block. The scheduler consumes the collection's elements
// one by one, signalling whether each of them was accepted or rejected. This
// allows the collection to effectively prune candidates that are not going to
// be accepted due to failed preconditions.
type PrioritizedTransactions interface {
	// Current returns the next transaction to be scheduled. It returns nil if
	// there are no more transactions to be scheduled.
	Current() *types.Transaction

	// Accept signals that current transaction was accepted and should be
	// removed from the collection, leading to the next transaction being
	// returned by the Current method.
	Accept()

	// Skip signals that current transaction was rejected and should be removed
	// from the collection. Furthermore, any transactions that depend on the
	// rejected transaction should also be removed from the collection.
	Skip()
}

// Limits defines the limits for the block being scheduled. Schedules produced
// by the scheduler are required to respect these limits.
type Limits struct {
	// GasLimit is the maximum amount of gas that can be used by the block. This
	// is used to limit the amount of computation that can be performed in the
	// block, preventing it from consuming too processing time.
	Gas uint64
	// SizeLimit is the maximum size of the block in bytes. This is used to
	// limit the size of the block to a reasonable value, preventing it from
	// growing too large and causing issues with the network.
	Size uint64
}

// BlockInfo contains all the block meta-information accessible within EVM
// code executions. These parameters are required to produce reliable results
// of transaction executions during the scheduling. They should thus be aligned
// with the parameters used once the block is confirmed and executed on the
// chain.
type BlockInfo struct {
	// Note: ChainID would be another candidate field to be included, but it is
	// not block specific, and thus not part of the block header to be configured
	// by the scheduler for try-running transactions.

	// Number of the block being scheduled, accessible by the NUMBER opcode.
	Number idx.Block

	// Time is the block time of the block being scheduled, accessible by the
	// TIMESTAMP opcode.
	Time inter.Timestamp

	// GasLimit for the full block, as visible within the EVM. This is not
	// aligned with the actual gas limit available for being scheduled in
	// a block since overheads for epoch sealing and other transactions need
	// to be accounted for. In practice, this is a constant set by the network
	// rules orders of magnitude larger than any realistic block limit.
	// This is accessible by the GASLIMIT opcode.
	GasLimit uint64

	// Coinbase, as seen by the COINBASE opcode.
	Coinbase common.Address

	// MixHash, as seen by the PREVRANDAO opcode.
	MixHash common.Hash

	// BaseFee, as seen by the BASEFEE opcode.
	BaseFee uint256.Int

	// BlobBaseFee, as seen by the BLOBBASEFEE opcode.
	BlobBaseFee uint256.Int
}

func (b *BlockInfo) toEvmBlock() *evmcore.EvmBlock {
	return &evmcore.EvmBlock{
		EvmHeader: evmcore.EvmHeader{
			Number:      new(big.Int).SetUint64(uint64(b.Number)),
			Time:        b.Time,
			GasLimit:    b.GasLimit,
			Coinbase:    b.Coinbase,
			PrevRandao:  b.MixHash,
			BaseFee:     b.BaseFee.ToBig(),
			BlobBaseFee: b.BlobBaseFee.ToBig(),
		},
	}
}
