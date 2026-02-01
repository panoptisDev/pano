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

package emitter

import (
	"testing"

	"github.com/panoptisDev/pano/gossip/emitter/config"
	"github.com/stretchr/testify/require"
)

func Test_DefaultMaxTxsPerAddress_Equals_txTurnNonces(t *testing.T) {

	// Although MaxTxsPerAddress can be configured, having a value less than txTurnNonces
	// could yield performance issues when dispatching batches of transactions.
	// MaxTxsPerAddress should be greater or equal to txTurnNonces to ensure timely
	// emission of transactions. Default value for this parameter should be exactly txTurnNonces.

	defaultConfig := config.DefaultConfig()
	require.EqualValues(t, txTurnNonces, defaultConfig.MaxTxsPerAddress, "Default MaxTxsPerAddress should equal txTurnNonces")
}
