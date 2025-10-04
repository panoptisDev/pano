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

package config

import (
	"bytes"
	"testing"

	"github.com/panoptisDev/lachesis-base/abft"
	"github.com/panoptisDev/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/gossip"
	"github.com/panoptisDev/pano/gossip/emitter"
	"github.com/panoptisDev/pano/vecmt"
)

func TestConfigFile(t *testing.T) {
	cacheRatio := cachescale.Ratio{
		Base:   uint64(DefaultCacheSize*1 - ConstantCacheSize),
		Target: uint64(DefaultCacheSize*2 - ConstantCacheSize),
	}

	src := Config{
		Node:          DefaultNodeConfig(),
		Opera:         gossip.DefaultConfig(cacheRatio),
		Emitter:       emitter.DefaultConfig(),
		TxPool:        evmcore.DefaultTxPoolConfig,
		OperaStore:    gossip.DefaultStoreConfig(cacheRatio),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(cacheRatio),
		VectorClock:   vecmt.DefaultConfig(cacheRatio),
	}

	canonical := func(nn []*enode.Node) []*enode.Node {
		if len(nn) == 0 {
			return []*enode.Node{}
		}
		return nn
	}

	for name, val := range map[string][]*enode.Node{
		"Nil":     nil,
		"Empty":   {},
		"Default": asDefault,
		"UserDefined": {enode.MustParse(
			"enr:-J-4QJmPmUmu14Pn7gUtRNfKHaWFQpcX6fgqrNheDSWUN6giKtix8Lh6EKfymTdXCI5HKGmyl0C5eOKvem5xdC70hLEBgmlkgnY0gmlwhMCoAQKFb3BlcmHHxoQHxfIKgIlzZWNwMjU2azGhAjYQROWoAXivxhtYYBXGXzQrBTAHGJT9XPP69oUzDDWwhHNuYXDAg3RjcIITuoN1ZHCCE7o",
		)},
	} {
		t.Run(name+"BootstrapNodes", func(t *testing.T) {
			require := require.New(t)

			src.Node.P2P.BootstrapNodes = val
			src.Node.P2P.BootstrapNodesV5 = val

			stream, err := TomlSettings.Marshal(&src)
			require.NoError(err)

			var got Config
			err = TomlSettings.NewDecoder(bytes.NewReader(stream)).Decode(&got)
			require.NoError(err)

			{ // toml workaround
				src.Node.P2P.BootstrapNodes = canonical(src.Node.P2P.BootstrapNodes)
				got.Node.P2P.BootstrapNodes = canonical(got.Node.P2P.BootstrapNodes)
				src.Node.P2P.BootstrapNodesV5 = canonical(src.Node.P2P.BootstrapNodesV5)
				got.Node.P2P.BootstrapNodesV5 = canonical(got.Node.P2P.BootstrapNodesV5)
			}

			require.Equal(src.Node.P2P.BootstrapNodes, got.Node.P2P.BootstrapNodes)
			require.Equal(src.Node.P2P.BootstrapNodesV5, got.Node.P2P.BootstrapNodesV5)
		})
	}
}
