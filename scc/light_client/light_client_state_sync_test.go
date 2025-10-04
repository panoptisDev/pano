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

package light_client

import (
	"crypto/sha256"
	"slices"
	"testing"

	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/pano/scc/bls"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLightClientState_CanSyncWithProvider(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	// generate history of blocks and committees certificates
	blockHeight := scc.BLOCKS_PER_PERIOD * 50 / 3
	firstCommittee, provider, err := generateCertificatesAndProvider(
		ctrl, idx.Block(blockHeight))
	require.NoError(err)

	// create a new state with the first committee
	state := newState(firstCommittee)
	headNumber, err := state.sync(provider)
	require.NoError(err)
	require.Equal(idx.Block(blockHeight), headNumber)
}

// /////////////////////////
// Helper functions
// /////////////////////////

// generateCertificatesAndProvider generates a committee and a provider for testing.
// The amount of committee certificates is generated based on the given
// block height.
// The provider is a mock provider that returns the generated committee and
// blocks certificates.
func generateCertificatesAndProvider(
	ctrl *gomock.Controller,
	blockHeight idx.Block,
) (scc.Committee, provider, error) {

	// generate first committee with committees and blocks certificates
	firstCommittee, blocks, committees, err := generateHistory(blockHeight)
	if err != nil {
		return scc.Committee{}, nil, err
	}

	// prepare mock provider
	prov := prepareProvider(ctrl, blockHeight, blocks, committees)

	return firstCommittee, prov, nil
}

// generateHistory generates a history of blocks and committees certificates.
// Certificates are signed by 3 committee members and the committee rotates
// every period.
func generateHistory(blockHeight idx.Block) (
	genesis scc.Committee,
	blocks []cert.BlockCertificate,
	committees []cert.CommitteeCertificate,
	err error,
) {

	keys := []bls.PrivateKey{
		bls.NewPrivateKey(),
		bls.NewPrivateKey(),
		bls.NewPrivateKey(),
	}

	genesis = scc.NewCommittee(
		makeMember(keys[0]),
		makeMember(keys[1]),
		makeMember(keys[2]))

	// generate first block and committee certificates.
	blocks = append(blocks, cert.NewCertificate(cert.BlockStatement{}))
	committees = append(committees, cert.NewCertificate(cert.CommitteeStatement{
		Committee: genesis,
	}))

	// generate certificates up to blockHeight.
	committee := genesis
	head := idx.Block(0)
	headHash := common.Hash{}
	for i := head; i < blockHeight; i++ {

		// Compute next block.
		head += 1
		// the next line is a dummy hash, only for testing purposes.
		headHash = common.Hash(sha256.Sum256(headHash[:]))

		// Add period boundaries, update the committee.
		if scc.IsFirstBlockOfPeriod(head) {
			committee := scc.NewCommittee(rotate(committee.Members())...)

			certificate := cert.NewCertificate(
				cert.NewCommitteeStatement(
					1234,
					scc.GetPeriod(head),
					committee))

			for i, key := range keys {
				err := certificate.Add(
					scc.MemberId(i),
					cert.Sign(certificate.Subject(), key))
				if err != nil {
					return scc.Committee{}, nil, nil, err
				}
			}
			committees = append(committees, certificate)
			keys = rotate(keys)
		}

		// Sign the new block using the current committee.
		block := cert.NewCertificate(
			cert.NewBlockStatement(
				1234,
				head,
				headHash,
				headHash,
			))

		for i, key := range keys {
			err := block.Add(scc.MemberId(i), cert.Sign(block.Subject(), key))
			if err != nil {
				return scc.Committee{}, nil, nil, err
			}
		}
		blocks = append(blocks, block)

	}

	return genesis, blocks, committees, nil
}

// prepareProvider prepares a mock provider that returns the given blocks and
// committees certificates.
// if the block number is LatestBlock, it returns the latest block.
func prepareProvider(
	ctrl *gomock.Controller,
	blockHeight idx.Block,
	blocks []cert.BlockCertificate,
	committees []cert.CommitteeCertificate,
) provider {

	prov := NewMockprovider(ctrl)
	prov.
		EXPECT().
		getBlockCertificates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(number idx.Block, max uint64) ([]cert.BlockCertificate, error) {
			if number == LatestBlock {
				return blocks[len(blocks)-1:], nil
			}
			start := uint64(number)
			end := start + max
			end = min(end, uint64(len(committees)))
			start = min(start, end)
			return blocks[start:end], nil
		}).
		AnyTimes()

	prov.EXPECT().
		getCommitteeCertificates(gomock.Any(), gomock.Any()).
		DoAndReturn(func(from scc.Period, max uint64) ([]cert.CommitteeCertificate, error) {
			start := uint64(from)
			end := start + max
			end = min(end, uint64(len(committees)))
			start = min(start, end)
			return committees[start:end], nil
		}).
		AnyTimes()

	return prov
}

func rotate[T any](list []T) []T {
	res := slices.Clone(list)
	res = append(res[1:], res[0])
	return res
}
