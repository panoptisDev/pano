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
	"math/big"
	"testing"

	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/tests/contracts/contractcreator"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// this constant comes from  https://eips.ethereum.org/EIPS/eip-3860#parameters
const MAX_CODE_SIZE uint64 = 24576
const MAX_INIT_CODE_SIZE uint64 = 2 * MAX_CODE_SIZE

const sufficientGas = uint64(150_000)

func TestInitCodeSizeLimitAndMetered(t *testing.T) {
	requireBase := require.New(t)

	session := getIntegrationTestNetSession(t, opera.GetPanoUpgrades())
	t.Parallel()

	contract, receipt, err := DeployContract(session, contractcreator.DeployContractcreator)
	requireBase.NoError(err)
	requireBase.Equal(types.ReceiptStatusSuccessful, receipt.Status, "failed to deploy contract")

	// run measureGasAndAssign to get cost of deploying a contract without create.
	receipt = createContractSuccessfully(t, session, contract.GetOverheadCost, 0, 0)

	// -- using CREATE instruction
	const wordCostCreate uint64 = 2
	// 32035 is the cost of creating a contract with code according to evm.codes calculator.
	var gasForCreate uint64 = 32035 + receipt.GasUsed
	t.Run("create", func(t *testing.T) {
		t.Parallel()
		// This tests creates multiple sessions in parallel inside sub-tests.
		testForVariant(t, session, contract, contract.CreatetWith, gasForCreate, wordCostCreate)
	})

	// -- using CREATE2 instruction
	// create2 has extra costs for hashing, this is explained in
	// https://eips.ethereum.org/EIPS/eip-3860 and reflected in evm.codes calculator.
	const wordCostCreate2 uint64 = wordCostCreate + 6
	t.Run("create2", func(t *testing.T) {
		t.Parallel()
		// This tests creates multiple sessions in parallel inside sub-tests.
		testForVariant(t, session, contract, contract.Create2With, gasForCreate, wordCostCreate2)
	})

	t.Run("create transaction", func(t *testing.T) {
		t.Parallel()
		// This tests creates multiple sessions in parallel inside sub-tests.
		testForTransaction(t, session)
	})
}

func testForVariant(t *testing.T, session IntegrationTestNetSession,
	contract *contractcreator.Contractcreator, variant variant,
	gasForContract, wordCost uint64) {

	t.Run("charges depending on the init code size", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		createAndGetCost := func(codeLen uint64) uint64 {
			receipt, err := createContractWithCodeLenAndGas(session, variant, codeLen, sufficientGas)
			require.NoError(err)
			require.Equal(types.ReceiptStatusSuccessful, receipt.Status, "failed to create contract with code length ", codeLen)
			log, err := contract.ParseLogCost(*receipt.Logs[0])
			require.NoError(err)
			return log.Cost.Uint64()
		}

		// since memory is expanded in words of 32 bytes, we want to check that the cost is proportional to the number of words.
		// hence 30 bytes fit in 1 word, 42 in 2 words and 90 in 3 words.
		cost1Word := createAndGetCost(30)
		cost2Words := createAndGetCost(42)
		cost3Words := createAndGetCost(90)

		require.Equal(wordCost, cost2Words-cost1Word, "cost difference between 1 and 2")
		require.Equal(wordCost*2, cost3Words-cost1Word, "cost difference between 1 and 3")
	})

	t.Run("fails without enough gas", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		// 4 for a zero byte, 1 to make it fail.
		receipt, err := createContractWithCodeLenAndGas(session, variant, 1, gasForContract-wordCost-1)
		require.NoError(err)
		require.Equal(types.ReceiptStatusFailed, receipt.Status,
			"unexpectedly succeeded to create contract without enough gas")
	})

	t.Run("with max init code size", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		receipt, err := createContractWithCodeLenAndGas(session, variant, MAX_INIT_CODE_SIZE, sufficientGas)
		require.NoError(err)
		require.Equal(types.ReceiptStatusSuccessful, receipt.Status,
			"failed to create contract with code length ", MAX_INIT_CODE_SIZE)
	})

	t.Run("aborts with init code size larger than MAX_INITCODE_SIZE", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		receipt, err := createContractWithCodeLenAndGas(session, variant, MAX_INIT_CODE_SIZE+1, sufficientGas)
		require.NoError(err)
		require.Equal(types.ReceiptStatusFailed, receipt.Status,
			"unexpectedly succeeded to create contract with init code length greater than MAX_INITCODE_SIZE")
	})
}

func testForTransaction(t *testing.T, session IntegrationTestNetSession) {
	t.Run("charges depending on the init code size", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		// transactions charge 4 gas for each zero byte in data.
		const zeroByteCost uint64 = 4
		// create a transaction with 1 byte of code.
		receipt1, err := runTransactionWithCodeSizeAndGas(t, session, 1, sufficientGas)
		require.NoError(err)
		require.Equal(types.ReceiptStatusSuccessful, receipt1.Status,
			"failed on transfer to empty receiver with valid code")

		// createa a transaction with 2 byte of code.
		receipt2, err := runTransactionWithCodeSizeAndGas(t, session, 2, sufficientGas)
		require.NoError(err)
		require.Equal(types.ReceiptStatusSuccessful, receipt2.Status,
			"failed on transfer to empty receiver with valid code")

		difference := receipt2.GasUsed - receipt1.GasUsed
		require.Equal(difference, zeroByteCost,
			"gas difference between 1 and 2 bytes should be 4, instead got", difference)
	})

	t.Run("aborts with init code size larger than MAX_INITCODE_SIZE", func(t *testing.T) {
		require := require.New(t)
		session := session.SpawnSession(t)
		t.Parallel()

		// as specified in https://eips.ethereum.org/EIPS/eip-3860#rules,
		// this is similar to transactions considered invalid for not meeting the intrinsic gas cost requirement.
		_, err := runTransactionWithCodeSizeAndGas(t, session, MAX_INIT_CODE_SIZE+1, sufficientGas)
		require.ErrorContains(
			err,
			"max initcode size exceeded",
			"unexpectedly succeeded to create contract with init code larger than MAX_INITCODE_SIZE",
		)
	})
}

func createContractSuccessfully(t *testing.T, session IntegrationTestNetSession,
	variant variant, codeLen, gasLimit uint64) *types.Receipt {

	receipt, err := createContractWithCodeLenAndGas(session, variant, codeLen, gasLimit)
	require := require.New(t)
	require.NoError(err)
	require.Equal(types.ReceiptStatusSuccessful, receipt.Status, "failed to create contract with code length ", codeLen)
	return receipt
}

func createContractWithCodeLenAndGas(session IntegrationTestNetSession,
	variant variant, codeLen, gasLimit uint64) (*types.Receipt, error) {
	return session.Apply(func(opts *bind.TransactOpts) (*types.Transaction, error) {
		opts.GasLimit = gasLimit
		return variant(opts, big.NewInt(int64(codeLen)))
	})
}

type variant func(opts *bind.TransactOpts, codeSize *big.Int) (*types.Transaction, error)

func runTransactionWithCodeSizeAndGas(
	t *testing.T,
	session IntegrationTestNetSession,
	codeSize, gas uint64,
) (*types.Receipt, error) {
	require := require.New(t)
	// these values are needed for the transaction but are irrelevant for the test
	client, err := session.GetClient()
	require.NoError(err, "failed to connect to the network:")
	defer client.Close()

	chainId, err := client.ChainID(t.Context())
	require.NoError(err, "failed to get chain ID::")

	nonce, err := client.NonceAt(t.Context(), session.GetSessionSponsor().Address(), nil)
	require.NoError(err, "failed to get nonce:")

	price, err := client.SuggestGasPrice(t.Context())
	require.NoError(err, "failed to get gas price:")
	// ---------

	transaction, err := types.SignTx(types.NewTx(&types.AccessListTx{
		ChainID:  chainId,
		Gas:      gas,
		GasPrice: price,
		To:       nil,
		Nonce:    nonce,
		Data:     make([]byte, codeSize),
	}), types.NewLondonSigner(chainId), session.GetSessionSponsor().PrivateKey)
	require.NoError(err, "failed to sign transaction:")
	return session.Run(transaction)
}
