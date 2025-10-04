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

package memory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTotalMemoryIsNotZero(t *testing.T) {
	require := require.New(t)
	require.Greater(TotalMemory(), uint64(0))
	require.Less(TotalMemory(), uint64(1<<50)) // 1 PiB (sanity check)
}
