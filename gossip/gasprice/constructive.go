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

package gasprice

import (
	"math/big"

	"github.com/panoptisDev/lachesis-base/utils/piecefunc"
)

func (gpo *Oracle) maxTotalGasPower() *big.Int {
	rules := gpo.backend.GetRules()

	allocBn := new(big.Int).SetUint64(rules.Economy.LongGasPower.AllocPerSec)
	periodBn := new(big.Int).SetUint64(uint64(rules.Economy.LongGasPower.MaxAllocPeriod))
	maxTotalGasPowerBn := new(big.Int).Mul(allocBn, periodBn)
	maxTotalGasPowerBn.Div(maxTotalGasPowerBn, secondBn)
	return maxTotalGasPowerBn
}

func (gpo *Oracle) constructiveGasPrice(gasOffestAbs uint64, gasOffestRatio uint64, adjustedMinPrice *big.Int) *big.Int {
	max := gpo.maxTotalGasPower()

	current64 := gpo.backend.TotalGasPowerLeft()
	if current64 > gasOffestAbs {
		current64 -= gasOffestAbs
	} else {
		current64 = 0
	}
	current := new(big.Int).SetUint64(current64)

	freeRatioBn := current.Mul(current, DecimalUnitBn)
	freeRatioBn.Div(freeRatioBn, max)
	freeRatio := freeRatioBn.Uint64()
	if freeRatio > gasOffestRatio {
		freeRatio -= gasOffestRatio
	} else {
		freeRatio = 0
	}
	if freeRatio > DecimalUnit {
		freeRatio = DecimalUnit
	}
	v := gpo.constructiveGasPriceOf(freeRatio, adjustedMinPrice)
	return v
}

var freeRatioToConstructiveGasRatio = piecefunc.NewFunc([]piecefunc.Dot{
	{
		X: 0,
		Y: 25 * DecimalUnit,
	},
	{
		X: 0.3 * DecimalUnit,
		Y: 9 * DecimalUnit,
	},
	{
		X: 0.5 * DecimalUnit,
		Y: 3.75 * DecimalUnit,
	},
	{
		X: 0.8 * DecimalUnit,
		Y: 1.5 * DecimalUnit,
	},
	{
		X: 0.95 * DecimalUnit,
		Y: 1.05 * DecimalUnit,
	},
	{
		X: DecimalUnit,
		Y: DecimalUnit,
	},
})

func (gpo *Oracle) constructiveGasPriceOf(freeRatio uint64, adjustedMinPrice *big.Int) *big.Int {
	multiplier := new(big.Int).SetUint64(freeRatioToConstructiveGasRatio(freeRatio))

	// gas price = multiplier * adjustedMinPrice
	price := multiplier.Mul(multiplier, adjustedMinPrice)
	return price.Div(price, DecimalUnitBn)
}
