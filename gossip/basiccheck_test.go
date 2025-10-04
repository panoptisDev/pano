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
	"bytes"
	"math"
	"math/big"
	"testing"

	lbasiccheck "github.com/panoptisDev/lachesis-base/eventcheck/basiccheck"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/panoptisDev/pano/eventcheck/basiccheck"
	"github.com/panoptisDev/pano/inter"
)

func setup(t *testing.T) (*testEnv, *inter.MutableEventPayload) {
	t.Helper()

	const (
		validatorsNum = 10
		startEpoch    = 1
	)

	env := newTestEnv(startEpoch, validatorsNum, t)

	em := env.emitters[0]
	e, err := em.EmitEvent()
	require.NoError(t, err)
	require.NotNil(t, e)

	me := mutableEventPayloadFromImmutable(e)
	return env, me
}

func TestBasicCheckValidate(t *testing.T) {

	testCases := map[string]struct {
		prepareTest func(*inter.MutableEventPayload)
		expectedErr error
	}{
		"ErrWrongNetForkID": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetNetForkID(1)
			},
			expectedErr: basiccheck.ErrWrongNetForkID,
		},
		"Validate checkLimits ErrHugeValue": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetEpoch(math.MaxInt32 - 1)
			},
			expectedErr: lbasiccheck.ErrHugeValue,
		},
		"Validate checkInited checkInited ErrNotInited": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(0)
			},
			expectedErr: lbasiccheck.ErrNotInited,
		},
		"Validate checkInited ErrNoParents": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				payload.SetSeq(idx.Event(2))
				parents := hash.Events{}
				payload.SetParents(parents)
			},
			expectedErr: lbasiccheck.ErrNoParents,
		},
		"Validate ErrHugeValue-1": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				payload.SetGasPowerUsed(math.MaxInt64 - 1)
			},
			expectedErr: lbasiccheck.ErrHugeValue,
		},
		"Validate ErrHugeValue-2": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				payload.SetGasPowerLeft(inter.GasPowerLeft{Gas: [2]uint64{math.MaxInt64 - 1, math.MaxInt64}})
			},
			expectedErr: lbasiccheck.ErrHugeValue,
		},
		"Validate ErrZeroTime-1": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				payload.SetCreationTime(0)
			},
			expectedErr: basiccheck.ErrZeroTime,
		},
		"Validate ErrZeroTime-2": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				payload.SetMedianTime(0)
			},
			expectedErr: basiccheck.ErrZeroTime,
		},
		"Validate checkTxs validateTx ErrNegativeValue-1": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: h.Big(),
					Gas:      math.MaxUint64,
					To:       nil,
					Value:    big.NewInt(-1000),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				payload.SetTxs(txs)
			},
			expectedErr: basiccheck.ErrNegativeValue,
		},
		"Validate checkTxs validateTx ErrNegativeValue-2": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: big.NewInt(-1000),
					Gas:      math.MaxUint64,
					To:       nil,
					Value:    h.Big(),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				payload.SetTxs(txs)
			},
			expectedErr: basiccheck.ErrNegativeValue,
		},
		"Validate checkTxs validateTx ErrIntrinsicGas": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: h.Big(),
					Gas:      0,
					To:       nil,
					Value:    h.Big(),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				payload.SetTxs(txs)
			},
			expectedErr: basiccheck.ErrIntrinsicGas,
		},
		"Validate checkTxs validateTx ErrTipAboveFeeCap": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.DynamicFeeTx{
					Nonce:     math.MaxUint64,
					To:        nil,
					Data:      []byte{},
					Gas:       math.MaxUint64,
					Value:     h.Big(),
					ChainID:   new(big.Int),
					GasTipCap: big.NewInt(1000),
					GasFeeCap: new(big.Int),
					V:         big.NewInt(0xff),
					R:         h.Big(),
					S:         h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				payload.SetTxs(txs)
			},
			expectedErr: basiccheck.ErrTipAboveFeeCap,
		},
		"Validate returns nil": {
			prepareTest: func(payload *inter.MutableEventPayload) {
				payload.SetSeq(idx.Event(1))
				payload.SetEpoch(idx.Epoch(1))
				payload.SetFrame(idx.Frame(1))
				payload.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.DynamicFeeTx{
					Nonce:     math.MaxUint64,
					To:        nil,
					Data:      []byte{},
					Gas:       math.MaxUint64,
					Value:     h.Big(),
					ChainID:   new(big.Int),
					GasTipCap: new(big.Int),
					GasFeeCap: big.NewInt(1000),
					V:         big.NewInt(0xff),
					R:         h.Big(),
					S:         h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				payload.SetTxs(txs)
			},
			expectedErr: nil,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			env, payload := setup(t)
			t.Cleanup(func() {
				err := env.Close()
				require.NoError(t, err)
			})

			test.prepareTest(payload)

			err := env.checkers.Basiccheck.Validate(payload)
			if test.expectedErr != nil {
				require.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func mutableEventPayloadFromImmutable(e *inter.EventPayload) *inter.MutableEventPayload {
	me := &inter.MutableEventPayload{}
	me.SetVersion(e.Version())
	me.SetNetForkID(e.NetForkID())
	me.SetCreator(e.Creator())
	me.SetEpoch(e.Epoch())
	me.SetCreationTime(e.CreationTime())
	me.SetMedianTime(e.MedianTime())
	me.SetPrevEpochHash(e.PrevEpochHash())
	me.SetExtra(e.Extra())
	me.SetGasPowerLeft(e.GasPowerLeft())
	me.SetGasPowerUsed(e.GasPowerUsed())
	me.SetPayload(*e.Payload())
	me.SetPayloadHash(e.PayloadHash())
	me.SetSig(e.Sig())
	me.SetTxs(e.Transactions())
	return me
}
