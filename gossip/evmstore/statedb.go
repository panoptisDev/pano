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

package evmstore

import (
	"fmt"
	"math/big"

	cc "github.com/panoptisDev/carmen/go/common"
	carmen "github.com/panoptisDev/carmen/go/state"
	_ "github.com/panoptisDev/carmen/go/state/gostate"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/utils/caution"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
)

// NoArchiveError is an error returned by implementation of the State interface
// for archive operations if no archive is maintained by this implementation.
const NoArchiveError = carmen.NoArchiveError

// GetLiveStateDb obtains StateDB for block processing - the live writable state
func (s *Store) GetLiveStateDb(stateRoot hash.Hash) (state.StateDB, error) {
	if s.liveStateDb == nil {
		return nil, fmt.Errorf("unable to get live StateDb - EvmStore is not open")
	}
	if s.liveStateDb.GetHash() != cc.Hash(stateRoot) {
		return nil, fmt.Errorf("unable to get Carmen live StateDB - unexpected state root (%x != %x)", s.liveStateDb.GetHash(), stateRoot)
	}
	return CreateCarmenStateDb(s.liveStateDb), nil
}

// GetTxPoolStateDB obtains StateDB for TxPool evaluation - the latest finalized, read-only.
// It is also used in emitter for emitterdriver contract reading at the start of an epoch.
func (s *Store) GetTxPoolStateDB() (state.StateDB, error) {
	// for TxPool and emitter it is ok to provide the newest state (and ignore the expected hash)
	if s.carmenState == nil {
		return nil, fmt.Errorf("unable to get TxPool StateDb - EvmStore is not open")
	}
	stateDb := carmen.CreateNonCommittableStateDBUsing(s.carmenState)
	return CreateCarmenStateDb(stateDb), nil
}

// GetArchiveBlockHeight provides the last block number available in the archive. Returns 0 if not known.
func (s *Store) GetArchiveBlockHeight() (height uint64, empty bool, err error) {
	if s.liveStateDb == nil {
		return 0, true, fmt.Errorf("unable to get archive block height - EvmStore is not open")
	}
	return s.liveStateDb.GetArchiveBlockHeight()
}

// GetRpcStateDb obtains archive StateDB for RPC requests evaluation
func (s *Store) GetRpcStateDb(blockNum *big.Int, stateRoot common.Hash) (state.StateDB, error) {
	// always use archive state (live state may mix data from various block heights)
	if s.liveStateDb == nil {
		return nil, fmt.Errorf("unable to get RPC StateDb - EvmStore is not open")
	}
	stateDb, err := s.liveStateDb.GetArchiveStateDB(blockNum.Uint64())
	if err != nil {
		return nil, err
	}
	if stateDb.GetHash() != cc.Hash(stateRoot) && blockNum.Sign() != 0 {
		return nil, fmt.Errorf("unable to get Carmen archive StateDB - unexpected state root (%x != %x)", stateDb.GetHash(), stateRoot)
	}
	return CreateCarmenStateDb(stateDb), nil
}

// CheckLiveStateHash returns if the hash of the current live StateDB hash matches (and fullsync is possible)
func (s *Store) CheckLiveStateHash(blockNum idx.Block, root hash.Hash) error {
	if s.liveStateDb == nil {
		return fmt.Errorf("unable to get live state - EvmStore is not open")
	}
	stateHash := s.liveStateDb.GetHash()
	if cc.Hash(root) != stateHash {
		return fmt.Errorf("hash of the EVM state is incorrect: blockNum: %d expected: %x reproducedHash: %x", blockNum, root, stateHash)
	}
	return nil
}

// CheckArchiveStateHash returns if the hash of the given archive StateDB hash matches
func (s *Store) CheckArchiveStateHash(blockNum idx.Block, root hash.Hash) (err error) {
	if s.carmenState == nil {
		return fmt.Errorf("unable to get live state - EvmStore is not open")
	}
	archiveState, err := s.carmenState.GetArchiveState(uint64(blockNum))
	if err != nil {
		return fmt.Errorf("unable to get archive state: %w", err)
	}
	defer caution.CloseAndReportError(&err, archiveState, "failed to close archive state")

	stateHash, err := archiveState.GetHash()
	if err != nil {
		return fmt.Errorf("unable to get archive state hash: %w", err)
	}
	if cc.Hash(root) != stateHash {
		return fmt.Errorf("hash of the archive EVM state is incorrect: blockNum: %d expected: %x actual: %x", blockNum, root, stateHash)
	}
	return nil
}
