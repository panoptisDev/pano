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

package drivertype

import (
	"math/big"

	"github.com/panoptisDev/lachesis-base-pano/inter/idx"

	"github.com/panoptisDev/pano/inter/validatorpk"
)

var (
	// DoublesignBit is set if validator has a confirmed pair of fork events
	DoublesignBit = uint64(1 << 7)
	OkStatus      = uint64(0)
)

// Validator is the node-side representation of Driver validator
type Validator struct {
	Weight *big.Int
	PubKey validatorpk.PubKey
}

// ValidatorAndID is pair Validator + ValidatorID
type ValidatorAndID struct {
	ValidatorID idx.ValidatorID
	Validator   Validator
}
