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

package logger

import (
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	Log log.Logger
}

func New(name ...string) Instance {
	if len(name) == 0 {
		return Instance{
			Log: log.New(),
		}
	}
	return Instance{
		Log: log.New("module", name[0]),
	}
}
