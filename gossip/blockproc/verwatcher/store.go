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

package verwatcher

import (
	"sync/atomic"

	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/panoptisDev/pano/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	mainDB kvdb.Store

	cache struct {
		networkVersion atomic.Value
		missedVersion  atomic.Value
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store) *Store {
	s := &Store{
		mainDB:   mainDB,
		Instance: logger.New("verwatcher-store"),
	}

	return s
}
