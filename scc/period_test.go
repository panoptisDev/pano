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

package scc

import (
	"testing"

	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
)

func TestPeriod_GetPeriod_MapsBlocksToCorrectPeriod(t *testing.T) {
	tests := []struct {
		block  idx.Block
		period Period
	}{
		{0, 0},
		{1, 0},
		{BLOCKS_PER_PERIOD - 1, 0},
		{BLOCKS_PER_PERIOD, 1},
		{BLOCKS_PER_PERIOD + 1, 1},
		{BLOCKS_PER_PERIOD*2 - 1, 1},
		{BLOCKS_PER_PERIOD * 2, 2},
		{BLOCKS_PER_PERIOD*2 + 1, 2},
	}

	for _, test := range tests {
		require.Equal(t, test.period, GetPeriod(test.block))
	}
}

func TestPeriod_IsFirstBlockInPeriod_IdentifiesFirstBlock(t *testing.T) {
	for i := idx.Block(0); i < BLOCKS_PER_PERIOD*10; i++ {
		cur := GetPeriod(i)
		next := GetPeriod(i + 1)
		if cur != next {
			require.True(t, IsFirstBlockOfPeriod(i+1))
		}
	}
}

func TestPeriod_IsLastBlockInPeriod_IdentifiesLastBlock(t *testing.T) {
	for i := idx.Block(0); i < BLOCKS_PER_PERIOD*10; i++ {
		cur := GetPeriod(i)
		next := GetPeriod(i + 1)
		if cur != next {
			require.True(t, IsLastBlockOfPeriod(i))
		}
	}
}
