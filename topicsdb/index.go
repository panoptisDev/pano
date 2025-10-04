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

package topicsdb

import (
	"context"

	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/kvdb"
	"github.com/panoptisDev/lachesis-base/kvdb/batched"
	"github.com/panoptisDev/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// index is a specialized indexes for log records storing and fetching.
type index struct {
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count (where topicN=0 is for address)
		Topic kvdb.Store `table:"t"`
		// (blockN+TxHash+logIndex) -> ordered topic_count topics, blockHash, address, data
		Logrec kvdb.Store `table:"r"`
	}
}

func newIndex(db kvdb.Store) *index {
	tt := &index{}
	table.MigrateTables(&tt.table, db)
	return tt
}

func (tt *index) WrapTablesAsBatched() (unwrap func()) {
	origTables := tt.table
	batchedTopic := batched.Wrap(tt.table.Topic)
	tt.table.Topic = batchedTopic
	batchedLogrec := batched.Wrap(tt.table.Logrec)
	tt.table.Logrec = batchedLogrec
	return func() {
		_ = batchedTopic.Flush()
		_ = batchedLogrec.Flush()
		tt.table = origTables
	}
}

// FindInBlocks returns all log records of block range by pattern. 1st pattern element is an address.
func (tt *index) FindInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash) (logs []*types.Log, err error) {
	err = tt.ForEachInBlocks(
		ctx,
		from, to,
		pattern,
		func(l *types.Log) bool {
			logs = append(logs, l)
			return true
		})

	return
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *index) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if 0 < to && to < from {
		return nil
	}

	pattern, err := limitPattern(pattern)
	if err != nil {
		return err
	}

	onMatched := func(rec *logrec) (gonext bool, err error) {
		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	return tt.searchParallel(ctx, pattern, uint64(from), uint64(to), onMatched, doNothing)
}

func doNothing() {}

// Push log record to database batch
func (tt *index) Push(recs ...*types.Log) error {
	for _, rec := range recs {
		if len(rec.Topics) > maxTopicsCount {
			return ErrTooBigTopics
		}

		id := NewID(rec.BlockNumber, rec.TxHash, rec.Index)

		// write data
		buf := make([]byte, 0, common.HashLength*len(rec.Topics)+common.HashLength+common.AddressLength+len(rec.Data))
		for _, topic := range rec.Topics {
			buf = append(buf, topic.Bytes()...)
		}
		buf = append(buf, rec.BlockHash.Bytes()...)
		buf = append(buf, rec.Address.Bytes()...)
		buf = append(buf, rec.Data...)
		if err := tt.table.Logrec.Put(id.Bytes(), buf); err != nil {
			return err
		}

		// write index
		var (
			count = posToBytes(uint8(len(rec.Topics)))
			pos   uint8
		)
		pushIndex := func(topic common.Hash) error {
			key := topicKey(topic, pos, id)
			if err := tt.table.Topic.Put(key, count); err != nil {
				return err
			}
			pos++
			return nil
		}

		if err := pushIndex(common.BytesToHash(rec.Address[:])); err != nil {
			return err
		}

		for _, topic := range rec.Topics {
			if err := pushIndex(topic); err != nil {
				return err
			}
		}

	}

	return nil
}

func (tt *index) Close() {
	_ = tt.table.Topic.Close()
	_ = tt.table.Logrec.Close()
}
