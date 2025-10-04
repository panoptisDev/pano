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
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/panoptisDev/pano/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestSetTransactionDefaults_CanInitializeAllTransactionTypes(t *testing.T) {
	session := getIntegrationTestNetSession(t, opera.GetAllegroUpgrades())
	t.Parallel()

	client, err := session.GetClient()
	require.NoError(t, err)
	defer client.Close()

	chainId, err := client.ChainID(t.Context())
	require.NoError(t, err)

	type modificationFunction func(t *testing.T, tx *types.TxData)

	transactionType := func(txType byte) modificationFunction {
		return func(t *testing.T, tx *types.TxData) {
			switch txType {
			case types.LegacyTxType:
				*tx = &types.LegacyTx{}
			case types.AccessListTxType:
				*tx = &types.AccessListTx{}
			case types.DynamicFeeTxType:
				*tx = &types.DynamicFeeTx{}
			case types.BlobTxType:
				*tx = &types.BlobTx{}
			case types.SetCodeTxType:
				*tx = &types.SetCodeTx{}
			default:
				t.Fatalf("unexpected transaction type: %d", txType)
			}
		}
	}

	noData := func() modificationFunction {
		return func(t *testing.T, tx *types.TxData) {}
	}

	withData := func(size, zeroes int) modificationFunction {

		makeData := func(t *testing.T, size, numZeroes int) []byte {
			require.GreaterOrEqual(t, size, zeroes)
			// please add one 0, so that init-code starts with STOP
			require.Greater(t, zeroes, 0)

			data := make([]byte, size)
			for i := numZeroes; i < size; i++ {
				data[i] = 1
			}
			return data
		}

		return func(t *testing.T, tx *types.TxData) {
			switch tx := (*tx).(type) {
			case *types.LegacyTx:
				tx.Data = makeData(t, size, zeroes)
			case *types.AccessListTx:
				tx.Data = makeData(t, size, zeroes)
			case *types.DynamicFeeTx:
				tx.Data = makeData(t, size, zeroes)
			case *types.BlobTx:
				tx.Data = makeData(t, size, zeroes)
			case *types.SetCodeTx:
				tx.Data = makeData(t, size, zeroes)
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
		}
	}

	noAccessList := func() modificationFunction {
		return func(t *testing.T, tx *types.TxData) {}
	}

	withAccessList := func(accounts, keysPerAccount int) modificationFunction {

		makeAccessList := func(t *testing.T, accounts, keysPerAccount int) []types.AccessTuple {
			accessList := make([]types.AccessTuple, accounts)
			for i := range accessList {
				accessList[i] = types.AccessTuple{
					Address:     NewAccount().Address(),
					StorageKeys: make([]common.Hash, keysPerAccount),
				}
				for j := range accessList[i].StorageKeys {
					accessList[i].StorageKeys[j] = common.BigToHash(big.NewInt(int64(j)))
				}
			}
			return accessList
		}
		return func(t *testing.T, tx *types.TxData) {
			switch tx := (*tx).(type) {
			case *types.LegacyTx:
				// ignore
			case *types.AccessListTx:
				tx.AccessList = makeAccessList(t, accounts, keysPerAccount)
			case *types.DynamicFeeTx:
				tx.AccessList = makeAccessList(t, accounts, keysPerAccount)
			case *types.BlobTx:
				tx.AccessList = makeAccessList(t, accounts, keysPerAccount)
			case *types.SetCodeTx:
				tx.AccessList = makeAccessList(t, accounts, keysPerAccount)
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
		}
	}
	withAuthorizations := func(chainId *big.Int, accounts int) modificationFunction {
		makeAuthList := func(t *testing.T, chainId *big.Int, accounts int) []types.SetCodeAuthorization {
			authList := make([]types.SetCodeAuthorization, accounts)
			for i := range authList {
				account := NewAccount()

				auth, err := types.SignSetCode(account.PrivateKey,
					types.SetCodeAuthorization{
						ChainID: *uint256.MustFromBig(chainId),
						Address: common.BigToAddress(big.NewInt(int64(i))),
						Nonce:   0,
					})
				require.NoError(t, err)
				authList[i] = auth
			}
			return authList
		}

		return func(t *testing.T, tx *types.TxData) {
			switch tx := (*tx).(type) {
			case *types.LegacyTx:
				// ignore
			case *types.AccessListTx:
				// ignore
			case *types.DynamicFeeTx:
				// ignore
			case *types.BlobTx:
				// ignore
			case *types.SetCodeTx:
				tx.AuthList = makeAuthList(t, chainId, accounts)
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
		}
	}

	withoutTo := func() modificationFunction {
		return func(t *testing.T, tx *types.TxData) {
			switch tx := (*tx).(type) {
			case *types.LegacyTx:
				tx.To = nil
			case *types.AccessListTx:
				tx.To = nil
			case *types.DynamicFeeTx:
				tx.To = nil
			case *types.BlobTx, *types.SetCodeTx:
				// ignore without to
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
		}
	}

	withTo := func(address common.Address) modificationFunction {
		return func(t *testing.T, tx *types.TxData) {
			switch tx := (*tx).(type) {
			case *types.LegacyTx:
				tx.To = &address
			case *types.AccessListTx:
				tx.To = &address
			case *types.DynamicFeeTx:
				tx.To = &address
			case *types.BlobTx:
				tx.To = address
			case *types.SetCodeTx:
				tx.To = address
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
		}
	}

	t.Run("filled transactions can be executed", func(t *testing.T) {
		session := session.SpawnSession(t)
		t.Parallel()
		tests := GenerateTestDataBasedOnModificationCombinations(
			func() types.TxData { return nil },
			[][]modificationFunction{
				// Transaction type
				{
					transactionType(types.LegacyTxType),
					transactionType(types.AccessListTxType),
					transactionType(types.DynamicFeeTxType),
					transactionType(types.BlobTxType),
					transactionType(types.SetCodeTxType),
				},
				// Data
				{noData(), withData(100, 1)},
				// To
				{withoutTo(), withTo(NewAccount().Address())},
				// AccessList
				{noAccessList(), withAccessList(1, 1), withAccessList(8, 4)},
				// Authorizations (for transactions that require them, one minimum)
				{withAuthorizations(chainId, 1), withAuthorizations(chainId, 8)},
			},
			func(tc types.TxData, pieces []modificationFunction) types.TxData {
				for _, piece := range pieces {
					piece(t, &tc)
				}
				return tc
			},
		)

		nonce, err := client.NonceAt(t.Context(), session.GetSessionSponsor().Address(), nil)
		require.NoError(t, err)

		pending := []common.Hash{}

		for tx := range tests {
			switch tx := tx.(type) {
			case *types.LegacyTx:
				tx.Nonce = nonce
			case *types.AccessListTx:
				tx.Nonce = nonce
			case *types.DynamicFeeTx:
				tx.Nonce = nonce
			case *types.BlobTx:
				tx.Nonce = nonce
			case *types.SetCodeTx:
				tx.Nonce = nonce
			default:
				t.Fatalf("unexpected transaction type: %T", tx)
			}
			nonce++

			tx := CreateTransaction(t, session, tx, session.GetSessionSponsor())

			// the filled values suffice to get the transaction accepted and executed
			err := client.SendTransaction(t.Context(), tx)
			require.NoError(t, err)
			pending = append(pending, tx.Hash())
		}

		for _, txHash := range pending {
			receipt, err := session.GetReceipt(txHash)
			require.NoError(t, err)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		}
	})

	t.Run("zero nonce is defaulted", func(t *testing.T) {
		// this generation is tested isolated because the previous test case
		// utilizes manual nonce setting to issue multiple transactions asynchronously
		session := session.SpawnSession(t)
		t.Parallel()
		// account has a non-zero nonce
		receipt, err := session.EndowAccount(common.Address{}, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

		tx := CreateTransaction(t, session, &types.LegacyTx{}, session.GetSessionSponsor())

		nonce, err := client.NonceAt(t.Context(), session.GetSessionSponsor().Address(), nil)
		require.NoError(t, err)

		require.Equal(t, nonce, tx.Nonce())
	})

	t.Run("non-zero nonce is not defaulted", func(t *testing.T) {
		session := session.SpawnSession(t)
		t.Parallel()

		// endowments modify the account nonce
		var receipt *types.Receipt
		var err error
		for range 2 {
			receipt, err = session.EndowAccount(common.Address{}, big.NewInt(1))
			require.NoError(t, err)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		}

		err = waitUntilTransactionIsRetiredFromPoolByHash(t, client, receipt.TxHash, session.GetSessionSponsor().Address())
		require.NoError(t, err)

		tx := CreateTransaction(t, session, &types.LegacyTx{Nonce: 1}, session.GetSessionSponsor())

		// the filled values suffice to get the transaction accepted and executed
		_, err = session.Run(tx)
		require.ErrorContains(t, err, "nonce too low")
	})

	t.Run("non-zero gas is not defaulted ", func(t *testing.T) {
		session := session.SpawnSession(t)
		t.Parallel()

		tx := CreateTransaction(t, session, &types.LegacyTx{Gas: 1}, session.GetSessionSponsor())

		// the filled values suffice to get the transaction accepted and executed
		_, err := session.Run(tx)
		require.ErrorContains(t, err, " intrinsic gas too low")
	})

	t.Run("non-zero gas-price is not defaulted ", func(t *testing.T) {
		session := session.SpawnSession(t)
		t.Parallel()

		tx := CreateTransaction(t, session, &types.LegacyTx{GasPrice: big.NewInt(1)}, session.GetSessionSponsor())

		// the filled values suffice to get the transaction accepted and executed
		_, err := session.Run(tx)
		require.ErrorContains(t, err, "underpriced")
	})
}

func TestSetTransactionDefaults_IsCorrectAfterUpgradesChange(t *testing.T) {
	net := StartIntegrationTestNetWithJsonGenesis(t)

	client, err := net.GetClient()
	require.NoError(t, err)
	defer client.Close()

	sender := MakeAccountWithBalance(t, net, big.NewInt(1e18))

	tx := SignTransaction(t, net.GetChainId(),
		SetTransactionDefaults(
			t, net,
			&types.LegacyTx{
				To:    &common.Address{0x42},
				Value: big.NewInt(1),
				// large data buffer, starting with an STOP opcode
				Data: []byte{0x0, 40_000: 0xff},
			},
			sender),
		sender)

	receipt, err := net.Run(tx)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	type rulesType struct {
		Upgrades struct{ Allegro bool }
	}
	rulesDiff := rulesType{
		Upgrades: struct{ Allegro bool }{Allegro: true},
	}
	UpdateNetworkRules(t, net, rulesDiff)
	AdvanceEpochAndWaitForBlocks(t, net)

	// Wait until tx pool updates
	tx2 := SignTransaction(t, net.GetChainId(),
		SetTransactionDefaults(
			t, net,
			&types.LegacyTx{
				To:    &common.Address{0x42},
				Value: big.NewInt(1),
				Nonce: 1,
				// large data buffer, starting with an STOP opcode
				Data: []byte{0x0, 40_000: 0xff},
			},
			sender),
		sender)
	receipt2, err := net.Run(tx2)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt2.Status)

	// This test relies on the fact that Allegro introduces extra gas for large data buffers.
	require.Greater(t, receipt2.GasUsed, receipt.GasUsed)
}

func TestWaitUntilTransactionIsRetiredFromPool_waitsFromCompletion(t *testing.T) {
	session := getIntegrationTestNetSession(t, opera.GetAllegroUpgrades())
	t.Parallel()

	// TODO: this test can benefit from using synctest once it is available
	// This test expects some large timers to time-out

	client, err := session.GetClient()
	require.NoError(t, err)
	defer client.Close()

	account := MakeAccountWithBalance(t, session, big.NewInt(1e18))

	chainId, err := client.ChainID(t.Context())
	require.NoError(t, err)

	txData := SetTransactionDefaults(t, session, &types.LegacyTx{}, account)
	txData.Nonce = 1
	txInvalidNonce := SignTransaction(t, chainId, txData, account)

	err = client.SendTransaction(t.Context(), txInvalidNonce)
	require.NoError(t, err)

	// Because nonce is set to current nonce + 1, the transaction will not be executed
	// waiting must time out
	err = WaitUntilTransactionIsRetiredFromPool(t, client, txInvalidNonce)
	require.ErrorContains(t, err, "wait timeout")

	txData.Nonce = 0
	txCorrectNonce := SignTransaction(t, chainId, txData, account)

	err = client.SendTransaction(t.Context(), txCorrectNonce)
	require.NoError(t, err)

	// Once the valid nonce transaction is sent, both transactions will be executed
	// and retired from the pool
	err = WaitUntilTransactionIsRetiredFromPool(t, client, txInvalidNonce)
	require.NoError(t, err)
	err = WaitUntilTransactionIsRetiredFromPool(t, client, txCorrectNonce)
	require.NoError(t, err)
}

func TestWaitFor_SucceedsWhenPredicateIsSatisfied(t *testing.T) {
	t.Parallel()

	// right away
	err := WaitFor(t.Context(), func(ctx context.Context) (bool, error) {
		return true, nil
	})
	require.NoError(t, err, "WaitFor should not error")

	// after some tries
	count := 0
	err = WaitFor(t.Context(), func(ctx context.Context) (bool, error) {
		count++
		return count == 10, nil
	})
	require.NoError(t, err, "WaitFor should not error")
}

func TestWaitFor_EventuallyTimesOut(t *testing.T) {
	t.Parallel()
	// TODO: once golang 1.25 is released, this is a candidate to run using
	// synctest
	// For the time being, this parallel routine takes ~100 seconds, which is
	// less than the longest running integration test. Therefore it should not
	// affect the total execution time.
	err := WaitFor(t.Context(), func(ctx context.Context) (bool, error) {
		return false, nil
	})
	require.ErrorContains(t, err, "wait timeout")
}

func TestWaitFor_ForwardsErrors(t *testing.T) {
	t.Parallel()

	// right away
	err := WaitFor(t.Context(), func(ctx context.Context) (bool, error) {
		return false, fmt.Errorf("some error")
	})
	require.ErrorContains(t, err, "some error")

	// after some tries
	count := 0
	err = WaitFor(t.Context(), func(ctx context.Context) (bool, error) {
		count++
		return count == 10, fmt.Errorf("some error")
	})
	require.ErrorContains(t, err, "some error")
}
