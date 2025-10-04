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

package pacing_of_empty_blocks

import (
	"math/big"
	"testing"
	"time"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/tests"
	"github.com/stretchr/testify/require"
)

func TestPacingOfEmptyBlocks(t *testing.T) {
	hardFork := map[string]opera.Upgrades{
		"pano":   opera.GetPanoUpgrades(),
		"allegro": opera.GetAllegroUpgrades(),
	}
	modes := map[string]bool{
		"single proposer":      true,
		"distributed proposer": false,
	}

	for name, upgrades := range hardFork {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for mode, singleProposer := range modes {
				upgrades := upgrades
				upgrades.SingleProposerBlockFormation = singleProposer
				t.Run(mode, func(t *testing.T) {
					t.Parallel()
					testPacingOfEmptyBlocks(t, upgrades)
				})
			}
		})
	}
}

func testPacingOfEmptyBlocks(
	t *testing.T,
	upgrades opera.Upgrades,
) {
	require := require.New(t)
	net := tests.StartIntegrationTestNet(t, tests.IntegrationTestNetOptions{
		Upgrades: &upgrades,
	})

	maxEmptyInterval := 4 * time.Second

	rules := tests.GetNetworkRules(t, net)
	rules.Blocks.MaxEmptyBlockSkipPeriod = inter.Timestamp(maxEmptyInterval)
	tests.UpdateNetworkRules(t, net, rules)

	rules = tests.GetNetworkRules(t, net)
	require.Equal(inter.Timestamp(maxEmptyInterval), rules.Blocks.MaxEmptyBlockSkipPeriod)

	client, err := net.GetClient()
	require.NoError(err)
	defer client.Close()

	start, err := client.BlockNumber(t.Context())
	require.NoError(err)

	// wait for a few empty blocks to be created; these empty blocks should
	// be produced every maxEmptyInterval seconds
	time.Sleep(5 * maxEmptyInterval)

	end, err := client.BlockNumber(t.Context())
	require.NoError(err)

	// there should be a few empty blocks
	require.Greater(end-start, uint64(3))

	var last time.Time
	for i := start + 1; i <= end; i++ {
		block, err := client.BlockByNumber(t.Context(), big.NewInt(int64(i)))
		require.NoError(err)

		// check if the block is empty
		require.Empty(block.Transactions(), "Block %d should be empty", i)

		// Check that the time since the last block is in [4s, 5s)
		header := block.Header()
		nanos, _, err := inter.DecodeExtraData(header.Extra)
		require.NoError(err)
		blockTime := time.Unix(int64(header.Time), int64(nanos))
		if last != (time.Time{}) {
			delay := blockTime.Sub(last)
			require.Less(maxEmptyInterval, delay)
			require.Less(delay, maxEmptyInterval+time.Second)
		}
		last = blockTime
	}
}
