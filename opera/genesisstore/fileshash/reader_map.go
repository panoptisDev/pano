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

package fileshash

import (
	"io"

	"github.com/Fantom-foundation/lachesis-base/hash"
)

func Wrap(backend func(string) (io.Reader, error), maxMemoryUsage uint64, roots map[string]hash.Hash) func(string) (io.Reader, error) {
	return func(name string) (io.Reader, error) {
		root, ok := roots[name]
		if !ok {
			return nil, ErrRootNotFound
		}
		f, err := backend(name)
		if err != nil {
			return nil, err
		}
		return WrapReader(f, maxMemoryUsage, root), nil
	}
}
