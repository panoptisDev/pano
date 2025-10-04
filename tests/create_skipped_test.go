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
	"slices"
	"testing"

	"github.com/panoptisDev/pano/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestAccountCreation_CreateCallsWithInitCodesTooLargeDoNotAlterBalance(t *testing.T) {
	versions := map[string]opera.Upgrades{
		"pano":   opera.GetPanoUpgrades(),
		"allegro": opera.GetAllegroUpgrades(),
	}

	for name, version := range versions {
		t.Run(name, func(t *testing.T) {
			net := StartIntegrationTestNetWithJsonGenesis(t, IntegrationTestNetOptions{
				Upgrades: &version,
				ClientExtraArguments: []string{
					"--disable-txPool-validation",
				},
			})

			client, err := net.GetClient()
			require.NoError(t, err)
			defer client.Close()

			sender := MakeAccountWithBalance(t, net, big.NewInt(1e18))

			gasPrice, err := client.SuggestGasPrice(t.Context())
			require.NoError(t, err)

			chainId, err := client.ChainID(t.Context())
			require.NoError(t, err)

			initCode := make([]byte, 50000)
			txData := &types.LegacyTx{
				Nonce:    0,
				Gas:      10000000,
				GasPrice: gasPrice,
				To:       nil, // address 0x00 for contract creation
				Value:    big.NewInt(0),
				Data:     initCode,
			}
			tx := SignTransaction(t, chainId, txData, sender)

			// Check balance before sending the transaction
			preBalance, err := client.BalanceAt(t.Context(), sender.Address(), nil)
			require.NoError(t, err)

			// Send the transaction
			err = client.SendTransaction(t.Context(), tx)
			require.NoError(t, err)

			// Send another simple transaction to ensure a block is created
			receipt, err := net.EndowAccount(common.Address{0x42}, big.NewInt(42))
			require.NoError(t, err)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

			// Check that the simple transaction was included in the block
			blockTransaction, err := client.BlockByNumber(t.Context(), receipt.BlockNumber)
			require.NoError(t, err)
			contains := slices.ContainsFunc(blockTransaction.Transactions(), func(tx *types.Transaction) bool {
				return tx.Hash() == receipt.TxHash
			})
			require.True(t, contains, "transaction should be included in the block")

			// Check that no other transactions were included in the block
			require.Len(t, blockTransaction.Transactions(), 1, "block should contain exactly one transaction")

			// Check balance after sending the transaction
			postBalance, err := client.BalanceAt(t.Context(), sender.Address(), receipt.BlockNumber)
			require.NoError(t, err)

			if version == opera.GetPanoUpgrades() {
				require.Less(t, postBalance.Uint64(), preBalance.Uint64(), "balance should decrease after failed contract creation")
			}
			if version == opera.GetAllegroUpgrades() {
				require.Equal(t, preBalance, postBalance, "balance should not change after failed contract creation")
			}
		})
	}
}

func TestAccountCreation_CreateCallsProducingCodesTooLargeProduceAUnsuccessfulReceipt(t *testing.T) {
	codeSize := uint256.NewInt(25000).Bytes32()
	initCode := []byte{byte(vm.PUSH32)}
	initCode = append(initCode, codeSize[:]...)
	initCode = append(initCode, []byte{
		byte(vm.PUSH1), byte(0),
		byte(vm.RETURN),
	}...)

	session := getIntegrationTestNetSession(t, opera.GetPanoUpgrades())
	t.Parallel()

	client, err := session.GetClient()
	require.NoError(t, err)
	defer client.Close()

	sender := MakeAccountWithBalance(t, session, big.NewInt(1e18))

	gasPrice, err := client.SuggestGasPrice(t.Context())
	require.NoError(t, err)

	chainId, err := client.ChainID(t.Context())
	require.NoError(t, err)

	txData := &types.LegacyTx{
		Nonce:    0,
		Gas:      100000,
		GasPrice: gasPrice,
		To:       nil, // address 0x00 for contract creation
		Value:    big.NewInt(0),
		Data:     initCode,
	}
	tx := SignTransaction(t, chainId, txData, sender)

	receipt, err := session.Run(tx)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusFailed, receipt.Status)
}
