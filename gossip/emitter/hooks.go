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

package emitter

import (
	"time"

	"github.com/panoptisDev/pano/utils/txtime"

	"github.com/panoptisDev/lachesis-base/emitter/ancestor"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/opera/contracts/emitterdriver"
	"github.com/panoptisDev/pano/utils"
	"github.com/panoptisDev/pano/utils/adapters/vecmt2dagidx"
)

// OnNewEpoch should be called after each epoch change, and on startup
func (em *Emitter) OnNewEpoch(newValidators *pos.Validators, newEpoch idx.Epoch) {
	em.maxParents = em.config.MaxParents
	rules := em.world.GetRules()
	if em.maxParents == 0 {
		em.maxParents = rules.Dag.MaxParents
	}
	if em.maxParents > rules.Dag.MaxParents {
		em.maxParents = rules.Dag.MaxParents
	}
	validators := em.validators.Load()
	if validators != nil && em.isValidator() && !validators.Exists(em.config.Validator.ID) && newValidators.Exists(em.config.Validator.ID) {
		em.syncStatus.becameValidator = time.Now()
	}

	em.validators.Store(newValidators)
	em.epoch.Store(uint32(newEpoch))

	if !em.isValidator() {
		return
	}
	lastEmit := em.loadPrevEmitTime()
	em.prevEmittedAtTime.Store(&lastEmit)

	em.originatedTxs.Clear()
	em.pendingGas = 0

	em.offlineValidators = make(map[idx.ValidatorID]bool)
	em.challenges = make(map[idx.ValidatorID]time.Time)
	em.expectedEmitIntervals = make(map[idx.ValidatorID]time.Duration)
	em.stakeRatio = make(map[idx.ValidatorID]uint64)

	// get current adjustments from emitterdriver contract
	statedb := em.world.StateDB()
	var (
		extMinInterval        time.Duration
		extConfirmingInterval time.Duration
		switchToFCIndexer     bool
	)
	if statedb != nil {
		switchToFCIndexer = statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(0)) != (common.Hash{0})
		extMinInterval = time.Duration(statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(1)).Big().Uint64())
		extConfirmingInterval = time.Duration(statedb.GetState(emitterdriver.ContractAddress, utils.U64to256(2)).Big().Uint64())
		statedb.Release()
	}
	if extMinInterval == 0 {
		extMinInterval = em.config.EmitIntervals.Min
	}
	if extConfirmingInterval == 0 {
		extConfirmingInterval = em.config.EmitIntervals.Confirming
	}

	// sanity check to ensure that durations aren't too small/large
	em.intervalsMinLock.Lock()
	em.intervals.Min = maxDuration(minDuration(em.config.EmitIntervals.Min*20, extMinInterval), em.config.EmitIntervals.Min/4)
	em.intervalsMinLock.Unlock()
	em.globalConfirmingInterval.Store(
		uint64(maxDuration(
			minDuration(em.config.EmitIntervals.Confirming*20, extConfirmingInterval),
			em.config.EmitIntervals.Confirming/4)))
	em.recountConfirmingIntervals(newValidators)

	if switchToFCIndexer {
		em.quorumIndexer = nil
		em.fcIndexer = ancestor.NewFCIndexer(newValidators, em.world.DagIndex(), em.config.Validator.ID)
	} else {
		em.quorumIndexer = ancestor.NewQuorumIndexer(newValidators, vecmt2dagidx.Wrap(em.world.DagIndex()),
			func(median, current, update idx.Event, validatorIdx idx.Validator) ancestor.Metric {
				return updMetric(median, current, update, validatorIdx, newValidators)
			})
		em.fcIndexer = nil
	}
	em.quorumIndexer = ancestor.NewQuorumIndexer(newValidators, vecmt2dagidx.Wrap(em.world.DagIndex()),
		func(median, current, update idx.Event, validatorIdx idx.Validator) ancestor.Metric {
			return updMetric(median, current, update, validatorIdx, newValidators)
		})
	em.payloadIndexer = ancestor.NewPayloadIndexer(PayloadIndexerSize)

	// forget all seen proposals of the previous epoch
	em.proposalTracker.Reset()
}

// OnEventConnected tracks new events
func (em *Emitter) OnEventConnected(e inter.EventPayloadI) {
	if !em.isValidator() {
		return
	}
	if em.fcIndexer != nil {
		em.fcIndexer.ProcessEvent(e)
	} else if em.quorumIndexer != nil {
		em.quorumIndexer.ProcessEvent(e, e.Creator() == em.config.Validator.ID)
	}
	em.payloadIndexer.ProcessEvent(e, ancestor.Metric(e.Transactions().Len()))
	for _, tx := range e.Transactions() {
		addr, _ := types.Sender(em.world.TransactionSigner, tx)
		em.originatedTxs.Inc(addr)
	}
	em.pendingGas += e.GasPowerUsed()
	if e.Creator() == em.config.Validator.ID && em.syncStatus.prevLocalEmittedID != e.ID() {
		// event was emitted by me on another instance
		em.onNewExternalEvent(e)
	}
	// if there was any challenge, erase it
	delete(em.challenges, e.Creator())
	// mark validator as online
	delete(em.offlineValidators, e.Creator())

	// track proposals to avoid proposing blocks that are already pending
	if proposal := e.Payload().Proposal; proposal != nil {
		em.proposalTracker.RegisterSeenProposal(e.Frame(), proposal.Number)
	}
}

func (em *Emitter) OnEventConfirmed(he inter.EventI) {
	if !em.isValidator() {
		return
	}
	now := time.Now()
	em.lastTimeAnEventWasConfirmed.Store(&now)
	if em.pendingGas > he.GasPowerUsed() {
		em.pendingGas -= he.GasPowerUsed()
	} else {
		em.pendingGas = 0
	}
	if he.AnyTxs() {
		e := em.world.GetEventPayload(he.ID())
		for _, tx := range e.Transactions() {
			addr, _ := types.Sender(em.world.TransactionSigner, tx)
			em.originatedTxs.Dec(addr)

			if he.Creator() == em.config.Validator.ID {
				txTime := txtime.Get(tx.Hash()) // time when was the tx seen first time
				if !txTime.Equal(time.Time{}) {
					txEndToEndTimer.Update(time.Since(txTime))
				}
			}
		}
	}

	// record event's time-to-confirm
	if he.Creator() == em.config.Validator.ID {
		eventTimeToConfirmTimer.Update(time.Since(he.CreationTime().Time()))
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
