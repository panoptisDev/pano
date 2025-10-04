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

package epochcheck

import (
	"errors"

	base "github.com/panoptisDev/lachesis-base/eventcheck/epochcheck"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/opera"
)

//go:generate mockgen -source=epoch_check.go -destination=epoch_check_mock.go -package=epochcheck

var (
	ErrTooManyParents    = errors.New("event has too many parents")
	ErrTooBigGasUsed     = errors.New("event uses too much gas power")
	ErrWrongGasUsed      = errors.New("event has incorrect gas power")
	ErrUnderpriced       = errors.New("event transaction underpriced")
	ErrTooBigExtra       = errors.New("event extra data is too large")
	ErrWrongVersion      = errors.New("event has wrong version")
	ErrUnsupportedTxType = errors.New("unsupported tx type")
	ErrNotRelevant       = base.ErrNotRelevant
	ErrAuth              = base.ErrAuth
)

// Reader returns currents epoch and its validators group.
type Reader interface {
	base.Reader
	GetEpochRules() (opera.Rules, idx.Epoch)
}

// Checker which require only current epoch info
type Checker struct {
	Base   *base.Checker
	reader Reader
}

func New(reader Reader) *Checker {
	return &Checker{
		Base:   base.New(reader),
		reader: reader,
	}
}

func CalcGasPowerUsed(e inter.EventPayloadI, rules opera.Rules) uint64 {
	txsGas := uint64(0)
	// In the single-proposer protocol, the gas usage of individual transactions
	// is not attributed to the individual proposer, since each proposer needs
	// to be able to create proposals with the full gas limit. Thus, only the
	// transactions being part of the distributed proposal protocol are counted.
	for _, tx := range e.TransactionsToMeter() {
		txsGas += tx.Gas()
	}

	gasCfg := rules.Economy.Gas

	parentsGas := uint64(0)
	if idx.Event(len(e.Parents())) > rules.Dag.MaxFreeParents {
		parentsGas = uint64(idx.Event(len(e.Parents()))-rules.Dag.MaxFreeParents) * gasCfg.ParentGas
	}
	extraGas := uint64(len(e.Extra())) * gasCfg.ExtraDataGas

	mpsGas := uint64(len(e.MisbehaviourProofs())) * gasCfg.MisbehaviourProofGas

	bvsGas := uint64(0)
	if e.BlockVotes().Start != 0 {
		bvsGas = gasCfg.BlockVotesBaseGas + uint64(len(e.BlockVotes().Votes))*gasCfg.BlockVoteGas
	}

	ersGas := uint64(0)
	if e.EpochVote().Epoch != 0 {
		ersGas = gasCfg.EpochVoteGas
	}

	return txsGas + parentsGas + extraGas + gasCfg.EventGas + mpsGas + bvsGas + ersGas
}

func (v *Checker) checkGas(e inter.EventPayloadI, rules opera.Rules) error {
	if e.GasPowerUsed() > rules.Economy.Gas.MaxEventGas {
		return ErrTooBigGasUsed
	}
	if e.GasPowerUsed() != CalcGasPowerUsed(e, rules) {
		return ErrWrongGasUsed
	}
	return nil
}

func CheckTxs(txs types.Transactions, rules opera.Rules) error {
	maxType := uint8(types.LegacyTxType)
	if rules.Upgrades.Berlin {
		maxType = types.AccessListTxType
	}
	if rules.Upgrades.London {
		maxType = types.DynamicFeeTxType
	}
	if rules.Upgrades.Pano {
		maxType = types.BlobTxType
	}
	if rules.Upgrades.Allegro {
		maxType = types.SetCodeTxType
	}
	for _, tx := range txs {
		if tx.Type() > maxType {
			return ErrUnsupportedTxType
		}
	}
	return nil
}

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if err := v.Base.Validate(e); err != nil {
		return err
	}
	rules, epoch := v.reader.GetEpochRules()
	// Check epoch of the rules to prevent a race condition
	if e.Epoch() != epoch {
		return base.ErrNotRelevant
	}
	if idx.Event(len(e.Parents())) > rules.Dag.MaxParents {
		return ErrTooManyParents
	}
	if uint32(len(e.Extra())) > rules.Dag.MaxExtraData {
		return ErrTooBigExtra
	}
	if err := v.checkGas(e, rules); err != nil {
		return err
	}
	if err := CheckTxs(e.Transactions(), rules); err != nil {
		return err
	}

	version := uint8(0)
	if rules.Upgrades.SingleProposerBlockFormation {
		version = 3
	} else if rules.Upgrades.Pano {
		version = 2
	} else if rules.Upgrades.Llr {
		version = 1
	}
	if e.Version() != version {
		return ErrWrongVersion
	}
	return nil
}
