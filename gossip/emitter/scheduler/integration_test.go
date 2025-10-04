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

package scheduler

import (
	"math"
	"testing"

	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// This file contains a range of integration tests combining the scheduler
// implementation with the real-world implementations of the evaluator. While
// specifics are tested in individual unit tests, the focus of these tests it to
// verify the overall integration of the components.

func TestIntegration_NoTransactions_ProducesAnEmptySchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	chain := NewMockChain(ctrl)
	state := state.NewMockStateDB(ctrl)

	chain.EXPECT().GetCurrentNetworkRules().Return(opera.Rules{}).AnyTimes()
	chain.EXPECT().GetEvmChainConfig(gomock.Any()).Return(&params.ChainConfig{})
	chain.EXPECT().StateDB().Return(state)
	state.EXPECT().Release()

	scheduler := NewScheduler(chain)
	require.Empty(t, scheduler.Schedule(
		t.Context(),
		&BlockInfo{},
		&fakeTxCollection{},
		Limits{
			Gas:  100_000_000,
			Size: 100_000,
		},
	))
}

func TestIntegration_OneTransactions_ProducesScheduleWithOneTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	chain := NewMockChain(ctrl)
	state := state.NewMockStateDB(ctrl)

	chain.EXPECT().GetCurrentNetworkRules().Return(opera.Rules{}).AnyTimes()
	chain.EXPECT().GetEvmChainConfig(gomock.Any()).Return(&params.ChainConfig{})
	chain.EXPECT().StateDB().Return(state)

	// The scheduler configured for production is running transactions on the
	// actual EVM state processor. Thus, various StateDB interactions are
	// expected, yet the specific details are not important for this test. The
	// main objective is to make the one transaction to be scheduled pass the
	// execution to make it eligible to be included in the resulting schedule.
	any := gomock.Any()
	state.EXPECT().SetTxContext(any, any)
	state.EXPECT().GetBalance(any).Return(uint256.NewInt(math.MaxInt64)).AnyTimes()
	state.EXPECT().AddBalance(any, any, any).AnyTimes()
	state.EXPECT().SubBalance(any, any, any).AnyTimes()
	state.EXPECT().Prepare(any, any, any, any, any, any).AnyTimes()
	state.EXPECT().GetNonce(any).AnyTimes()
	state.EXPECT().SetNonce(any, any, any).AnyTimes()
	state.EXPECT().GetCodeHash(any).Return(types.EmptyCodeHash).AnyTimes()
	state.EXPECT().GetCode(any).Return(nil).AnyTimes()
	state.EXPECT().Snapshot().AnyTimes()
	state.EXPECT().Exist(any).Return(true).AnyTimes()
	state.EXPECT().GetRefund().AnyTimes()
	state.EXPECT().GetLogs(any, any).AnyTimes()
	state.EXPECT().EndTransaction().AnyTimes()
	state.EXPECT().TxIndex().AnyTimes()
	state.EXPECT().Release()

	txs := []*types.Transaction{
		types.NewTx(&types.LegacyTx{
			To:  &common.Address{},
			Gas: 21_000,
		}),
	}

	schedule := NewScheduler(chain).Schedule(
		t.Context(),
		&BlockInfo{
			GasLimit: 100_000_000,
		},
		&fakeTxCollection{transactions: txs},
		Limits{
			Gas:  100_000_000,
			Size: 100_000,
		},
	)

	require.Equal(t, txs, schedule)
}
