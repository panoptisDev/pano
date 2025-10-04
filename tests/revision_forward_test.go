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

	"github.com/panoptisDev/pano/integration/makefakegenesis"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/tosca/go/tosca/vm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestTransaction_DelegationDesignationAddressAccessIsConsideredInAllegro(t *testing.T) {
	gas := uint64(21_000) // transaction base
	gas += 7 * 3          // 7 push instructions
	gas += 2_600          // cold access to recipient
	gas += 10             // gas in recursive call (is fully consumed due to failed execution)

	tests := map[string]struct {
		upgrades opera.Upgrades
		gas      uint64
	}{
		"Pano": {
			upgrades: opera.GetPanoUpgrades(),
			gas:      gas, // delegate designator ignored, no address access.
		},
		"Allegro": {
			upgrades: opera.GetAllegroUpgrades(),
			gas:      gas + 2_600, // cold access to delegate billed in interpreter.
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			net := StartIntegrationTestNetWithJsonGenesis(t, IntegrationTestNetOptions{
				Upgrades: &test.upgrades,
				Accounts: accountsToDeploy(),
			})

			client, err := net.GetClient()
			require.NoError(t, err)
			defer client.Close()

			sender := MakeAccountWithBalance(t, net, big.NewInt(1e18))

			gasPrice, err := client.SuggestGasPrice(t.Context())
			require.NoError(t, err)

			chainId, err := client.ChainID(t.Context())
			require.NoError(t, err)

			recipient := common.HexToAddress("0x44")
			txData := &types.AccessListTx{
				ChainID:    chainId,
				Nonce:      0,
				GasPrice:   gasPrice,
				Gas:        test.gas + 1, // +1 to ensure there was no error which consumed the gas
				To:         &recipient,
				Value:      big.NewInt(0),
				Data:       []byte{},
				AccessList: types.AccessList{},
			}
			tx := SignTransaction(t, chainId, txData, sender)

			receipt, err := net.Run(tx)
			require.NoError(t, err)

			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			require.Equal(t, test.gas, receipt.GasUsed)
		})
	}
}

func accountsToDeploy() []makefakegenesis.Account {
	// account 0x42 code: single invalid instruction (0xee)
	// account 0x43 code: delegation designation to 0x42: 0xef0100...042
	// account 0x44 code: code that calls 0x43

	account42 := makefakegenesis.Account{
		Name:    "account42",
		Address: common.HexToAddress("0x42"),
		Code:    []byte{0xee},
		Nonce:   1,
	}

	account43 := makefakegenesis.Account{
		Name:    "account43",
		Address: common.HexToAddress("0x43"),
		Code:    append([]byte{0xef, 0x01, 0x00}, common.HexToAddress("0x42").Bytes()...),
		Nonce:   1,
	}

	code44 := []byte{
		byte(vm.PUSH1), 0x00, // retSize
		byte(vm.PUSH1), 0x00, // retOffset
		byte(vm.PUSH1), 0x00, // argSize
		byte(vm.PUSH1), 0x00, // argOffset
		byte(vm.PUSH1), 0x00, // value
		byte(vm.PUSH20),
	}
	code44 = append(code44, common.HexToAddress("0x43").Bytes()...) // address
	code44 = append(code44,
		byte(vm.PUSH1), 0x0a, // gas
		byte(vm.CALL), // call
		byte(vm.STOP), // return
	)

	account44 := makefakegenesis.Account{
		Name:    "account44",
		Address: common.HexToAddress("0x44"),
		Code:    code44,
		Nonce:   1,
	}

	return []makefakegenesis.Account{account42, account43, account44}
}
