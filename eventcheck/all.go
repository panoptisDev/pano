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

package eventcheck

import (
	"github.com/panoptisDev/pano/eventcheck/basiccheck"
	"github.com/panoptisDev/pano/eventcheck/epochcheck"
	"github.com/panoptisDev/pano/eventcheck/gaspowercheck"
	"github.com/panoptisDev/pano/eventcheck/heavycheck"
	"github.com/panoptisDev/pano/eventcheck/parentscheck"
	"github.com/panoptisDev/pano/eventcheck/proposalcheck"
	"github.com/panoptisDev/pano/inter"
)

// Checkers is collection of all the checkers
type Checkers struct {
	Basiccheck    *basiccheck.Checker
	Epochcheck    *epochcheck.Checker
	Parentscheck  *parentscheck.Checker
	Gaspowercheck *gaspowercheck.Checker
	Proposalcheck *proposalcheck.Checker
	Heavycheck    *heavycheck.Checker
}

// Validate runs all the checks except Poset-related
func (v *Checkers) Validate(e inter.EventPayloadI, parents inter.EventIs) error {
	if err := v.Basiccheck.Validate(e); err != nil {
		return err
	}
	if err := v.Epochcheck.Validate(e); err != nil {
		return err
	}
	if err := v.Parentscheck.Validate(e, parents); err != nil {
		return err
	}
	var selfParent inter.EventI
	if e.SelfParent() != nil {
		selfParent = parents[0]
	}
	if err := v.Gaspowercheck.Validate(e, selfParent); err != nil {
		return err
	}
	if err := v.Proposalcheck.Validate(e); err != nil {
		return err
	}
	if err := v.Heavycheck.ValidateEvent(e); err != nil {
		return err
	}
	return nil
}
