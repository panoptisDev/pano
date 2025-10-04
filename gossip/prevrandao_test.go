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
	"math/rand"
	"strings"
	"testing"

	"github.com/panoptisDev/lachesis-base/hash"
)

func TestComputePrevRandao_ComputationIsDeterministic(t *testing.T) {
	events := hash.FakeEvents(5)
	randao1 := computePrevRandao(events)
	rand.Shuffle(len(events), func(i, j int) {
		events[i], events[j] = events[j], events[i]
	})
	randao2 := computePrevRandao(events)
	if randao1 != randao2 {
		t.Error("computation is not deterministic")
	}
}

func TestComputePrevRandao_ComputationProducesCorrectValue(t *testing.T) {
	tests := []struct {
		name   string
		events hash.Events
		want   string
	}{
		{
			name:   "empty_events",
			events: hash.Events{},
			want:   "0x9d908ecfb6b256def8b49a7c504e6c889c4b0e41fe6ce3e01863dd7b61a20aa0",
		},
		{
			name: "one_event",
			events: hash.Events{
				hash.HexToEventHash("0x1234"),
			},
			want: "0x445c47179cf0e0e25fc47fcd611f2fff71742cfa2da9f42ff1a2aba577562bde",
		},
		{
			name: "multiple_events",
			events: hash.Events{
				hash.HexToEventHash("0x5678"),
				hash.HexToEventHash("0x9012"),
				hash.HexToEventHash("0x3456"),
			},
			want: "0xd260b051cbc12b222995f09e75d1596850a94bb257015bee25b84c7e8015de06",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := computePrevRandao(test.events)
			if !strings.EqualFold(got.String(), test.want) {
				t.Errorf("unexpected hash; got: %s, want: %s", got, test.want)
			}
		})
	}
}
