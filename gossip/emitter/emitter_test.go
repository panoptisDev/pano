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
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/panoptisDev/pano/integration/makefakegenesis"
	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/logger"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/utils/txtime"
	"github.com/panoptisDev/pano/valkeystore"
	"github.com/panoptisDev/pano/vecmt"
)

func TestEmitter(t *testing.T) {
	cfg := DefaultConfig()
	gValidators := makefakegenesis.GetFakeValidators(3)
	vv := pos.NewBuilder()
	for _, v := range gValidators {
		vv.Set(v.ID, pos.Weight(1))
	}
	validators := vv.Build()
	cfg.Validator.ID = gValidators[0].ID

	ctrl := gomock.NewController(t)
	external := NewMockExternal(ctrl)
	txPool := NewMockTxPool(ctrl)
	signer := valkeystore.NewMockSignerAuthority(ctrl)
	txSigner := NewMockTxSigner(ctrl)

	external.EXPECT().Lock().
		AnyTimes()
	external.EXPECT().Unlock().
		AnyTimes()
	external.EXPECT().DagIndex().
		Return((*vecmt.Index)(nil)).
		AnyTimes()
	external.EXPECT().IsSynced().
		Return(true).
		AnyTimes()
	external.EXPECT().PeersNum().
		Return(int(3)).
		AnyTimes()
	external.EXPECT().StateDB().
		Return(nil).
		AnyTimes()

	em := NewEmitter(cfg, World{
		External:          external,
		TxPool:            txPool,
		EventsSigner:      signer,
		TransactionSigner: txSigner,
	}, fixedPriceBaseFeeSource{}, nil)

	t.Run("init", func(t *testing.T) {
		external.EXPECT().GetRules().
			Return(opera.FakeNetRules(opera.GetPanoUpgrades())).
			AnyTimes()

		external.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		external.EXPECT().GetGenesisTime().
			Return(inter.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("memorizeTxTimes", func(t *testing.T) {
		txtime.Enabled.Store(true)
		require := require.New(t)
		tx1 := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
		tx2 := types.NewTransaction(2, common.Address{}, big.NewInt(2), 2, big.NewInt(2), nil)

		external.EXPECT().IsBusy().
			Return(true).
			AnyTimes()

		txtime.Saw(tx1.Hash(), time.Unix(1, 0))

		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Saw(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Validated(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))

		// reversed order
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		txtime.Saw(tx2.Hash(), time.Unix(2, 0))

		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Saw(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
	})

	t.Run("tick", func(t *testing.T) {
		em.tick()
	})
}

type fixedPriceBaseFeeSource struct{}

func (fixedPriceBaseFeeSource) GetCurrentBaseFee() *big.Int {
	return big.NewInt(1e6)
}

func TestEmitter_CreateEvent_CreatesCorrectEventVersion(t *testing.T) {

	tests := map[string]opera.Upgrades{
		"pano": {
			Pano:   true,
			Allegro: false,
		},
		"allegro": {
			Pano:   true,
			Allegro: true,
		},
	}

	validator := idx.ValidatorID(1)
	builder := pos.NewBuilder()
	builder.Set(validator, pos.Weight(1))
	validators := builder.Build()

	for name, upgrades := range tests {
		t.Run(name, func(t *testing.T) {

			cases := map[bool]uint8{
				false: 2, // Single-Proposer upgrade is not enabled
				true:  3, // Single-Proposer upgrade is enabled
			}
			for singleProposer, version := range cases {
				t.Run(fmt.Sprintf("singleProposer=%t", singleProposer), func(t *testing.T) {
					ctrl := gomock.NewController(t)
					world := NewMockExternal(ctrl)
					signer := valkeystore.NewMockSignerAuthority(ctrl)

					rules := opera.Rules{
						Upgrades: upgrades,
					}
					rules.Upgrades.SingleProposerBlockFormation = singleProposer

					em := &Emitter{
						config: Config{
							Validator: ValidatorConfig{
								ID: validator,
							},
						},
						world: World{
							External:     world,
							EventsSigner: signer,
						},
					}
					em.validators.Store(validators)

					any := gomock.Any()
					world.EXPECT().GetRules().Return(rules).AnyTimes()
					world.EXPECT().GetLastEvent(any, any).AnyTimes()
					world.EXPECT().Build(any, any).AnyTimes()
					world.EXPECT().Check(any, any).Return(nil).AnyTimes()
					world.EXPECT().GetLatestBlock().Return(&inter.Block{}).AnyTimes()

					signer.EXPECT().Sign(any).AnyTimes()

					event, err := em.createEvent(nil)
					require.NoError(t, err)
					require.Equal(t, version, event.Version())
				})
			}
		})
	}
}

func TestEmitter_CreateEvent_InvalidValidatorSetIsDetected(t *testing.T) {

	ctrl := gomock.NewController(t)
	world := NewMockExternal(ctrl)
	signer := valkeystore.NewMockSignerAuthority(ctrl)
	log := logger.NewMockLogger(ctrl)

	validator := idx.ValidatorID(1)
	validators := pos.NewBuilder().Build() // invalid empty validator set

	rules := opera.Rules{
		Upgrades: opera.Upgrades{
			SingleProposerBlockFormation: true,
		},
	}

	em := &Emitter{
		Periodic: logger.Periodic{
			Instance: logger.Instance{
				Log: log,
			},
		},
		config: Config{
			Validator: ValidatorConfig{
				ID: validator,
			},
		},
		world: World{
			External:     world,
			EventsSigner: signer,
		},
	}
	em.validators.Store(validators)

	any := gomock.Any()
	world.EXPECT().GetRules().Return(rules).AnyTimes()
	world.EXPECT().GetLastEvent(any, any).AnyTimes()
	world.EXPECT().Build(any, any).AnyTimes()
	world.EXPECT().Check(any, any).Return(nil).AnyTimes()
	world.EXPECT().GetLatestBlock().Return(&inter.Block{}).AnyTimes()

	signer.EXPECT().Sign(any).AnyTimes()

	log.EXPECT().Error("Failed to create payload", "err", any)

	_, err := em.createEvent(nil)
	require.ErrorContains(t, err, "no validators")
}
