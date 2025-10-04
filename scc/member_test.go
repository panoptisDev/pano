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

package scc

import (
	"math"
	"strings"
	"testing"

	"github.com/panoptisDev/pano/scc/bls"
	"github.com/stretchr/testify/require"
)

func TestMember_Default_IsInvalid(t *testing.T) {
	require.Error(t, Member{}.Validate())
}

func TestMember_String_CanProduceHumanReadableSummary(t *testing.T) {
	require := require.New(t)
	member := Member{}
	print := member.String()

	require.Contains(print, "PublicKey: 0xc000..0000")
	require.Contains(print, "Valid: false")
	require.Contains(print, "VotingPower: 0")

	key := bls.NewPrivateKeyForTests()
	member = Member{
		PublicKey:         key.PublicKey(),
		ProofOfPossession: key.GetProofOfPossession(),
		VotingPower:       12,
	}
	print = member.String()
	require.Contains(print, "PublicKey: 0xa695..8759")
	require.Contains(print, "Valid: true")
	require.Contains(print, "VotingPower: 12")
}

func TestMember_Validate_AcceptsValidMembers(t *testing.T) {
	key := bls.NewPrivateKey()
	pub := key.PublicKey()
	proof := key.GetProofOfPossession()

	tests := map[string]Member{
		"regular": {
			PublicKey:         pub,
			ProofOfPossession: proof,
			VotingPower:       12,
		},
		"huge voting power": {
			PublicKey:         pub,
			ProofOfPossession: proof,
			VotingPower:       math.MaxUint64,
		},
	}

	for name, m := range tests {
		t.Run(name, func(t *testing.T) {
			if err := m.Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMember_Validate_DetectsInvalidMembers(t *testing.T) {
	key := bls.NewPrivateKey()
	pub := key.PublicKey()
	proof := key.GetProofOfPossession()

	tests := map[string]Member{
		"invalid public key": {
			PublicKey:         bls.PublicKey{},
			ProofOfPossession: proof,
			VotingPower:       12,
		},
		"invalid proof of possession": {
			PublicKey:         pub,
			ProofOfPossession: bls.Signature{},
			VotingPower:       12,
		},
		"zero voting power": {
			PublicKey:         pub,
			ProofOfPossession: proof,
			VotingPower:       0,
		},
	}

	for name, m := range tests {
		t.Run(name, func(t *testing.T) {
			err := m.Validate()
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("expected error, got %v", err)
			}
		})
	}
}

func TestMember_Serialization_CanEncodeAndDecodeMember(t *testing.T) {
	key := bls.NewPrivateKey()
	original := Member{
		PublicKey:         key.PublicKey(),
		ProofOfPossession: key.GetProofOfPossession(),
		VotingPower:       12,
	}
	recovered, err := DeserializeMember(original.Serialize())
	require.NoError(t, err)
	require.Equal(t, original, recovered)
}

func TestMember_Deserialize_DetectsEncodingErrors(t *testing.T) {
	encoded := [152]byte{}
	_, err := DeserializeMember(encoded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid public key")

	key := bls.NewPrivateKey()
	*(*[48]byte)(encoded[:]) = key.PublicKey().Serialize()

	_, err = DeserializeMember(encoded)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid proof of possession")
}
