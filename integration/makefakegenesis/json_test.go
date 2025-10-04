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

package makefakegenesis

import (
	"testing"

	"github.com/panoptisDev/pano/opera"
	"github.com/stretchr/testify/require"
)

func TestJsonGenesis_CanApplyGeneratedFakeJsonGensis(t *testing.T) {
	genesis := GenerateFakeJsonGenesis(1, opera.GetPanoUpgrades())
	_, err := ApplyGenesisJson(genesis)
	require.NoError(t, err)
}

func TestJsonGenesis_AcceptsGenesisWithoutCommittee(t *testing.T) {
	genesis := GenerateFakeJsonGenesis(1, opera.GetPanoUpgrades())
	genesis.GenesisCommittee = nil
	_, err := ApplyGenesisJson(genesis)
	require.NoError(t, err)
}

func TestJsonGenesis_Network_Rules_Validated_Allegro_Only(t *testing.T) {
	tests := map[string]struct {
		featureSet opera.Upgrades
		assert     func(t *testing.T, err error)
	}{
		"pano": {
			featureSet: opera.GetPanoUpgrades(),
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		"allegro": {
			featureSet: opera.GetAllegroUpgrades(),
			assert: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "LLR upgrade is not supported")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			genesis := GenerateFakeJsonGenesis(1, test.featureSet)
			genesis.Rules.Upgrades.Llr = true // LLR is not supported in Allegro and Pano
			_, err := ApplyGenesisJson(genesis)
			test.assert(t, err)
		})
	}
}
