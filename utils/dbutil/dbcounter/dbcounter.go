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

package dbcounter

import (
	"fmt"
	"sync/atomic"

	"github.com/panoptisDev/pano/utils/dbutil"

	"github.com/panoptisDev/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/log"
)

type DBProducer struct {
	kvdb.IterableDBProducer
	warn bool
}

type Iterator struct {
	kvdb.Iterator
	itCounter *int64
	start     []byte
	prefix    []byte
}

type Snapshot struct {
	kvdb.Snapshot
	snCounter *int64
}

type Store struct {
	kvdb.Store
	name      string
	snCounter int64
	itCounter int64
	warn      bool
}

func Wrap(db kvdb.IterableDBProducer, warn bool) kvdb.IterableDBProducer {
	return &DBProducer{db, warn}
}

func WrapStore(s kvdb.Store, name string, warn bool) *Store {
	return &Store{
		Store: s,
		name:  name,
		warn:  warn,
	}
}

func (ds *Store) Close() error {
	itCounter, snCounter := atomic.LoadInt64(&ds.itCounter), atomic.LoadInt64(&ds.snCounter)
	if itCounter != 0 || snCounter != 0 {
		err := fmt.Errorf("%s DB leak: %d iterators, %d snapshots", ds.name, itCounter, snCounter)
		if ds.warn {
			log.Warn("Possible " + err.Error())
		} else {
			return err
		}
	}
	return ds.Store.Close()
}

func (ds *Snapshot) Release() {
	atomic.AddInt64(ds.snCounter, -1)
	ds.Snapshot.Release()
}

func (ds *Store) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	atomic.AddInt64(&ds.itCounter, 1)
	return &Iterator{
		Iterator:  ds.Store.NewIterator(prefix, start),
		itCounter: &ds.itCounter,
		start:     start,
		prefix:    prefix,
	}
}

func (it *Iterator) Release() {
	atomic.AddInt64(it.itCounter, -1)
	it.Iterator.Release()
}

func (ds *Store) GetSnapshot() (kvdb.Snapshot, error) {
	atomic.AddInt64(&ds.snCounter, 1)
	snapshot, err := ds.Store.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		Snapshot:  snapshot,
		snCounter: &ds.snCounter,
	}, nil
}

func (ds *Store) IoStats() (string, error) {
	ms, ok := ds.Store.(dbutil.MeasurableStore)
	if !ok {
		return "", fmt.Errorf("IoStats not available for %t", ds.Store)
	}
	return ms.IoStats()
}

func (ds *Store) UsedDiskSpace() (string, error) {
	ms, ok := ds.Store.(dbutil.MeasurableStore)
	if !ok {
		return "", fmt.Errorf("UsedDiskSpace not available for %t", ds.Store)
	}
	return ms.UsedDiskSpace()
}

func (db *DBProducer) OpenDB(name string) (kvdb.Store, error) {
	s, err := db.IterableDBProducer.OpenDB(name)
	if err != nil {
		return nil, err
	}
	return WrapStore(s, name, db.warn), nil
}
