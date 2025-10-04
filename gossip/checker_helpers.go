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

package gossip

import (
	"sync/atomic"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"

	"github.com/panoptisDev/pano/eventcheck/gaspowercheck"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/inter/validatorpk"
	"github.com/panoptisDev/pano/opera"
)

// GasPowerCheckReader is a helper to run gas power check
type GasPowerCheckReader struct {
	Ctx atomic.Value
}

// GetValidationContext returns current validation context for gaspowercheck
func (r *GasPowerCheckReader) GetValidationContext() *gaspowercheck.ValidationContext {
	return r.Ctx.Load().(*gaspowercheck.ValidationContext)
}

// NewGasPowerContext reads current validation context for gaspowercheck
func NewGasPowerContext(s *Store, validators *pos.Validators, epoch idx.Epoch, cfg opera.EconomyRules) *gaspowercheck.ValidationContext {
	// engineMu is locked here

	short := cfg.ShortGasPower
	shortTermConfig := gaspowercheck.Config{
		Idx:                inter.ShortTermGas,
		AllocPerSec:        short.AllocPerSec,
		MaxAllocPeriod:     short.MaxAllocPeriod,
		MinEnsuredAlloc:    cfg.Gas.MaxEventGas,
		StartupAllocPeriod: short.StartupAllocPeriod,
		MinStartupGas:      short.MinStartupGas,
	}

	long := cfg.LongGasPower
	longTermConfig := gaspowercheck.Config{
		Idx:                inter.LongTermGas,
		AllocPerSec:        long.AllocPerSec,
		MaxAllocPeriod:     long.MaxAllocPeriod,
		MinEnsuredAlloc:    cfg.Gas.MaxEventGas,
		StartupAllocPeriod: long.StartupAllocPeriod,
		MinStartupGas:      long.MinStartupGas,
	}

	validatorStates := make([]gaspowercheck.ValidatorState, validators.Len())
	es := s.GetEpochState()
	for i, val := range es.ValidatorStates {
		validatorStates[i].GasRefund = val.GasRefund
		validatorStates[i].PrevEpochEvent = val.PrevEpochEvent
	}

	return &gaspowercheck.ValidationContext{
		Epoch:           epoch,
		Validators:      validators,
		EpochStart:      es.EpochStart,
		ValidatorStates: validatorStates,
		Configs: [inter.GasPowerConfigs]gaspowercheck.Config{
			inter.ShortTermGas: shortTermConfig,
			inter.LongTermGas:  longTermConfig,
		},
	}
}

// ValidatorsPubKeys stores info to authenticate validators
type ValidatorsPubKeys struct {
	Epoch   idx.Epoch
	PubKeys map[idx.ValidatorID]validatorpk.PubKey
}

// HeavyCheckReader is a helper to run heavy power checks
type HeavyCheckReader struct {
	Pubkeys atomic.Value
	Store   *Store
}

// GetEpochPubKeys is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch) {
	auth := r.Pubkeys.Load().(*ValidatorsPubKeys)

	return auth.PubKeys, auth.Epoch
}

// GetEpochPubKeysOf is safe for concurrent use
func (r *HeavyCheckReader) GetEpochPubKeysOf(epoch idx.Epoch) map[idx.ValidatorID]validatorpk.PubKey {
	auth := readEpochPubKeys(r.Store, epoch)
	if auth == nil {
		return nil
	}
	return auth.PubKeys
}

// GetEpochBlockStart is safe for concurrent use
func (r *HeavyCheckReader) GetEpochBlockStart(epoch idx.Epoch) idx.Block {
	bs, _ := r.Store.GetHistoryBlockEpochState(epoch)
	if bs == nil {
		return 0
	}
	return bs.LastBlock.Idx
}

// readEpochPubKeys reads epoch pubkeys
func readEpochPubKeys(s *Store, epoch idx.Epoch) *ValidatorsPubKeys {
	es := s.GetHistoryEpochState(epoch)
	if es == nil {
		return nil
	}
	var pubkeys = make(map[idx.ValidatorID]validatorpk.PubKey, len(es.ValidatorProfiles))
	for id, profile := range es.ValidatorProfiles {
		pubkeys[id] = profile.PubKey
	}
	return &ValidatorsPubKeys{
		Epoch:   epoch,
		PubKeys: pubkeys,
	}
}

// proposalCheckReader is an implementation of the proposalcheck.Reader
// interface providing access to event payload data and epoch validators.
type proposalCheckReader struct {
	store *Store
}

func newProposalCheckReader(store *Store) proposalCheckReader {
	return proposalCheckReader{
		store: store,
	}
}

func (r *proposalCheckReader) GetEpochValidators() *pos.Validators {
	return r.store.GetValidators()
}

func (r *proposalCheckReader) GetEventPayload(eventID hash.Event) inter.Payload {
	return *r.store.GetEventPayload(eventID).Payload()
}
