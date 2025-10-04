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

package autocompact

import (
	"bytes"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type ContainerI interface {
	Add(key []byte, size uint64)
	Merge(c ContainerI)
	Error() error
	Reset()
	Size() uint64
	Ranges() []Range
}

type Range struct {
	minKey []byte
	maxKey []byte
}

// MonotonicContainer implements tracking of compaction ranges in cases when keys are inserted as series of monotonic ranges
type MonotonicContainer struct {
	forward bool
	ranges  []Range
	size    uint64
	err     error
}

type DevnullContainer struct{}

func (d DevnullContainer) Add(key []byte, size uint64) {}

func (d DevnullContainer) Merge(c ContainerI) {}

func (d DevnullContainer) Error() error {
	return nil
}

func (d DevnullContainer) Reset() {

}

func (d DevnullContainer) Size() uint64 {
	return 0
}

func (d DevnullContainer) Ranges() []Range {
	return []Range{}
}

func NewForwardCont() ContainerI {
	return &MonotonicContainer{
		forward: true,
	}
}

func NewBackwardsCont() ContainerI {
	return &MonotonicContainer{
		forward: false,
	}
}

func NewDevnullCont() ContainerI {
	return DevnullContainer{}
}

func (m *MonotonicContainer) addRange(key []byte) {
	m.ranges = append(m.ranges, Range{
		minKey: common.CopyBytes(key),
		maxKey: common.CopyBytes(key),
	})
}

func (m *MonotonicContainer) Add(key []byte, size uint64) {
	m.size += size
	if len(m.ranges) == 0 {
		m.addRange(key)
	}
	// extend the last range if it's a monotonic addition or start new range otherwise
	l := len(m.ranges) - 1
	if m.forward {
		if bytes.Compare(key, m.ranges[l].maxKey) >= 0 {
			m.ranges[l].maxKey = common.CopyBytes(key)
		} else {
			m.addRange(key)
		}
	} else {
		if bytes.Compare(key, m.ranges[l].minKey) <= 0 {
			m.ranges[l].minKey = common.CopyBytes(key)
		} else {
			m.addRange(key)
		}
	}
}

func (m *MonotonicContainer) Merge(c ContainerI) {
	if err := c.Error(); err != nil {
		m.err = err
	}

	for _, r := range c.Ranges() {
		if m.forward {
			m.Add(r.minKey, 0)
			m.Add(r.maxKey, 0)
		} else {
			m.Add(r.maxKey, 0)
			m.Add(r.minKey, 0)
		}
	}
	m.size += c.Size()
}

func (m *MonotonicContainer) Error() error {
	if m.err != nil {
		return m.err
	}
	if len(m.ranges) > 2 {
		return errors.New("too many compaction ranges, it's likely that dataset isn't monotonic enough")
	}
	return nil
}

func (m *MonotonicContainer) Reset() {
	m.ranges = nil
	m.size = 0
}

func (m *MonotonicContainer) Size() uint64 {
	return m.size
}

func (m *MonotonicContainer) Ranges() []Range {
	return m.ranges
}
