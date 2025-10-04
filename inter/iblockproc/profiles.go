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

package iblockproc

import (
	"io"
	"math/big"

	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/panoptisDev/pano/inter/drivertype"
)

type ValidatorProfiles map[idx.ValidatorID]drivertype.Validator

func (vv ValidatorProfiles) Copy() ValidatorProfiles {
	cp := make(ValidatorProfiles, len(vv))
	for k, v := range vv {
		cpv := v
		cpv.Weight = new(big.Int).Set(cpv.Weight)
		cpv.PubKey = cpv.PubKey.Copy()
		cp[k] = cpv
	}
	return cp
}

func (vv ValidatorProfiles) SortedArray() []drivertype.ValidatorAndID {
	builder := pos.NewBigBuilder()
	for id, profile := range vv {
		builder.Set(id, profile.Weight)
	}
	validators := builder.Build()
	sortedIds := validators.SortedIDs()
	arr := make([]drivertype.ValidatorAndID, validators.Len())
	for i, id := range sortedIds {
		arr[i] = drivertype.ValidatorAndID{
			ValidatorID: id,
			Validator:   vv[id],
		}
	}
	return arr
}

// EncodeRLP is for RLP serialization.
func (vv ValidatorProfiles) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.SortedArray())
}

// DecodeRLP is for RLP deserialization.
func (vv *ValidatorProfiles) DecodeRLP(s *rlp.Stream) error {
	var arr []drivertype.ValidatorAndID
	if err := s.Decode(&arr); err != nil {
		return err
	}

	*vv = make(ValidatorProfiles, len(arr))

	for _, it := range arr {
		(*vv)[it.ValidatorID] = it.Validator
	}

	return nil
}
