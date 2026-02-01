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
	"github.com/panoptisDev/lachesis-base-pano/inter/dag"
	"github.com/panoptisDev/lachesis-base-pano/inter/idx"
	"github.com/panoptisDev/lachesis-base-pano/vecengine"
	"github.com/panoptisDev/lachesis-base-pano/vecfc"

	"github.com/panoptisDev/pano/inter"
)

type CreationTimer interface {
	CreationTime() inter.Timestamp
}

func (hb *HighestBefore) InitWithEvent(i idx.Validator, e dag.Event) {
	hb.VSeq.InitWithEvent(i, e)
	hb.VTime.Set(i, e.(CreationTimer).CreationTime())
}

func (hb *HighestBefore) IsEmpty(i idx.Validator) bool {
	return hb.VSeq.IsEmpty(i)
}

func (hb *HighestBefore) IsForkDetected(i idx.Validator) bool {
	return hb.VSeq.IsForkDetected(i)
}

func (hb *HighestBefore) Seq(i idx.Validator) idx.Event {
	return hb.VSeq.Seq(i)
}

func (hb *HighestBefore) MinSeq(i idx.Validator) idx.Event {
	return hb.VSeq.MinSeq(i)
}

func (hb *HighestBefore) SetForkDetected(i idx.Validator) {
	hb.VSeq.SetForkDetected(i)
}

func (hb *HighestBefore) CollectFrom(_other vecengine.HighestBeforeI, num idx.Validator) {
	other := _other.(*HighestBefore)
	for branchID := idx.Validator(0); branchID < num; branchID++ {
		hisSeq := other.VSeq.Get(branchID)
		if hisSeq.Seq == 0 && !hisSeq.IsForkDetected() {
			// hisSeq doesn't observe anything about this branchID
			continue
		}
		mySeq := hb.VSeq.Get(branchID)

		if mySeq.IsForkDetected() {
			// mySeq observes the maximum already
			continue
		}
		if hisSeq.IsForkDetected() {
			// set fork detected
			hb.SetForkDetected(branchID)
		} else {
			if mySeq.Seq == 0 || mySeq.MinSeq > hisSeq.MinSeq {
				// take hisSeq.MinSeq
				mySeq.MinSeq = hisSeq.MinSeq
				hb.VSeq.Set(branchID, mySeq)
			}
			if mySeq.Seq < hisSeq.Seq {
				// take hisSeq.Seq
				mySeq.Seq = hisSeq.Seq
				hb.VSeq.Set(branchID, mySeq)
				hb.VTime.Set(branchID, other.VTime.Get(branchID))
			}
		}
	}
}

func (hb *HighestBefore) GatherFrom(to idx.Validator, _other vecengine.HighestBeforeI, from []idx.Validator) {
	other := _other.(*HighestBefore)
	// read all branches to find highest event
	highestBranchSeq := vecfc.BranchSeq{}
	highestBranchTime := inter.Timestamp(0)
	for _, branchID := range from {
		vseq := other.VSeq.Get(branchID)
		vtime := other.VTime.Get(branchID)
		if vseq.IsForkDetected() {
			highestBranchSeq = vseq
			break
		}
		if vseq.Seq > highestBranchSeq.Seq {
			highestBranchSeq = vseq
			highestBranchTime = vtime
		}
	}
	hb.VSeq.Set(to, highestBranchSeq)
	hb.VTime.Set(to, highestBranchTime)
}
