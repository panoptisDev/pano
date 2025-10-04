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
	"fmt"
	"testing"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
)

func TestEthApiBackend_GetNetworkRules_LoadsRulesFromEpoch(t *testing.T) {
	require := require.New(t)

	blockNumber := idx.Block(12)
	epoch := idx.Epoch(3)

	store, err := NewMemStore(t)
	require.NoError(err)

	store.SetBlock(
		blockNumber,
		inter.NewBlockBuilder().
			WithNumber(uint64(blockNumber)).
			WithEpoch(epoch).
			Build(),
	)
	require.True(store.HasBlock(blockNumber))

	rules := opera.FakeNetRules(opera.Upgrades{})
	rules.Name = "test-rules"

	store.SetHistoryBlockEpochState(
		epoch,
		iblockproc.BlockState{},
		iblockproc.EpochState{
			Epoch: epoch,
			Rules: rules,
		},
	)

	backend := &EthAPIBackend{
		svc: &Service{
			store: store,
		},
		state: &EvmStateReader{
			store: store,
		},
	}

	got, err := backend.GetNetworkRules(t.Context(), blockNumber)
	require.NoError(err)

	// Rules contain functions that cannot be compared directly,
	// so we compare their string representations.
	want := fmt.Sprintf("%+v", rules)
	have := fmt.Sprintf("%+v", got)
	require.Equal(want, have, "Network rules do not match")
}

func TestEthApiBackend_GetNetworkRules_MissingBlockReturnsNilRules(t *testing.T) {
	require := require.New(t)

	blockNumber := idx.Block(12)

	store, err := NewMemStore(t)
	require.NoError(err)
	require.False(store.HasBlock(blockNumber))

	backend := &EthAPIBackend{
		state: &EvmStateReader{
			store: store,
		},
	}

	rules, err := backend.GetNetworkRules(t.Context(), blockNumber)
	require.NoError(err)
	require.Nil(rules)
}
