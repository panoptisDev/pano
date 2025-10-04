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

package validatorpk

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestFromString(t *testing.T) {
	require := require.New(t)
	exp := PubKey{
		Type: Types.Secp256k1,
		Raw:  common.FromHex("45b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1"),
	}
	{
		got, err := FromString("c0045b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1")
		require.NoError(err)
		require.Equal(exp, got)
	}
	{
		got, err := FromString("0xc0045b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1")
		require.NoError(err)
		require.Equal(exp, got)
	}
	{
		_, err := FromString("")
		require.Error(err)
	}
	{
		_, err := FromString("0x")
		require.Error(err)
	}
	{
		_, err := FromString("-")
		require.Error(err)
	}
}

func TestString(t *testing.T) {
	require := require.New(t)
	pk := PubKey{
		Type: Types.Secp256k1,
		Raw:  common.FromHex("45b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1"),
	}
	require.Equal("0xc0045b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1", pk.String())
}

func TestRlp(t *testing.T) {
	require := require.New(t)
	v := PubKey{
		Type: Types.Secp256k1,
		Raw:  common.FromHex("45b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1"),
	}
	b, err := rlp.EncodeToBytes(v)
	require.NoError(err)
	require.Equal("f84581c0b841045b86101f804f3f4f2012ef31fff807e87de579a3faa7947d1b487a810e35dc2c3b6071ac465046634b5f4a8e09bf8e1f2e7eccb699356b9e6fd496ca4b1677d1", common.Bytes2Hex(b))
}
