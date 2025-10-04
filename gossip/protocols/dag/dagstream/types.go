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

package dagstream

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/panoptisDev/lachesis-base/gossip/basestream"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/dag"
)

type Request struct {
	Session   Session
	Limit     dag.Metric
	Type      basestream.RequestType
	MaxChunks uint32
}

type Response struct {
	SessionID uint32
	Done      bool
	IDs       hash.Events
	Events    []rlp.RawValue
}

type Session struct {
	ID    uint32
	Start Locator
	Stop  Locator
}

type Locator []byte

func (l Locator) Compare(b basestream.Locator) int {
	return bytes.Compare(l, b.(Locator))
}

func (l Locator) Inc() basestream.Locator {
	nextBn := new(big.Int).SetBytes(l)
	nextBn.Add(nextBn, common.Big1)
	return Locator(common.LeftPadBytes(nextBn.Bytes(), len(l)))
}

type Payload struct {
	IDs    hash.Events
	Events []rlp.RawValue
	Size   uint64
}

func (p *Payload) AddEvent(id hash.Event, eventB rlp.RawValue) {
	p.IDs = append(p.IDs, id)
	p.Events = append(p.Events, eventB)
	p.Size += uint64(len(eventB))
}

func (p *Payload) AddID(id hash.Event, size int) {
	p.IDs = append(p.IDs, id)
	p.Size += uint64(size)
}

func (p Payload) Len() int {
	return len(p.IDs)
}

func (p Payload) TotalSize() uint64 {
	return p.Size
}

func (p Payload) TotalMemSize() int {
	if len(p.Events) != 0 {
		return int(p.Size) + len(p.IDs)*128
	}
	return len(p.IDs) * 128
}

const (
	RequestIDs    basestream.RequestType = 0
	RequestEvents basestream.RequestType = 2
)
