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
	"fmt"
	"math"
	"testing"

	"github.com/panoptisDev/carmen/go/carmen"
	"github.com/panoptisDev/pano/ethapi"
	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_NewServer_CanInitializeFromUrl(t *testing.T) {
	require := require.New(t)

	server, err := newServerFromURL("http://localhost:8545")
	t.Cleanup(server.close)
	require.NoError(err)
	require.NotNil(server)
	require.False(server.IsClosed())
}

func TestServer_NewServer_ReportsErrorForNilClient(t *testing.T) {
	require := require.New(t)

	server, err := newServerFromClient(nil)
	require.Error(err)
	require.Nil(server)
}

func TestServer_NewServer_ReportsErrorForInvalidURL(t *testing.T) {
	require := require.New(t)

	server, err := newServerFromURL("not-a-url")
	require.Error(err)
	require.Nil(server)
}

func TestServer_IsClosed_Reports(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	t.Run("server with client is not closed", func(t *testing.T) {
		client := NewMockrpcClient(ctrl)
		client.EXPECT().Close().AnyTimes()
		server, err := newServerFromClient(client)
		require.NoError(err)
		require.False(server.IsClosed())
		server.close()
	})

	t.Run("server with client can be closed", func(t *testing.T) {
		client := NewMockrpcClient(ctrl)
		client.EXPECT().Close()
		server, err := newServerFromClient(client)
		require.NoError(err)
		require.False(server.IsClosed())
		server.close()
		require.True(server.IsClosed())
	})

	t.Run("closed server can be re-closed", func(t *testing.T) {
		client := NewMockrpcClient(ctrl)
		client.EXPECT().Close()
		server, err := newServerFromClient(client)
		require.NoError(err)
		server.close()
		require.True(server.IsClosed())
		server.close()
		require.True(server.IsClosed())
	})
}

func TestServer_FailsToRequestAfterClose(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)

	server, err := newServerFromClient(client)
	require.NoError(err)

	// close server
	client.EXPECT().Close()
	server.close()

	// get committee certificates
	_, err = server.getCommitteeCertificates(0, 1)
	require.Error(err)

	// get block certificates
	_, err = server.getBlockCertificates(0, 1)
	require.Error(err)
}

func TestServer_GetCertificates_PropagatesErrorFromClientCall(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)

	committeeError := fmt.Errorf("committee error")
	client.EXPECT().Call(gomock.Any(), "pano_getCommitteeCertificates",
		gomock.Any(), gomock.Any()).Return(committeeError)

	blockError := fmt.Errorf("block error")
	client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
		gomock.Any(), gomock.Any()).Return(blockError)

	server, err := newServerFromClient(client)
	require.NoError(err)

	// get committee certificates
	_, err = server.getCommitteeCertificates(0, 1)
	require.ErrorIs(err, committeeError)

	// get block certificates
	_, err = server.getBlockCertificates(0, 1)
	require.ErrorIs(err, blockError)
}

func TestServer_GetCommitteeCertificates_ReportsCorruptedCertificatesOutOfOrder(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)

	tests := [][]ethapi.CommitteeCertificate{
		{
			ethapi.CommitteeCertificate{Period: uint64(1)},
		},
		{
			ethapi.CommitteeCertificate{Period: uint64(0)},
			ethapi.CommitteeCertificate{Period: uint64(2)},
		},
	}

	for _, committees := range tests {
		client := NewMockrpcClient(ctrl)
		server, err := newServerFromClient(client)
		require.NoError(err)

		// client setup
		client.EXPECT().Call(gomock.Any(), "pano_getCommitteeCertificates",
			gomock.Any(), gomock.Any()).
			DoAndReturn(
				func(result *[]ethapi.CommitteeCertificate, method string, args ...interface{}) error {
					*result = committees
					return nil
				})

		// get committee certificates
		_, err = server.getCommitteeCertificates(0, 3)
		require.ErrorContains(err, "out of order")
	}
}

func TestServer_GetCommitteeCertificates_DropsExcessCertificates(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	client.EXPECT().Call(gomock.Any(), "pano_getCommitteeCertificates",
		gomock.Any(), gomock.Any()).DoAndReturn(
		func(result *[]ethapi.CommitteeCertificate, method string, args ...interface{}) error {
			*result = []ethapi.CommitteeCertificate{
				{Period: uint64(0)},
				{Period: uint64(1)},
			}
			return nil
		})

	// get committee certificates
	certs, err := server.getCommitteeCertificates(0, 1)
	require.NoError(err)
	require.Len(certs, 1)
}

func TestServer_GetBlockCertificates_ReportsCorruptedCertificatesOutOfOrder(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	tests := [][]ethapi.BlockCertificate{
		{
			ethapi.BlockCertificate{Number: uint64(1)},
		},
		{
			ethapi.BlockCertificate{Number: uint64(0)},
			ethapi.BlockCertificate{Number: uint64(2)},
		},
	}

	for _, blocks := range tests {
		client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
			gomock.Any(), gomock.Any()).DoAndReturn(
			func(result *[]ethapi.BlockCertificate, method string, args ...interface{}) error {
				*result = blocks
				return nil
			})

		// get block certificates
		_, err := server.getBlockCertificates(0, 3)
		require.ErrorContains(err, "out of order")
	}
}

func TestServer_GetBlockCertificates_DropsExcessCertificates(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
		gomock.Any(), gomock.Any()).DoAndReturn(
		func(result *[]ethapi.BlockCertificate, method string, args ...interface{}) error {
			*result = []ethapi.BlockCertificate{
				{Number: uint64(0)},
				{Number: uint64(1)},
			}
			return nil
		})

	// get block certificates
	certs, err := server.getBlockCertificates(0, 1)
	require.NoError(err)
	require.Len(certs, 1)
}

func TestServer_GetBlockCertificates_FailsWhenNoCertificatesReturned(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
		gomock.Any(), gomock.Any()).DoAndReturn(
		func(result *[]ethapi.BlockCertificate, method string, args ...interface{}) error {
			*result = []ethapi.BlockCertificate{}
			return nil
		})

	// get block certificates
	_, err = server.getBlockCertificates(0, 1)
	require.ErrorContains(err, "no block certificates found")
}

func TestServer_GetBlockCertificates_CanFetchLatestBlock(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	latestBlockNumber := idx.Block(1024)
	// block certificates
	client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
		"latest", "0x1").DoAndReturn(
		func(result *[]ethapi.BlockCertificate, method string, args ...interface{}) error {
			*result = []ethapi.BlockCertificate{
				{Number: uint64(latestBlockNumber)},
			}
			return nil
		})

	// get block certificates
	blockCerts, err := server.getBlockCertificates(LatestBlock, 1)
	require.NoError(err)
	require.Len(blockCerts, 1)
	require.Equal(latestBlockNumber, blockCerts[0].Subject().Number)
}

func TestServer_GetCertificates_IgnoresRequestForZeroCertificates(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	// get committee certificates
	_, err = server.getCommitteeCertificates(0, 0)
	require.NoError(err)

	// get block certificates
	_, err = server.getBlockCertificates(0, 0)
	require.NoError(err)
}

func TestServer_GetCertificates_ReturnsCertificates(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	// committee certificates
	client.EXPECT().Call(gomock.Any(), "pano_getCommitteeCertificates",
		gomock.Any(), gomock.Any()).DoAndReturn(
		func(result *[]ethapi.CommitteeCertificate, method string, args ...interface{}) error {
			*result = []ethapi.CommitteeCertificate{
				{Period: uint64(0)},
				{Period: uint64(1)},
			}
			return nil
		})

	// get committee certificates
	comCerts, err := server.getCommitteeCertificates(0, 2)
	require.NoError(err)
	require.Len(comCerts, 2)
	require.Equal(scc.Period(0), comCerts[0].Subject().Period)
	require.Equal(scc.Period(1), comCerts[1].Subject().Period)

	// block certificates
	client.EXPECT().Call(gomock.Any(), "pano_getBlockCertificates",
		gomock.Any(), gomock.Any()).DoAndReturn(
		func(result *[]ethapi.BlockCertificate, method string, args ...interface{}) error {
			*result = []ethapi.BlockCertificate{
				{Number: uint64(0)},
				{Number: uint64(1)},
			}
			return nil
		})

	// get block certificates
	blockCerts, err := server.getBlockCertificates(0, 2)
	require.NoError(err)
	require.Len(blockCerts, 2)
}

func TestServer_GetAccountProof_PropagatesClientError(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)

	someError := fmt.Errorf("some error")
	addr := common.Address{0x1}
	client.EXPECT().Call(
		gomock.Any(), // any result variable
		"eth_getProof",
		fmt.Sprintf("%v", addr),
		gomock.Any(), // any storage key
		"latest").
		Return(someError)

	_, err = server.getAccountProof(addr, math.MaxUint64)
	require.ErrorIs(err, someError)
}

func TestServer_GetAccountProof_FailsToDecodeAddressProof(t *testing.T) {
	// setup
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)
	// expexted error
	addr := common.Address{0x1}
	client.EXPECT().Call(
		gomock.Any(),
		"eth_getProof",
		fmt.Sprintf("%v", addr),
		gomock.Any(),
		"latest").DoAndReturn(
		func(result *struct {
			AccountProof []string
		}, method string, args ...interface{}) error {
			// invalid proof
			result.AccountProof = []string{"invalid"}
			return nil
		})

	got, err := server.getAccountProof(addr, math.MaxUint64)
	require.ErrorContains(err, "failed to decode proof element")
	require.Nil(got)
}

func TestServer_GetAccountProof_ReportsInvalidProof(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)
	addr := common.Address{0x1}
	client.EXPECT().Call(
		gomock.Any(),
		"eth_getProof",
		fmt.Sprintf("%v", addr),
		gomock.Any(),
		"latest").DoAndReturn(
		func(result *struct {
			AccountProof []string
		}, method string, args ...interface{}) error {
			// invalid proof
			result.AccountProof = []string{"0x01", "0x02"}
			return nil
		})

	got, err := server.getAccountProof(addr, math.MaxUint64)
	require.ErrorContains(err, "invalid proof")
	require.Nil(got)
}

func TestServer_GetAccountProof_ReturnsAccountProof(t *testing.T) {
	require := require.New(t)
	ctrl := gomock.NewController(t)
	client := NewMockrpcClient(ctrl)
	server, err := newServerFromClient(client)
	require.NoError(err)
	addr := common.Address{0x1}
	elementsString := []string{}

	client.EXPECT().Call(
		gomock.Any(),
		"eth_getProof",
		fmt.Sprintf("%v", addr),
		gomock.Any(),
		"latest").DoAndReturn(
		func(result *struct {
			AccountProof []string
		}, method string, args ...interface{}) error {
			result.AccountProof = elementsString
			return nil
		})

	got, err := server.getAccountProof(addr, math.MaxUint64)
	require.NoError(err)
	// decode elements for the proof.
	elements := []carmen.Bytes{}

	want := carmen.CreateWitnessProofFromNodes(elements...)
	require.Equal(want, got)
}
