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

package inter

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetEffectiveGasLimit_IsProportionalToDelay(t *testing.T) {
	rates := []uint64{0, 1, 20, 1234, 10_000_000_000} // < gas/sec
	delay := []time.Duration{
		0, 1 * time.Nanosecond, 50 * time.Microsecond,
		100 * time.Millisecond, 1500 * time.Millisecond,
	}

	for _, rate := range rates {
		for _, d := range delay {
			got := GetEffectiveGasLimit(d, rate, math.MaxUint64)
			want := rate * uint64(d) / uint64(time.Second)
			require.Equal(t, want, got, "rate %d, delay %v", rate, d)
		}
	}
}

func TestGetEffectiveGasLimit_IsZeroForNegativeDelay(t *testing.T) {
	blockLimit := uint64(math.MaxUint64)
	require.Equal(t, uint64(0), GetEffectiveGasLimit(-1*time.Nanosecond, 100, blockLimit))
	require.Equal(t, uint64(0), GetEffectiveGasLimit(-1*time.Second, 100, blockLimit))
	require.Equal(t, uint64(0), GetEffectiveGasLimit(-1*time.Hour, 100, blockLimit))
}

func TestGetEffectiveGasLimit_IsCappedAtMaximumAccumulationTime(t *testing.T) {
	rate := uint64(100)
	maxAccumulationTime := maxAccumulationTime
	for _, d := range []time.Duration{
		maxAccumulationTime,
		maxAccumulationTime + 1*time.Nanosecond,
		maxAccumulationTime + 1*time.Second,
		maxAccumulationTime + 1*time.Hour,
	} {
		got := GetEffectiveGasLimit(d, rate, math.MaxUint64)
		want := GetEffectiveGasLimit(maxAccumulationTime, rate, math.MaxUint64)
		require.Equal(t, want, got, "delay %v", d)
	}
}

func TestGetEffectiveGasLimit_IsCappedByBlockGasLimit(t *testing.T) {
	delta := 100 * time.Millisecond
	rate := uint64(100_000)
	allocation := rate * uint64(delta) / uint64(time.Second)

	limits := []uint64{
		0,
		1,
		allocation - 1,
		allocation,
		allocation + 1,
		math.MaxUint64,
	}

	for _, blockLimit := range limits {
		got := GetEffectiveGasLimit(delta, rate, blockLimit)
		want := min(allocation, blockLimit)
		require.Equal(t, want, got, "block limit %d", blockLimit)
	}
}
