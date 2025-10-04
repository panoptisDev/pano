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

package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	testData := map[string]int{}
	curVer := &inmemIDStore{}

	t.Run("native", func(t *testing.T) {
		require := require.New(t)

		native := Begin("native")
		curVer.lastID = idOf("native")

		num := 1
		lastGood := native.Next("01",
			func() error {
				testData["migration1"] = num
				num++
				return nil
			},
		).Next("02",
			func() error {
				testData["migration2"] = num
				num++
				return nil
			},
		)

		afterBad := lastGood.Next("03",
			func() error {
				testData["migration3"] = num
				num++
				return errors.New("test migration error")
			},
		).Next("04",
			func() error {
				testData["migration4"] = num
				num++
				return nil
			},
		)

		err := afterBad.Exec(curVer, flush)
		require.Error(err, "Success after a migration error")

		lastID := curVer.GetID()
		require.Equal(lastGood.ID(), lastID, "Bad last id in idProducer after a migration error")

		require.Equal(1, testData["migration1"], "Bad value after run migration1")
		require.Equal(2, testData["migration2"], "Bad value after run migration2")
		require.Equal(3, testData["migration3"], "Bad value after run migration3")
		require.Empty(testData["migration4"], "Bad data for migration4 - should by empty")

		// Continue with fixed transactions

		num = 3
		fixed := lastGood.Next("03",
			func() error {
				testData["migration3"] = num
				num++
				return nil
			},
		).Next("04",
			func() error {
				testData["migration4"] = num
				num++
				return nil
			},
		)

		err = fixed.Exec(curVer, flush)
		require.NoError(err, "Error when run migration manager")

		require.Equal(1, testData["migration1"], "Bad value after run migration1")
		require.Equal(2, testData["migration2"], "Bad value after run migration2")
		require.Equal(3, testData["migration3"], "Bad value after run migration3")
		require.Equal(4, testData["migration4"], "Bad value after run migration4")
	})

	t.Run("nonnative", func(t *testing.T) {
		require := require.New(t)

		nonnative := Begin("nonnative").
			Next("01",
				func() error {
					testData["migration1"] = 999
					return nil
				},
			)

		err := nonnative.Exec(curVer, flush)
		require.NotEqual(999, testData["migration1"], "nonnative migration is applied")
		require.Error(err)
	})
}

func flush() error {
	return nil
}
