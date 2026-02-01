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

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnotateIfError_PropagatesNil(t *testing.T) {
	if AnnotateIfError(nil, "message") != nil {
		t.Error("AnnotateIfError should return nil when err is nil")
	}
}

func TestAnnotateIfError_AddsContextToError(t *testing.T) {
	err := fmt.Errorf("someError")
	errWithContext := AnnotateIfError(err, "message")
	require.ErrorContains(t, errWithContext, "message: someError")
}
