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
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/log"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/panoptisDev/pano/inter"
	"github.com/panoptisDev/pano/inter/iblockproc"
	"github.com/panoptisDev/pano/opera"
)

const sKey = "s"

type BlockEpochState struct {
	BlockState *iblockproc.BlockState
	EpochState *iblockproc.EpochState
}

// SetHistoryBlockEpochState stores the block and epoch state in the history table.
// TODO propose to pass bs, es arguments by pointer
func (s *Store) SetHistoryBlockEpochState(epoch idx.Epoch, bs iblockproc.BlockState, es iblockproc.EpochState) {
	bs, es = bs.Copy(), es.Copy()
	bes := &BlockEpochState{
		BlockState: &bs,
		EpochState: &es,
	}
	// Write to the DB
	s.rlp.Set(s.table.BlockEpochStateHistory, epoch.Bytes(), bes)
	// Save to the LRU cache
	s.cache.BlockEpochStateHistory.Add(epoch, bes, nominalSize)
}

func (s *Store) GetHistoryBlockEpochState(epoch idx.Epoch) (*iblockproc.BlockState, *iblockproc.EpochState) {
	// Get HistoryBlockEpochState from LRU cache first.
	if v, ok := s.cache.BlockEpochStateHistory.Get(epoch); ok {
		bes := v.(*BlockEpochState)
		if bes.EpochState.Epoch == epoch {
			bs := bes.BlockState.Copy()
			es := bes.EpochState.Copy()
			return &bs, &es
		}
	}
	// read from DB
	v, ok := s.rlp.Get(s.table.BlockEpochStateHistory, epoch.Bytes(), &BlockEpochState{}).(*BlockEpochState)
	if !ok {
		return nil, nil
	}
	// Save to the LRU cache
	s.cache.BlockEpochStateHistory.Add(epoch, v, nominalSize)
	bs := v.BlockState.Copy()
	es := v.EpochState.Copy()
	return &bs, &es
}

func (s *Store) ForEachHistoryBlockEpochState(fn func(iblockproc.BlockState, iblockproc.EpochState) bool) {
	it := s.table.BlockEpochStateHistory.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		bes := BlockEpochState{}
		err := rlp.DecodeBytes(it.Value(), &bes)
		if err != nil {
			s.Log.Crit("Failed to decode BlockEpochState", "err", err)
		}
		if !fn(*bes.BlockState, *bes.EpochState) {
			break
		}
	}
}

func (s *Store) GetHistoryEpochState(epoch idx.Epoch) *iblockproc.EpochState {
	// check current BlockEpochState as a cache
	if v := s.cache.BlockEpochState.Load(); v != nil {
		bes := v.(*BlockEpochState)
		if bes.EpochState.Epoch == epoch {
			es := bes.EpochState.Copy()
			return &es
		}
	}
	_, es := s.GetHistoryBlockEpochState(epoch)
	return es
}

func (s *Store) HasHistoryBlockEpochState(epoch idx.Epoch) bool {
	has, _ := s.table.BlockEpochStateHistory.Has(epoch.Bytes())
	return has
}

func (s *Store) HasBlockEpochState() bool {
	has, _ := s.table.BlockEpochState.Has([]byte(sKey))
	return has
}

// SetBlockEpochState stores the latest block and epoch state in memory
func (s *Store) SetBlockEpochState(bs iblockproc.BlockState, es iblockproc.EpochState) {
	bs, es = bs.Copy(), es.Copy()
	s.cache.BlockEpochState.Store(&BlockEpochState{&bs, &es})
}

func (s *Store) getBlockEpochState() BlockEpochState {
	if v := s.cache.BlockEpochState.Load(); v != nil {
		return *v.(*BlockEpochState)
	}
	v, ok := s.rlp.Get(s.table.BlockEpochState, []byte(sKey), &BlockEpochState{}).(*BlockEpochState)
	if !ok {
		log.Crit("Epoch state reading failed: genesis not applied")
	}
	s.cache.BlockEpochState.Store(v)
	return *v
}

// FlushBlockEpochState stores the latest epoch and block state in DB
func (s *Store) FlushBlockEpochState() {
	s.rlp.Set(s.table.BlockEpochState, []byte(sKey), s.getBlockEpochState())
}

// GetBlockState retrieves the latest block state
func (s *Store) GetBlockState() iblockproc.BlockState {
	return *s.getBlockEpochState().BlockState
}

// GetEpochState retrieves the latest epoch state
func (s *Store) GetEpochState() iblockproc.EpochState {
	return *s.getBlockEpochState().EpochState
}

func (s *Store) GetBlockEpochState() (iblockproc.BlockState, iblockproc.EpochState) {
	bes := s.getBlockEpochState()
	return *bes.BlockState, *bes.EpochState
}

// GetEpoch retrieves the current epoch
func (s *Store) GetEpoch() idx.Epoch {
	return s.GetEpochState().Epoch
}

// GetValidators retrieves current validators
func (s *Store) GetValidators() *pos.Validators {
	return s.GetEpochState().Validators
}

// GetEpochValidators retrieves the current epoch and validators atomically
func (s *Store) GetEpochValidators() (*pos.Validators, idx.Epoch) {
	es := s.GetEpochState()
	return es.Validators, es.Epoch
}

// GetLatestBlockIndex retrieves the current block number
func (s *Store) GetLatestBlockIndex() idx.Block {
	return s.GetBlockState().LastBlock.Idx
}

// GetLatestBlock retrieves the last completed block.
func (s *Store) GetLatestBlock() *inter.Block {
	return s.GetBlock(s.GetLatestBlockIndex())
}

// GetRules retrieves current network rules
func (s *Store) GetRules() opera.Rules {
	return s.GetEpochState().Rules
}

// GetEvmChainConfig retrieves current EVM chain config
func (s *Store) GetEvmChainConfig(blockHeight idx.Block) *ethparams.ChainConfig {
	return opera.CreateTransientEvmChainConfig(
		s.GetRules().NetworkID,
		s.GetUpgradeHeights(),
		blockHeight,
	)
}

// GetEpochRules retrieves current network rules and epoch atomically
func (s *Store) GetEpochRules() (opera.Rules, idx.Epoch) {
	es := s.GetEpochState()
	return es.Rules, es.Epoch
}
