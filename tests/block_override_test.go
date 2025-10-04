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

package tests

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/panoptisDev/pano/ethapi"
	"github.com/panoptisDev/pano/opera"
	block_override "github.com/panoptisDev/pano/tests/contracts/blockoverride"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	req "github.com/stretchr/testify/require"
)

const (
	contractFunction_getBlockParameters = "0xa3289b77"
)

func TestBlockOverride(t *testing.T) {
	require := req.New(t)
	session := getIntegrationTestNetSession(t, opera.GetPanoUpgrades())
	t.Parallel()

	// Deploy the block override observer contract.
	contract, receipt, err := DeployContract(session, block_override.DeployBlockOverride)
	require.NoError(err, "failed to deploy contract; %v", err)
	require.Equal(types.ReceiptStatusSuccessful, receipt.Status, "Deployment Unsuccessful")
	contractAddress := receipt.ContractAddress

	//Call contract method to be sure it is deployed.
	receiptLog, err := session.Apply(contract.LogBlockNumber)
	require.NoError(err, "failed to create a log")
	require.Equal(types.ReceiptStatusSuccessful, receiptLog.Status,
		"failed to create a log")

	// Need valid block number for eth_call and debug_traceCall
	blockNumber := receiptLog.BlockNumber.Uint64()

	// Set parameters to be overridden
	time := uint64(1234)
	gasLimit := uint64(567890)
	blockOverrides := &ethapi.BlockOverrides{
		Number:      (*hexutil.Big)(big.NewInt(42)),
		Difficulty:  (*hexutil.Big)(big.NewInt(1)),
		Time:        (*hexutil.Uint64)(&time),
		GasLimit:    (*hexutil.Uint64)(&gasLimit),
		Coinbase:    &common.Address{1},
		Random:      &common.Hash{2},
		BaseFee:     (*hexutil.Big)(big.NewInt(1_000)),
		BlobBaseFee: (*hexutil.Big)(big.NewInt(100)),
	}

	t.Run("eth_call block override", func(t *testing.T) {
		newClient, err := session.GetClient()
		require.NoError(err, "failed to get client; %v", err)
		defer newClient.Close()
		compareCalls(t, newClient.Client(), contractAddress, blockNumber, blockOverrides, makeEthCall)
	})

	t.Run("debug_traceCall block override", func(t *testing.T) {
		newClient, err := session.GetClient()
		require.NoError(err, "failed to get client; %v", err)
		defer newClient.Close()
		compareCalls(t, newClient.Client(), contractAddress, blockNumber, blockOverrides, makeDebugTraceCall)
	})

}

func compareCalls(t *testing.T, rpcClient *rpc.Client, contractAddress common.Address, blockNumber uint64, blockOverrides *ethapi.BlockOverrides,
	callFunc func(t *testing.T, rpcClient *rpc.Client, contractAddress common.Address, blockNumber uint64, blockOverrides *ethapi.BlockOverrides) (BlockParameters, error)) {

	require := req.New(t)
	params, err := callFunc(t, rpcClient, contractAddress, blockNumber, nil)
	require.NoError(err, "failed to make eth_call; %v", err)

	paramsOverride, err := callFunc(t, rpcClient, contractAddress, blockNumber, blockOverrides)
	require.NoError(err, "failed to make eth_call; %v", err)

	t.Logf("params: %v", params)
	t.Logf("params: %v", paramsOverride)

	err = checkAllFieldsAreDifferent(params, paramsOverride)
	require.NoError(err, "failed to compare block parameters; %v", err)

	err = checkOverrides(paramsOverride, *blockOverrides)
	require.NoError(err, "failed to compare block parameters; %v", err)
}

// BlockParameters is a struct created from a response from a call
type BlockParameters struct {
	Number      *big.Int
	Difficulty  *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	Coinbase    common.Address
	Random      common.Hash
	BaseFee     *big.Int
	BlobBaseFee *big.Int
}

func getBlockParameters(data []byte) (BlockParameters, error) {

	if len(data) != 256 {
		return BlockParameters{}, fmt.Errorf("invalid data length: %d, expected 256", len(data))
	}

	return BlockParameters{
		Number:      new(big.Int).SetBytes(data[:32]),
		Difficulty:  new(big.Int).SetBytes(data[32:64]),
		Time:        new(big.Int).SetBytes(data[64:96]),
		GasLimit:    new(big.Int).SetBytes(data[96:128]),
		Coinbase:    common.BytesToAddress(data[128:160]),
		Random:      common.BytesToHash(data[160:192]),
		BaseFee:     new(big.Int).SetBytes(data[192:224]),
		BlobBaseFee: new(big.Int).SetBytes(data[224:]),
	}, nil
}

func (bp *BlockParameters) String() string {
	return fmt.Sprintf(
		"number: %v difficulty: %v time: %v gasLimit: %v coinbase: %v random: %v baseFee: %v blobbasefee: %v",
		bp.Number, bp.Difficulty, bp.Time, bp.GasLimit, bp.Coinbase, bp.Random, bp.BaseFee, bp.BlobBaseFee,
	)
}

func getFunctionCallParameters(contractAddress *common.Address) map[string]interface{} {
	return map[string]interface{}{
		"to":   contractAddress.String(),
		"data": contractFunction_getBlockParameters,
	}
}

func makeEthCall(t *testing.T, rpcClient *rpc.Client, contractAddress common.Address, blockNumber uint64, blockOverrides *ethapi.BlockOverrides) (BlockParameters, error) {
	require := req.New(t)

	var res interface{}
	err := rpcClient.Call(&res, "eth_call", getFunctionCallParameters(&contractAddress), hexutil.EncodeUint64(blockNumber), nil, blockOverrides)
	require.NoError(err, "failed to call eth_call; %v", err)

	if s, ok := res.(string); ok {
		b, err := hexutil.Decode(s)
		require.NoError(err, "failed to decode result hex; %v", err)

		params, err := getBlockParameters(b)
		require.NoError(err, "failed to decode block parameters; %v", err)

		return params, nil
	} else {
		return BlockParameters{}, fmt.Errorf("invalid result type: %T", res)
	}
}

func makeDebugTraceCall(t *testing.T, rpcClient *rpc.Client, contractAddress common.Address, blockNumber uint64, blockOverrides *ethapi.BlockOverrides) (BlockParameters, error) {
	require := req.New(t)

	traceConfig := &ethapi.TraceCallConfig{
		BlockOverrides: blockOverrides,
	}

	var res interface{}
	err := rpcClient.Call(&res, "debug_traceCall", getFunctionCallParameters(&contractAddress), hexutil.EncodeUint64(blockNumber), traceConfig)
	require.NoError(err, "failed to call eth_call; %v", err)

	if data, ok := res.(map[string]interface{}); ok {
		if s, ok := data["returnValue"].(string); ok {
			if !strings.HasPrefix(s, "0x") {
				s = "0x" + s
			}
			b, err := hexutil.Decode(s)
			require.NoError(err, "failed to decode result hex; %v", err)

			params, err := getBlockParameters(b)
			require.NoError(err, "failed to decode block parameters; %v", err)

			return params, nil
		} else {
			return BlockParameters{}, fmt.Errorf("invalid result type: %T", res)
		}
	} else {
		return BlockParameters{}, fmt.Errorf("invalid result type: %T", res)
	}
}

// checkAllFieldsAreDifferent compares two BlockParameters objects and returns an error if any fields were not overridden
func checkAllFieldsAreDifferent(params1, params2 BlockParameters) error {
	if params1.Number.Cmp(params2.Number) == 0 {
		return fmt.Errorf("Number field was not overridden: %v", params1.Number)
	}
	if params1.Difficulty.Cmp(params2.Difficulty) == 0 {
		return fmt.Errorf("Difficulty field was not overridden: %v", params1.Difficulty)
	}
	if params1.Time.Cmp(params2.Time) == 0 {
		return fmt.Errorf("Time field was not overridden: %v", params1.Time)
	}
	if params1.GasLimit.Cmp(params2.GasLimit) == 0 {
		return fmt.Errorf("GasLimit field was not overridden: %v", params1.GasLimit)
	}
	if params1.Coinbase == params2.Coinbase {
		return fmt.Errorf("Coinbase field was not overridden: %v", params1.Coinbase)
	}
	if params1.Random == params2.Random {
		return fmt.Errorf("Random field was not overridden: %v", params1.Random)
	}
	if params1.BaseFee.Cmp(params2.BaseFee) == 0 {
		return fmt.Errorf("BaseFee field was not overridden: %v", params1.BaseFee)
	}
	if params1.BlobBaseFee.Cmp(params2.BlobBaseFee) == 0 {
		return fmt.Errorf("BlobBaseFee field was not overridden: %v", params1.BlobBaseFee)
	}
	return nil
}

// checkOverrides checks that the given BlockParameters are the same as the overrides
func checkOverrides(params BlockParameters, overrides ethapi.BlockOverrides) error {
	if want, got := overrides.Number, params.Number; got.Cmp((*big.Int)(want)) != 0 {
		return fmt.Errorf("Number override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.Time, params.Time; got.Uint64() != (uint64(*want)) {
		return fmt.Errorf("Time override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.GasLimit, params.GasLimit; got.Uint64() != (uint64(*want)) {
		return fmt.Errorf("GasLimit override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.Coinbase, params.Coinbase; *want != got {
		return fmt.Errorf("Coinbase override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.Random, params.Random; *want != got {
		return fmt.Errorf("Random override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.BaseFee, params.BaseFee; got.Cmp((*big.Int)(want)) != 0 {
		return fmt.Errorf("BaseFee override incorrect, wanted %v, got %v", want, got)
	}
	if want, got := overrides.BlobBaseFee, params.BlobBaseFee; got.Cmp((*big.Int)(want)) != 0 {
		return fmt.Errorf("BlobBaseFee override incorrect, wanted %v, got %v", want, got)
	}
	return nil
}
