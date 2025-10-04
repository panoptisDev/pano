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
	"crypto/sha256"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/common"
)

// computePrevRandao computes the prevRandao from event hashes.
func computePrevRandao(events []hash.Event) common.Hash {
	bts := [24]byte{}
	for _, event := range events {
		for i := 0; i < 24; i++ {
			// first 8 bytes should be ignored as they are not pseudo-random.
			bts[i] = bts[i] ^ event[i+8]
		}
	}
	return sha256.Sum256(bts[:])
}
