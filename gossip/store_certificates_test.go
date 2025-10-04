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
	"math"
	"math/rand"
	"testing"

	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/pano/scc/cert"
	scc_node "github.com/panoptisDev/pano/scc/node"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
)

var _ scc_node.Store = (*Store)(nil)

func TestStore_GetCommitteeCertificate_FailsIfNotPresent(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)
	_, err = store.GetCommitteeCertificate(1)
	require.ErrorContains(err, "no such certificate")
}

func TestStore_GetCommitteeCertificate_RetrievesPresentEntries(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original := cert.NewCertificate(cert.CommitteeStatement{
		Period: 1,
	})

	require.NoError(store.UpdateCommitteeCertificate(original))

	restored, err := store.GetCommitteeCertificate(1)
	require.NoError(err)
	require.Equal(original, restored)
}

func TestStore_GetCommitteeCertificate_DistinguishesBetweenPeriods(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original1 := cert.NewCertificate(cert.CommitteeStatement{
		Period:    1,
		Committee: scc.NewCommittee(),
	})

	original2 := cert.NewCertificate(cert.CommitteeStatement{
		Period:    2,
		Committee: scc.NewCommittee(scc.Member{}),
	})
	require.NotEqual(original1, original2)

	require.NoError(store.UpdateCommitteeCertificate(original1))
	require.NoError(store.UpdateCommitteeCertificate(original2))

	restored1, err := store.GetCommitteeCertificate(1)
	require.NoError(err)
	require.Equal(original1, restored1)

	restored2, err := store.GetCommitteeCertificate(2)
	require.NoError(err)
	require.Equal(original2, restored2)
}

func TestStore_GetLatestCommitteeCertificate_FailsIfNotPresent(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)
	_, err = store.GetLatestCommitteeCertificate()
	require.ErrorContains(err, "no such element")
}

func TestStore_GetLatestCommitteeCertificate_LocatesLatest(t *testing.T) {
	periods := []scc.Period{0, 1, math.MaxUint64}

	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	for _, period := range periods {
		cur := cert.NewCertificate(cert.CommitteeStatement{
			Period: period,
		})
		require.NoError(store.UpdateCommitteeCertificate(cur))

		restored, err := store.GetLatestCommitteeCertificate()
		require.NoError(err)
		require.Equal(cur, restored)
	}
}

func TestStore_EnumerateCommitteeCertificates_ReturnsAllCertificates(t *testing.T) {
	const N = 5
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	var members []scc.Member
	var originals []cert.CommitteeCertificate
	for period := range scc.Period(N) {
		cur := cert.NewCertificate(cert.CommitteeStatement{
			Period:    period,
			Committee: scc.NewCommittee(members...),
		})
		require.NoError(store.UpdateCommitteeCertificate(cur))

		originals = append(originals, cur)
		members = append(members, scc.Member{})
	}

	for first := range scc.Period(N) {
		last := first + 2
		if last > N {
			last = N
		}
		restored := []cert.CommitteeCertificate{}
		for c := range store.EnumerateCommitteeCertificates(first) {
			cur, err := c.Unwrap()
			require.NoError(err)
			restored = append(restored, cur)
			if len(restored) == 2 {
				break
			}
		}
		require.Equal(originals[first:last], restored)
	}
}

func TestStore_UpdateCommitteeCertificate_CanOverrideExisting(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original1 := cert.NewCertificate(cert.CommitteeStatement{
		Period:    1,
		Committee: scc.NewCommittee(scc.Member{}),
	})
	require.NoError(store.UpdateCommitteeCertificate(original1))

	original2 := cert.NewCertificate(cert.CommitteeStatement{
		Period:    1,
		Committee: scc.NewCommittee(scc.Member{}, scc.Member{}),
	})
	require.NotEqual(original1, original2)
	require.NoError(store.UpdateCommitteeCertificate(original2))

	restored, err := store.GetCommitteeCertificate(1)
	require.NoError(err)
	require.Equal(original2, restored)
}

func TestStore_GetBlockCertificate_FailsIfNotPresent(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)
	_, err = store.GetBlockCertificate(1)
	require.ErrorContains(err, "no such certificate")
}

func TestStore_GetBlockCertificate_RetrievesPresentEntries(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original := cert.NewCertificate(cert.BlockStatement{
		Number: 1,
	})

	require.NoError(store.UpdateBlockCertificate(original))

	restored, err := store.GetBlockCertificate(1)
	require.NoError(err)
	require.Equal(original, restored)
}

func TestStore_GetBlockCertificate_DistinguishesBetweenPeriods(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original1 := cert.NewCertificate(cert.BlockStatement{
		Number: 1,
		Hash:   [32]byte{1, 2, 3},
	})

	original2 := cert.NewCertificate(cert.BlockStatement{
		Number: 2,
		Hash:   [32]byte{4, 5, 6},
	})
	require.NotEqual(original1, original2)

	require.NoError(store.UpdateBlockCertificate(original1))
	require.NoError(store.UpdateBlockCertificate(original2))

	restored1, err := store.GetBlockCertificate(1)
	require.NoError(err)
	require.Equal(original1, restored1)

	restored2, err := store.GetBlockCertificate(2)
	require.NoError(err)
	require.Equal(original2, restored2)
}

func TestStore_GetLatestBlockCertificate_FailsIfNotPresent(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)
	_, err = store.GetLatestBlockCertificate()
	require.ErrorContains(err, "no such element")
}

func TestStore_GetLatestBlockCertificate_LocatesLatest(t *testing.T) {
	blocks := []idx.Block{0, 1, math.MaxUint64}

	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	for _, block := range blocks {
		cur := cert.NewCertificate(cert.BlockStatement{
			Number: block,
		})
		require.NoError(store.UpdateBlockCertificate(cur))

		restored, err := store.GetLatestBlockCertificate()
		require.NoError(err)
		require.Equal(cur, restored)
	}
}

func TestStore_EnumerateBlockCertificates_ReturnsAllCertificates(t *testing.T) {
	const N = 5
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	var originals []cert.BlockCertificate
	for number := range idx.Block(N) {
		cur := cert.NewCertificate(cert.BlockStatement{
			Number: number,
			Hash:   [32]byte{byte(number)},
		})
		require.NoError(store.UpdateBlockCertificate(cur))

		originals = append(originals, cur)
	}

	for first := range idx.Block(N) {
		last := first + 2
		if last > N {
			last = N
		}
		restored := []cert.BlockCertificate{}
		for c := range store.EnumerateBlockCertificates(first) {
			cur, err := c.Unwrap()
			require.NoError(err)
			restored = append(restored, cur)
			if len(restored) == 2 {
				break
			}
		}
		require.Equal(originals[first:last], restored)
	}
}

func TestStore_UpdateBlockCertificate_CanOverrideExisting(t *testing.T) {
	require := require.New(t)
	store, err := NewMemStore(t)
	require.NoError(err)

	original1 := cert.NewCertificate(cert.BlockStatement{
		Number: 1,
		Hash:   [32]byte{1, 2, 3},
	})
	require.NoError(store.UpdateBlockCertificate(original1))

	original2 := cert.NewCertificate(cert.BlockStatement{
		Number: 1,
		Hash:   [32]byte{4, 5, 6},
	})
	require.NotEqual(original1, original2)
	require.NoError(store.UpdateBlockCertificate(original2))

	restored, err := store.GetBlockCertificate(1)
	require.NoError(err)
	require.Equal(original2, restored)
}

func TestBinarySearch_CanFindTarget(t *testing.T) {
	tests := []uint64{
		0, 1, 2, 100, 256, 1024,
		math.MaxInt64 - 1, math.MaxInt64, math.MaxInt64 + 1,
		math.MaxUint64 - 1, math.MaxUint64,
	}
	for range 50 {
		tests = append(tests, rand.Uint64())
	}

	require := require.New(t)
	for _, target := range tests {
		res, err := binarySearch(func(x uint64) (bool, error) {
			return x <= target, nil
		})
		require.NoError(err)
		require.Equal(target, res)
	}
}

func TestBinarySearch_DetectsEmptyRange(t *testing.T) {
	require := require.New(t)
	_, err := binarySearch(func(uint64) (bool, error) {
		return false, nil
	})
	require.ErrorContains(err, "no such element")
}
