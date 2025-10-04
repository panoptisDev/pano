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

package compactdb

import (
	"path"
	"testing"

	"github.com/panoptisDev/lachesis-base/kvdb"
	"github.com/panoptisDev/lachesis-base/kvdb/leveldb"
	"github.com/panoptisDev/lachesis-base/kvdb/memorydb"
	"github.com/panoptisDev/lachesis-base/kvdb/pebble"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func TestLastKey(t *testing.T) {
	testLastKey(t, memorydb.New())
	dir := t.TempDir()
	ldb, err := leveldb.New(path.Join(dir, "ldb"), 16*opt.MiB, 64, nil, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, ldb.Close()) }()
	testLastKey(t, ldb)
	pbl, err := pebble.New(path.Join(dir, "pbl"), 16*opt.MiB, 64, nil, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, pbl.Close()) }()
	testLastKey(t, pbl)
}

func testLastKey(t *testing.T, db kvdb.Store) {
	var err error
	require.Nil(t, lastKey(db))

	err = db.Put([]byte{0}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{0}, lastKey(db))

	err = db.Put([]byte{1}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{1}, lastKey(db))

	err = db.Put([]byte{2}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{2}, lastKey(db))

	err = db.Put([]byte{1, 0}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{2}, lastKey(db))

	err = db.Put([]byte{3}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{3}, lastKey(db))

	err = db.Put([]byte{3, 0}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{3, 0}, lastKey(db))

	err = db.Put([]byte{3, 1}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{3, 1}, lastKey(db))

	err = db.Put([]byte{4}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{4}, lastKey(db))

	err = db.Put([]byte{4, 0, 0, 0}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{4, 0, 0, 0}, lastKey(db))

	err = db.Put([]byte{4, 0, 1, 0}, []byte{0})
	require.NoError(t, err)
	require.Equal(t, []byte{4, 0, 1, 0}, lastKey(db))
}

func TestTrimAfterDiff(t *testing.T) {
	a, b := trimAfterDiff([]byte{}, []byte{}, 1)
	require.Equal(t, []byte{}, a)
	require.Equal(t, []byte{}, b)

	a, b = trimAfterDiff([]byte{1, 2}, []byte{1, 3}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4}, []byte{1, 3, 4}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5}, []byte{1, 3, 4, 6}, 1)
	require.Equal(t, []byte{1, 2}, a)
	require.Equal(t, []byte{1, 3}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5}, []byte{1, 3, 4, 6}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4, 6}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4, 6, 7}, 2)
	require.Equal(t, []byte{1, 2, 4, 5}, a)
	require.Equal(t, []byte{1, 3, 4, 6}, b)

	a, b = trimAfterDiff([]byte{1, 2, 4, 5, 7}, []byte{1, 3, 4}, 2)
	require.Equal(t, []byte{1, 2, 4}, a)
	require.Equal(t, []byte{1, 3, 4}, b)
}
