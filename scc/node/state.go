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

package node

import "github.com/panoptisDev/pano/scc"

//go:generate mockgen -source=state.go -destination=state_mock.go -package=node

// State is an interface for the current state of an SCC node. The state's main
// responsibility is to track the composition of the current committee.
type State interface {
	// GetCurrentCommittee returns a snapshot of the current committee.
	GetCurrentCommittee() scc.Committee

	// TODO: add committee mutation support
}

// inMemoryState is an in-memory implementation of the State interface. It
// retains all state information in memory and does not persist it.
type inMemoryState struct {
	committee scc.Committee

	// TODO: update internal structure to support committee mutation
}

func newInMemoryState(committee scc.Committee) State {
	return &inMemoryState{
		committee: committee,
	}
}

func (s *inMemoryState) GetCurrentCommittee() scc.Committee {
	return s.committee
}
