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

	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// PublicAbftAPI provides an API to access consensus related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicAbftAPI struct {
	b Backend
}

// NewPublicAbftAPI creates a new SFC protocol API.
func NewPublicAbftAPI(b Backend) *PublicAbftAPI {
	return &PublicAbftAPI{b}
}

func (s *PublicAbftAPI) GetValidators(ctx context.Context, epoch rpc.BlockNumber) (map[hexutil.Uint64]interface{}, error) {
	bs, es, err := s.b.GetEpochBlockState(ctx, epoch)
	if err != nil {
		return nil, err
	}
	if es == nil {
		return nil, nil
	}
	res := map[hexutil.Uint64]interface{}{}
	for _, vid := range es.Validators.IDs() {
		profiles := es.ValidatorProfiles
		if epoch == rpc.PendingBlockNumber {
			profiles = bs.NextValidatorProfiles
		}
		res[hexutil.Uint64(vid)] = map[string]interface{}{
			"weight": (*hexutil.Big)(profiles[vid].Weight),
			"pubkey": profiles[vid].PubKey.String(),
		}
	}
	return res, nil
}

// GetDowntime returns validator's downtime.
func (s *PublicAbftAPI) GetDowntime(ctx context.Context, validatorID hexutil.Uint) (map[string]interface{}, error) {
	blocks, period, err := s.b.GetDowntime(ctx, idx.ValidatorID(validatorID))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"offlineBlocks": hexutil.Uint64(blocks),
		"offlineTime":   hexutil.Uint64(period),
	}, nil
}

// GetEpochUptime returns validator's epoch uptime in nanoseconds.
func (s *PublicAbftAPI) GetEpochUptime(ctx context.Context, validatorID hexutil.Uint) (hexutil.Uint64, error) {
	v, err := s.b.GetUptime(ctx, idx.ValidatorID(validatorID))
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	return hexutil.Uint64(v.Uint64()), nil
}

// GetOriginatedEpochFee returns validator's originated epoch fee.
func (s *PublicAbftAPI) GetOriginatedEpochFee(ctx context.Context, validatorID hexutil.Uint) (*hexutil.Big, error) {
	v, err := s.b.GetOriginatedFee(ctx, idx.ValidatorID(validatorID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), nil
}
