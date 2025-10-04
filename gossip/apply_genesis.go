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
	"errors"
	"fmt"

	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/inter/ibr"
	"github.com/panoptisDev/pano/inter/ier"
	"github.com/panoptisDev/pano/opera/genesis"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/pano/utils/dbutil/autocompact"
	"github.com/panoptisDev/lachesis-base/kvdb/batched"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g genesis.Genesis) (err error) {
	// use batching wrapper for hot tables
	unwrap := s.WrapTablesAsBatched()
	defer unwrap()

	// write epochs
	var topEr *ier.LlrIdxFullEpochRecord
	g.Epochs.ForEach(func(er ier.LlrIdxFullEpochRecord) bool {
		if er.EpochState.Rules.NetworkID != g.NetworkID || er.EpochState.Rules.Name != g.NetworkName {
			err = errors.New("network ID/name mismatch")
			return false
		}
		if topEr == nil {
			topEr = &er
		}
		s.WriteFullEpochRecord(er)
		return true
	})
	if err != nil {
		return err
	}
	if topEr == nil {
		return errors.New("no ERs in genesis")
	}
	var prevEs *iblockproc.EpochState
	s.ForEachHistoryBlockEpochState(func(bs iblockproc.BlockState, es iblockproc.EpochState) bool {
		s.WriteUpgradeHeight(bs, es, prevEs)
		prevEs = &es
		return true
	})
	s.SetBlockEpochState(topEr.BlockState, topEr.EpochState)
	s.FlushBlockEpochState()

	s.SetGenesisID(g.GenesisID)
	s.SetGenesisBlockIndex(topEr.BlockState.LastBlock.Idx)

	// write blocks
	var lastBlock ibr.LlrIdxFullBlockRecord
	g.Blocks.ForEach(func(br ibr.LlrIdxFullBlockRecord) bool {
		err = s.WriteFullBlockRecord(br)
		if err != nil {
			s.Log.Crit(err.Error())
			return false
		}

		if br.Idx > lastBlock.Idx {
			lastBlock = br
		}
		return true
	})

	// write EVM items
	liveReader, err := g.FwsLiveSection.GetReader()
	if err != nil {
		s.Log.Info("Pano World State Live data not available in the genesis", "err", err)
	}

	if liveReader != nil { // has S5 section - import S5 data
		s.Log.Info("Importing Pano World State Live data from genesis")
		err = s.evm.ImportLiveWorldState(liveReader)
		if err != nil {
			return fmt.Errorf("failed to import Pano World State data from genesis; %v", err)
		}

		// import S5 archive
		archiveReader, _ := g.FwsArchiveSection.GetReader()
		if archiveReader != nil { // has archive section
			s.Log.Info("Importing Pano World State Archive data from genesis")
			err = s.evm.ImportArchiveWorldState(archiveReader)
			if err != nil {
				return fmt.Errorf("failed to import Pano World State Archive data from genesis; %v", err)
			}
		} else { // no archive section - initialize archive from the live section
			s.Log.Info("No archive in the genesis file - initializing the archive from the live state", "blockNum", lastBlock.Idx)
			liveToArchiveReader, err := g.FwsLiveSection.GetReader() // second reader of the same section for the archive import
			if err != nil {
				return fmt.Errorf("failed to get second FWS section reader; %v", err)
			}
			err = s.evm.InitializeArchiveWorldState(liveToArchiveReader, uint64(lastBlock.Idx))
			if err != nil {
				return fmt.Errorf("failed to import Pano World State data from genesis; %v", err)
			}
		}
	} else { // no S5 section in the genesis file
		// Import legacy EVM genesis section
		err = s.evm.ImportLegacyEvmData(g.RawEvmItems, uint64(lastBlock.Idx), common.Hash(lastBlock.StateRoot))
		if err != nil {
			return fmt.Errorf("import of legacy genesis data into StateDB failed; %v", err)
		}
	}

	if err := s.evm.Open(); err != nil {
		return fmt.Errorf("unable to open EvmStore to check imported state: %w", err)
	}
	if err := s.evm.CheckLiveStateHash(lastBlock.Idx, lastBlock.StateRoot); err != nil {
		return fmt.Errorf("checking imported live state failed: %w", err)
	} else {
		s.Log.Info("StateDB imported successfully, stateRoot matches", "index", lastBlock.Idx, "root", lastBlock.StateRoot)
	}

	g.CommitteeCertificates.ForEach(func(c cert.Certificate[cert.CommitteeStatement]) bool {
		if err := s.UpdateCommitteeCertificate(c); err != nil {
			s.Log.Crit("Failed to write committee certificate", "err", err)
			return false
		}
		return true
	})

	g.BlockCertificates.ForEach(func(c cert.Certificate[cert.BlockStatement]) bool {
		if err := s.UpdateBlockCertificate(c); err != nil {
			s.Log.Crit("Failed to write block certificate", "err", err)
			return false
		}
		return true
	})

	return nil
}

func (s *Store) WrapTablesAsBatched() (unwrap func()) {
	origTables := s.table

	batchedBlocks := batched.Wrap(autocompact.Wrap2M(s.table.Blocks, opt.GiB, 16*opt.GiB, false, "blocks"))
	s.table.Blocks = batchedBlocks

	batchedBlockHashes := batched.Wrap(s.table.BlockHashes)
	s.table.BlockHashes = batchedBlockHashes

	unwrapEVM := s.evm.WrapTablesAsBatched()
	return func() {
		unwrapEVM()
		_ = batchedBlocks.Flush()
		_ = batchedBlockHashes.Flush()
		s.table = origTables
	}
}
