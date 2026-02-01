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

package ibr_test

import (
	"bytes"
	"github.com/panoptisDev/pano/inter/ibr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"testing"
)

// Verify that [] is equivalent to nil for log's Topics in the genesis file
func TestNilTopicsDoesNotMatterInGenesis(t *testing.T) {
	rlp1, err := rlp.EncodeToBytes(ibr.LlrIdxFullBlockRecord{
		LlrFullBlockRecord: ibr.LlrFullBlockRecord{
			Receipts: []*types.ReceiptForStorage{
				{
					Logs: []*types.Log{
						{
							Topics: nil,
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	rlp2, err := rlp.EncodeToBytes(ibr.LlrIdxFullBlockRecord{
		LlrFullBlockRecord: ibr.LlrFullBlockRecord{
			Receipts: []*types.ReceiptForStorage{
				{
					Logs: []*types.Log{
						{
							Topics: []common.Hash{},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rlp1, rlp2) {
		t.Errorf("serialized byte slices does not match: %x != %x", rlp1, rlp2)
	}
}
