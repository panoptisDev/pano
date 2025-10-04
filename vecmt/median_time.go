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

package vecmt

import (
	"fmt"
	"sort"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"

	"github.com/panoptisDev/pano/inter"
)

// medianTimeIndex is a handy index for the MedianTime() func
type medianTimeIndex struct {
	weight       pos.Weight
	creationTime inter.Timestamp
}

// MedianTime calculates weighted median of claimed time within highest observed events.
func (vi *Index) MedianTime(id hash.Event, defaultTime inter.Timestamp) inter.Timestamp {
	vi.InitBranchesInfo()
	// Get event by hash
	_before := vi.Engine.GetMergedHighestBefore(id)
	if _before == nil {
		vi.crit(fmt.Errorf("event=%s not found", id.String()))
	}
	before := _before.(*HighestBefore)

	honestTotalWeight := pos.Weight(0) // isn't equal to validators.TotalWeight(), because doesn't count cheaters
	highests := make([]medianTimeIndex, 0, len(vi.validatorIdxs))
	// convert []HighestBefore -> []medianTimeIndex
	for creatorIdxI := range vi.validators.IDs() {
		creatorIdx := idx.Validator(creatorIdxI)
		highest := medianTimeIndex{}
		highest.weight = vi.validators.GetWeightByIdx(creatorIdx)
		highest.creationTime = before.VTime.Get(creatorIdx)
		seq := before.VSeq.Get(creatorIdx)

		// edge cases
		if seq.IsForkDetected() {
			// cheaters don't influence medianTime
			highest.weight = 0
		} else if seq.Seq == 0 {
			// if no event was observed from this node, then use genesisTime
			highest.creationTime = defaultTime
		}

		highests = append(highests, highest)
		honestTotalWeight += highest.weight
	}
	// it's technically possible honestTotalWeight == 0 (all validators are cheaters)

	// sort by claimed time (partial order is enough here, because we need only creationTime)
	sort.Slice(highests, func(i, j int) bool {
		a, b := highests[i], highests[j]
		return a.creationTime < b.creationTime
	})

	// Calculate weighted median
	halfWeight := honestTotalWeight / 2
	var currWeight pos.Weight
	var median inter.Timestamp
	for _, highest := range highests {
		currWeight += highest.weight
		if currWeight >= halfWeight {
			median = highest.creationTime
			break
		}
	}

	// sanity check
	if currWeight < halfWeight || currWeight > honestTotalWeight {
		vi.crit(fmt.Errorf("median wasn't calculated correctly, median=%d, currWeight=%d, totalWeight=%d, len(highests)=%d, id=%s",
			median,
			currWeight,
			honestTotalWeight,
			len(highests),
			id.String(),
		))
	}

	return median
}
