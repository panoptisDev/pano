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
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
)

func TestStore_GetLatestBlock_ReportsLatestBlock(t *testing.T) {
	require := require.New(t)
	store := initStoreForTests(t)

	require.Equal(idx.Block(2), store.GetLatestBlockIndex())
	got := store.GetLatestBlock()
	want := store.GetBlock(idx.Block(2))
	require.Equal(uint64(2), got.Number)
	require.Equal(want, got)
}
