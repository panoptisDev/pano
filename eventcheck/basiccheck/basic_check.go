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

package basiccheck

import (
	"errors"
	"math"

	base "github.com/panoptisDev/lachesis-base/eventcheck/basiccheck"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/panoptisDev/pano/inter"
)

var (
	ErrWrongNetForkID = errors.New("wrong network fork ID")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	// ErrTipAboveFeeCap is a sanity error to ensure no one is able to specify a
	// transaction with a tip higher than the total fee cap.
	ErrTipAboveFeeCap = errors.New("max priority fee per gas higher than max fee per gas")
)

type Checker struct {
	base base.Checker
}

// New validator which performs checks which don't require anything except event
func New() *Checker {
	return &Checker{
		base: base.Checker{},
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules
func validateTx(tx *types.Transaction) error {
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 || tx.GasPrice().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction has more gas than the basic tx fee.
	intrGas, err := intrinsicGasLegacy(tx.Data(), tx.AccessList(), tx.To() == nil)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}

	if tx.GasFeeCapIntCmp(tx.GasTipCap()) < 0 {
		return ErrTipAboveFeeCap
	}
	return nil
}

func (v *Checker) checkTxs(e inter.EventPayloadI) error {
	for _, tx := range e.Transactions() {
		if err := validateTx(tx); err != nil {
			return err
		}
	}
	return nil
}

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if e.NetForkID() != 0 {
		return ErrWrongNetForkID
	}
	if err := v.base.Validate(e); err != nil {
		return err
	}
	if e.GasPowerUsed() >= math.MaxInt64-1 || e.GasPowerLeft().Max() >= math.MaxInt64-1 {
		return base.ErrHugeValue
	}
	if e.CreationTime() <= 0 || e.MedianTime() <= 0 {
		return ErrZeroTime
	}
	if err := v.checkTxs(e); err != nil {
		return err
	}

	return nil
}
