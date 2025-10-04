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

package subsidies

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/panoptisDev/pano/gossip/blockproc/subsidies/registry"
	"github.com/panoptisDev/pano/opera"
	"github.com/panoptisDev/pano/utils/signers/internaltx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

//go:generate mockgen -source=subsidies.go -destination=subsidies_mock.go -package=subsidies

// IsSponsorshipRequest checks if a transaction is requesting sponsorship from
// a pre-allocated sponsorship pool. A sponsorship request is defined as a
// transaction with a maximum gas price of zero.
func IsSponsorshipRequest(tx *types.Transaction) bool {
	return tx != nil &&
		!internaltx.IsInternal(tx) &&
		tx.To() != nil &&
		tx.GasPrice().Sign() == 0
}

// FundId is an identifier for a fund in the subsidies registry contract.
type FundId [32]byte

func (id FundId) String() string {
	return fmt.Sprintf("0x%x", id[:])
}

// IsCovered checks if the given transaction is covered by available subsidies.
// If preconditions are met, it queries the subsidies registry contract. If
// there are sufficient funds, it returns true, otherwise false.
func IsCovered(
	upgrades opera.Upgrades,
	vm VirtualMachine,
	reader SenderReader,
	tx *types.Transaction,
	baseFee *big.Int,
) (bool, FundId, GasConfig, error) {
	if !upgrades.GasSubsidies {
		return false, FundId{}, GasConfig{}, nil
	}
	if !IsSponsorshipRequest(tx) {
		return false, FundId{}, GasConfig{}, nil
	}

	// Derive the sender of the transaction before interacting with the EVM.
	sender, err := reader.Sender(tx)
	if err != nil {
		return false, FundId{}, GasConfig{}, fmt.Errorf("failed to derive sender: %w", err)
	}

	// Fetch the current configuration from the subsidies registry.
	gasConfig, err := getGasConfig(vm)
	if err != nil {
		return false, FundId{}, GasConfig{}, fmt.Errorf("failed to get gas config: %w", err)
	}

	// Build the choose-fund query call to the subsidies registry contract.
	caller := common.Address{}
	target := registry.GetAddress()

	// Build the input data for the IsCovered call.
	maxGas := tx.Gas() + gasConfig.overheadToCharge // < maximum what is charged for
	maxFee := new(big.Int).Mul(baseFee, new(big.Int).SetUint64(maxGas))
	input, err := createChooseFundInput(sender, tx, maxFee)
	if err != nil {
		return false, FundId{}, GasConfig{}, fmt.Errorf("failed to create input for subsidies registry call: %w", err)
	}

	// Run the query on the EVM and the provided state.
	result, _, err := vm.Call(caller, target, input, gasConfig.chooseFundGasLimit, uint256.NewInt(0))
	if err != nil {
		return false, FundId{}, GasConfig{}, fmt.Errorf("EVM call failed: %w", err)
	}

	// An empty result indicates that there is no contract installed.
	if len(result) == 0 {
		return false, FundId{}, GasConfig{}, fmt.Errorf("subsidies registry contract not found")
	}

	// Parse the result of the call.
	covered, fundID, err := parseChooseFundResult(result)
	if err != nil {
		return false, FundId{}, GasConfig{}, fmt.Errorf("failed to parse result of subsidies registry call: %w", err)
	}
	return covered, fundID, GasConfig{
		DeductFeesGasCost:          gasConfig.deductFeesGasLimit,
		SponsorshipOverheadGasCost: gasConfig.overheadToCharge,
	}, nil
}

type GasConfig struct {
	SponsorshipOverheadGasCost uint64
	DeductFeesGasCost          uint64
}

// VirtualMachine is a minimal interface for an EVM instance that can be used
// to query the subsidies registry contract.
type VirtualMachine interface {
	Call(
		from common.Address,
		to common.Address,
		input []byte,
		gas uint64,
		value *uint256.Int,
	) (
		result []byte,
		gasLeft uint64,
		err error,
	)
}

// GetFeeChargeTransaction builds a transaction that deducts the given fee
// amount from the sponsorship pool of the given subsidies registry contract.
// The returned transaction is unsigned and has zero value and gas price. It is
// intended to be introduced by the block processor after the sponsored
// transaction has been executed.
func GetFeeChargeTransaction(
	nonceSource NonceSource,
	fundId FundId,
	config GasConfig,
	gasUsed uint64,
	gasPrice *big.Int,
) (*types.Transaction, error) {
	sender := common.Address{}
	nonce := nonceSource.GetNonce(sender)

	// Calculate the fee to be deducted: (gasUsed + overhead) * gasPrice
	fee, overflow := uint256.FromBig(new(big.Int).Mul(
		new(big.Int).Add(
			new(big.Int).SetUint64(gasUsed),
			new(big.Int).SetUint64(config.SponsorshipOverheadGasCost),
		),
		gasPrice,
	))
	if overflow {
		return nil, fmt.Errorf("fee calculation overflow")
	}

	input := createDeductFeesInput(fundId, *fee)
	return types.NewTransaction(
		nonce, registry.GetAddress(), common.Big0,
		config.DeductFeesGasCost, common.Big0, input,
	), nil
}

// NonceSource provides nonces for addresses. It is used to determine the
// correct nonce for the fee deduction transaction.
type NonceSource interface {
	GetNonce(addr common.Address) uint64
}

// SenderReader is an interface for types that can extract the sender
// address from a transaction. Typically, this is an implementation of
// types.Signer.
type SenderReader interface {
	Sender(*types.Transaction) (common.Address, error)
}

// --- utility functions ---

// getGasConfig queries the subsidies registry contract for the current gas
// configuration. It returns the gas limits to be used when calling the
// `chooseFund` and `deductFees` functions, as well as the overhead to charge
// for sponsored transactions.
func getGasConfig(
	vm VirtualMachine,
) (gasConfig, error) {
	// Call the getGasConfig function on the subsidies registry contract, which
	// takes no arguments and returns three uint64 values.
	caller := common.Address{}
	target := registry.GetAddress()

	// Build the input data for the IsCovered call.
	input := make([]byte, 4) // function selector only
	binary.BigEndian.PutUint32(input, registry.GetGasConfigFunctionSelector)

	// Run the query on the EVM and the provided state.
	const initialGas = registry.GasLimitForGetGasConfig
	result, _, err := vm.Call(caller, target, input, initialGas, uint256.NewInt(0))
	if err != nil {
		return gasConfig{}, fmt.Errorf("EVM call failed: %w", err)
	}

	// An empty result indicates that there is no contract installed.
	if len(result) == 0 {
		return gasConfig{}, fmt.Errorf("subsidies registry contract not found")
	}

	if len(result) != 3*32 {
		return gasConfig{}, fmt.Errorf("invalid result length from getGasConfig call, wanted %d, got %d", 3*32, len(result))
	}

	// check for uint64 overflows
	type bytes24 [24]byte
	zero := bytes24{}
	if bytes24(result[0:32-8]) != zero ||
		bytes24(result[32:64-8]) != zero ||
		bytes24(result[64:96-8]) != zero {
		return gasConfig{}, fmt.Errorf("invalid result from getGasConfig call, values do not fit into uint64")
	}

	chooseFundGasLimit := binary.BigEndian.Uint64(result[32-8 : 32])
	deductFeesGasLimit := binary.BigEndian.Uint64(result[64-8 : 64])
	overheadToCharge := binary.BigEndian.Uint64(result[96-8 : 96])
	return gasConfig{
		chooseFundGasLimit: chooseFundGasLimit,
		deductFeesGasLimit: deductFeesGasLimit,
		overheadToCharge:   overheadToCharge,
	}, nil
}

type gasConfig struct {
	chooseFundGasLimit uint64
	deductFeesGasLimit uint64
	overheadToCharge   uint64
}

// createChooseFundInput creates the input data for the chooseFund call to the
// subsidies registry contract.
func createChooseFundInput(
	sender common.Address,
	tx *types.Transaction,
	fee *big.Int,
) ([]byte, error) {
	if tx == nil || fee == nil {
		return nil, fmt.Errorf("invalid transaction or fee")
	}
	if fee.BitLen() > 256 {
		return nil, fmt.Errorf("fee does not fit into 32 bytes")
	}

	to := common.Address{}
	if tx.To() != nil {
		to = *tx.To()
	}

	// Add the function selector for `isCovered`.
	input := []byte{}
	input = binary.BigEndian.AppendUint32(input, registry.ChooseFundFunctionSelector)

	// The from and to addresses are padded to 32 bytes.
	addressPadding := [12]byte{}
	input = append(input, addressPadding[:]...)
	input = append(input, sender[:]...)
	input = append(input, addressPadding[:]...)
	input = append(input, to[:]...)

	// The value is padded to 32 bytes.
	input = append(input, tx.Value().FillBytes(make([]byte, 32))...)

	// The nonce is padded to 32 bytes.
	uint64Padding := [24]byte{}
	input = append(input, uint64Padding[:]...)
	input = binary.BigEndian.AppendUint64(input, tx.Nonce())

	// The calldata is a dynamic parameter, encoded as its offset in the input
	// data. Dynamic sized parameters are at the end of the input data.
	input = append(input, uint64Padding[:]...)
	input = binary.BigEndian.AppendUint64(input, 32*6) // 6 32-byte parameters

	// The fee is padded to 32 bytes.
	input = append(input, fee.FillBytes(make([]byte, 32))...)

	// -- dynamic sized parameters --

	// The input data is prefixed by its length as a 32-byte value,
	// followed by the actual data, padded to a multiple of 32 bytes.
	input = append(input, uint64Padding[:]...)
	input = binary.BigEndian.AppendUint64(input, uint64(len(tx.Data())))
	input = append(input, tx.Data()...)
	if len(tx.Data())%32 != 0 {
		dataPadding := make([]byte, 32-len(tx.Data())%32)
		input = append(input, dataPadding...)
	}

	return input, nil
}

// parseChooseFundResult parses the result of the IsCovered call to the
// subsidies registry contract.
func parseChooseFundResult(data []byte) (covered bool, fundID FundId, err error) {
	// The result is a 32-byte long FundId.
	if len(data) != 32 {
		return false, FundId{}, fmt.Errorf("invalid result length from chooseFund call: %d", len(data))
	}
	fundId := FundId(data[0:32])
	return fundId != (FundId{}), fundId, nil
}

// createDeductFeesInput creates the input data for the DeductFees call to the
// subsidies registry contract.
func createDeductFeesInput(fundId FundId, fee uint256.Int) []byte {
	// Signature: deductFees(bytes32 fundId, uint256 fee)
	input := make([]byte, 4+2*32) // function selector + 2 parameters

	binary.BigEndian.PutUint32(input, registry.DeductFeesFunctionSelector)
	copy(input[4:36], fundId[:])
	fee.WriteToArray32((*[32]byte)(input[36:68]))
	return input
}
