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

package threads

import (
	"github.com/panoptisDev/lachesis-base/kvdb"

	"github.com/panoptisDev/pano/logger"
)

type (
	countedFullDbProducer struct {
		kvdb.FullDBProducer
	}

	countedStore struct {
		kvdb.Store
	}

	countedIterator struct {
		kvdb.Iterator
		release func(count int)
	}
)

// CountedFullDBProducer obtains one thread from the GlobalPool for each opened iterator.
func CountedFullDBProducer(dbs kvdb.FullDBProducer) kvdb.FullDBProducer {
	return &countedFullDbProducer{dbs}
}

func (p *countedFullDbProducer) OpenDB(name string) (kvdb.Store, error) {
	s, err := p.FullDBProducer.OpenDB(name)
	return &countedStore{s}, err
}

var notifier = logger.New("threads-pool")

func (s *countedStore) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	got, release := GlobalPool.Lock(1)
	if got < 1 {
		notifier.Log.Warn("Too many DB iterators")
	}

	return &countedIterator{
		Iterator: s.Store.NewIterator(prefix, start),
		release:  release,
	}
}

func (it *countedIterator) Release() {
	it.Iterator.Release()
	it.release(1)
}
