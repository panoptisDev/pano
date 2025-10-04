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

package genesisstore

import (
	"io"

	"github.com/panoptisDev/pano/logger"
	"github.com/panoptisDev/pano/opera/genesis"
)

func BlocksSection(i int) string {
	return getSectionName("brs", i)
}

func EpochsSection(i int) string {
	return getSectionName("ers", i)
}

func EvmSection(i int) string {
	return getSectionName("evm", i)
}

func FwsLiveSection(i int) string {
	return getSectionName("fws", i)
}

func FwsArchiveSection(i int) string {
	return getSectionName("fwa", i)
}

func SccCommitteeSection(i int) string {
	return getSectionName("scc_cc", i)
}

func SccBlockSection(i int) string {
	return getSectionName("scc_bc", i)
}

type FilesMap func(string) (io.Reader, error)

// Store is a node persistent storage working over a physical zip archive.
type Store struct {
	fMap  FilesMap
	head  genesis.Header
	close func() error

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(fMap FilesMap, head genesis.Header, close func() error) *Store {
	return &Store{
		fMap:     fMap,
		head:     head,
		close:    close,
		Instance: logger.New("genesis-store"),
	}
}

// Close leaves underlying database.
func (s *Store) Close() error {
	s.fMap = nil
	return s.close()
}
