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
	"testing"

	"github.com/panoptisDev/pano/evmcore"
	"github.com/panoptisDev/pano/inter/state"
	"github.com/panoptisDev/pano/opera"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEvmProcessorFactory_BeginBlock_CreatesProcessor(t *testing.T) {
	ctrl := gomock.NewController(t)
	chain := NewMockChain(ctrl)

	chain.EXPECT().StateDB().Return(state.NewMockStateDB(ctrl))
	chain.EXPECT().GetCurrentNetworkRules().Return(opera.Rules{}).AnyTimes()
	chain.EXPECT().GetEvmChainConfig(gomock.Any()).Return(&params.ChainConfig{})

	info := BlockInfo{}
	factory := &evmProcessorFactory{chain: chain}
	result := factory.beginBlock(info.toEvmBlock())
	require.NotNil(t, result)
}

func TestEvmProcessor_Run_IfExecutionSucceeds_ReportsSuccessAndGasUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockevmProcessorRunner(ctrl)

	runner.EXPECT().Run(0, nil).Return([]evmcore.ProcessedTransaction{{
		Receipt: &types.Receipt{
			GasUsed: 10,
		},
	}})

	processor := &evmProcessor{processor: runner}
	success, gasUsed := processor.run(nil)
	require.True(t, success)
	require.Equal(t, uint64(10), gasUsed)
}

func TestEvmProcessor_Run_IfExecutionProducesMultipleProcessedTransactions_SumsUpGasUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockevmProcessorRunner(ctrl)

	runner.EXPECT().Run(0, nil).Return([]evmcore.ProcessedTransaction{
		{Receipt: &types.Receipt{GasUsed: 10}},
		{Receipt: nil}, // skipped transaction
		{Receipt: &types.Receipt{GasUsed: 20}},
	})

	processor := &evmProcessor{processor: runner}
	success, gasUsed := processor.run(nil)
	require.True(t, success)
	require.Equal(t, uint64(30), gasUsed)
}

func TestEvmProcessor_Run_IfRequestedTransactionIsNotExecuted_AFailedExecutionIsReported(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockevmProcessorRunner(ctrl)

	tx := &types.Transaction{}
	runner.EXPECT().Run(0, tx).Return([]evmcore.ProcessedTransaction{{
		Transaction: &types.Transaction{}, // different transaction
		Receipt:     &types.Receipt{GasUsed: 10},
	}})

	processor := &evmProcessor{processor: runner}
	success, _ := processor.run(tx)
	require.False(t, success)
}

func TestEvmProcessor_Run_IfExecutionFailed_ReportsAFailedExecution(t *testing.T) {
	t.Run("not processed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		runner := NewMockevmProcessorRunner(ctrl)
		runner.EXPECT().Run(0, nil).Return(nil)
		processor := &evmProcessor{processor: runner}
		success, _ := processor.run(nil)
		require.False(t, success)
	})

	t.Run("no receipt", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		runner := NewMockevmProcessorRunner(ctrl)
		tx := &types.Transaction{}
		runner.EXPECT().Run(0, gomock.Any()).Return([]evmcore.ProcessedTransaction{
			{Transaction: tx, Receipt: nil},
		})
		processor := &evmProcessor{processor: runner}
		success, _ := processor.run(tx)
		require.False(t, success)
	})

	t.Run("different transaction", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		runner := NewMockevmProcessorRunner(ctrl)
		txA := &types.Transaction{}
		txB := &types.Transaction{}
		runner.EXPECT().Run(0, gomock.Any()).Return([]evmcore.ProcessedTransaction{
			{Transaction: txB, Receipt: &types.Receipt{GasUsed: 10}},
		})
		processor := &evmProcessor{processor: runner}
		success, gasUsed := processor.run(txA)
		require.False(t, success)
		require.Equal(t, uint64(10), gasUsed)
	})
}

func TestEvmProcessor_Release_ReleasesStateDb(t *testing.T) {
	ctrl := gomock.NewController(t)
	stateDb := state.NewMockStateDB(ctrl)
	processor := &evmProcessor{stateDb: stateDb}
	stateDb.EXPECT().Release()
	processor.release()
}
