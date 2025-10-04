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

package ethapi

import (
	"context"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/pano/utils/result"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPanoApi_GetCommitteeCertificates_CanProduceCertificates(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificates := []cert.CommitteeCertificate{
		cert.NewCertificate(cert.CommitteeStatement{Period: 1}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 2}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 3}),
	}

	results := []result.T[cert.CommitteeCertificate]{}
	for _, c := range certificates {
		results = append(results, result.New(c))
	}

	backend.EXPECT().EnumerateCommitteeCertificates(scc.Period(1)).Return(slices.Values(results))

	first := NewIndex(scc.Period(1))
	res, err := api.GetCommitteeCertificates(context.Background(), first, 10)
	require.NoError(t, err)

	require.Len(t, res, len(certificates))
	for i, c := range certificates {
		require.Equal(t, c, res[i].ToCertificate())
	}
}

func TestPanoApi_GetCommitteeCertificates_CanReturnLatestCertificate(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificate := cert.NewCertificate(cert.CommitteeStatement{Period: 12})
	backend.EXPECT().GetLatestCommitteeCertificate().Return(certificate, nil)

	latest := NewLatest[scc.Period]()
	res, err := api.GetCommitteeCertificates(context.Background(), latest, 1)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, certificate, res[0].ToCertificate())
}

func TestPanoApi_GetCommitteeCertificates_ReportsErrorIfLatestCouldNotBeFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	injected := fmt.Errorf("injected error")
	backend.EXPECT().GetLatestCommitteeCertificate().Return(cert.CommitteeCertificate{}, injected)

	latest := NewLatest[scc.Period]()
	_, err := api.GetCommitteeCertificates(context.Background(), latest, 1)
	require.ErrorIs(t, err, injected)
}

func TestPanoApi_GetCommitteeCertificates_CanBeCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificates := []cert.CommitteeCertificate{
		cert.NewCertificate(cert.CommitteeStatement{Period: 1}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 2}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 3}),
	}

	results := []result.T[cert.CommitteeCertificate]{}
	for _, c := range certificates {
		results = append(results, result.New(c))
	}

	backend.EXPECT().EnumerateCommitteeCertificates(scc.Period(1)).Return(slices.Values(results))

	context, cancel := context.WithCancel(context.Background())
	cancel()
	first := NewIndex(scc.Period(1))
	res, err := api.GetCommitteeCertificates(context, first, 10)
	require.ErrorIs(t, err, context.Err())
	require.Empty(t, res)
}

func TestPanoApi_GetCommitteeCertificates_RespectsUserLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificates := []cert.CommitteeCertificate{
		cert.NewCertificate(cert.CommitteeStatement{Period: 1}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 2}),
		cert.NewCertificate(cert.CommitteeStatement{Period: 3}),
	}

	results := []result.T[cert.CommitteeCertificate]{}
	for _, c := range certificates {
		results = append(results, result.New(c))
	}

	backend.EXPECT().EnumerateCommitteeCertificates(scc.Period(1)).
		Return(slices.Values(results)).AnyTimes()

	context := context.Background()
	for _, limit := range []Number{0, 1, 2, 3, math.MaxUint64} {
		first := NewIndex(scc.Period(1))
		res, err := api.GetCommitteeCertificates(context, first, limit)
		require.NoError(t, err)
		want := limit.UInt64()
		if have := uint64(len(certificates)); want > have {
			want = have
		}
		require.Len(t, res, int(want))
	}
}

func TestPanoApi_GetCommitteeCertificates_ReportsFetchErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	injected := fmt.Errorf("injected error")
	results := []result.T[cert.CommitteeCertificate]{
		result.Error[cert.CommitteeCertificate](injected),
	}

	backend.EXPECT().EnumerateCommitteeCertificates(scc.Period(1)).Return(slices.Values(results))

	first := NewIndex(scc.Period(1))
	res, err := api.GetCommitteeCertificates(context.Background(), first, 10)
	require.ErrorIs(t, err, injected)
	require.Empty(t, res)
}

func TestPanoApi_GetBlockCertificate_CanProduceBlockCertificates(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificates := []cert.BlockCertificate{
		cert.NewCertificate(cert.BlockStatement{Number: 1}),
		cert.NewCertificate(cert.BlockStatement{Number: 2}),
		cert.NewCertificate(cert.BlockStatement{Number: 3}),
	}

	results := []result.T[cert.BlockCertificate]{}
	for _, c := range certificates {
		results = append(results, result.New(c))
	}

	backend.EXPECT().EnumerateBlockCertificates(idx.Block(1)).Return(slices.Values(results))

	first := NewIndex(idx.Block(1))
	res, err := api.GetBlockCertificates(context.Background(), first, 10)
	require.NoError(t, err)
	require.Equal(t, len(certificates), len(res))

	for i, c := range certificates {
		require.Equal(t, c, res[i].ToCertificate())
	}
}

func TestPanoApi_GetBlockCertificates_CanReturnLatestCertificate(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	certificate := cert.NewCertificate(cert.BlockStatement{Number: 12})
	backend.EXPECT().GetLatestBlockCertificate().Return(certificate, nil)

	latest := NewLatest[idx.Block]()
	res, err := api.GetBlockCertificates(context.Background(), latest, 1)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, certificate, res[0].ToCertificate())
}

func TestPanoApi_GetBlockCertificates_ReportsErrorIfLatestCouldNotBeFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	backend := NewMockSccApiBackend(ctrl)
	api := NewPublicSccApi(backend)

	injected := fmt.Errorf("injected error")
	backend.EXPECT().GetLatestBlockCertificate().Return(cert.BlockCertificate{}, injected)

	latest := NewLatest[idx.Block]()
	_, err := api.GetBlockCertificates(context.Background(), latest, 1)
	require.ErrorIs(t, err, injected)
}

func TestPeriodNumber_Unmarshaling_HandlesMultipleFormats(t *testing.T) {
	tests := map[string]scc.Period{
		"1":      1,
		"2":      2,
		"012":    012,
		"0x12":   0x12,
		"0b1010": 0b1010,
		"0xaBc":  0xabc,
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			var p PeriodNumber
			err := p.UnmarshalJSON([]byte(`"` + input + `"`))
			require.NoError(t, err)
			require.Equal(t, expected, p.Index())
		})
	}
}

func TestPeriodNumber_Unmarshaling_HandlesLatest(t *testing.T) {
	var p PeriodNumber
	err := p.UnmarshalJSON([]byte(`"latest"`))
	require.NoError(t, err)
	require.True(t, p.IsLatest())
}

func TestPeriodNumber_Unmarshaling_FailsOnTooLargeNumber(t *testing.T) {
	var p PeriodNumber
	err := p.UnmarshalJSON([]byte{0: 1, 9: 0})
	require.Error(t, err)
}

func TestNumber_Unmarshaling_HandlesMultipleFormats(t *testing.T) {
	tests := map[string]uint64{
		"1":      1,
		"2":      2,
		"012":    012,
		"0x12":   0x12,
		"0b1010": 0b1010,
		"0xaBc":  0xabc,
		"max":    math.MaxUint64,
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			var p Number
			err := p.UnmarshalJSON([]byte(`"` + input + `"`))
			require.NoError(t, err)
			require.Equal(t, expected, p.UInt64())
		})
	}
}

func TestNumber_Unmarshaling_FailsOnTooLargeNumber(t *testing.T) {
	var p Number
	err := p.UnmarshalJSON([]byte{0: 1, 9: 0})
	require.Error(t, err)
}

func TestNumber_Unmarshaling_FailsOnNegativeValue(t *testing.T) {
	var p Number
	err := p.UnmarshalJSON([]byte("-1"))
	require.Error(t, err)
}
