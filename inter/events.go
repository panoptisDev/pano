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

package inter

import (
	"bytes"

	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/dag"
)

// Events is a ordered slice of events.
type Events []*Event

// String returns human readable representation.
func (ee Events) String() string {
	return ee.Bases().String()
}

// Add appends hash to the slice.
func (ee *Events) Add(e ...*Event) {
	*ee = append(*ee, e...)
}

func (ee Events) IDs() hash.Events {
	res := make(hash.Events, 0, len(ee))
	for _, e := range ee {
		res.Add(e.ID())
	}
	return res
}

func (ee Events) Bases() dag.Events {
	res := make(dag.Events, 0, ee.Len())
	for _, e := range ee {
		res = append(res, e)
	}
	return res
}

func (ee Events) Interfaces() EventIs {
	res := make(EventIs, 0, ee.Len())
	for _, e := range ee {
		res = append(res, e)
	}
	return res
}

func (ee Events) Len() int      { return len(ee) }
func (ee Events) Swap(i, j int) { ee[i], ee[j] = ee[j], ee[i] }
func (ee Events) Less(i, j int) bool {
	return bytes.Compare(ee[i].ID().Bytes(), ee[j].ID().Bytes()) < 0
}

// EventPayloads is a ordered slice of EventPayload.
type EventPayloads []*EventPayload

// String returns human readable representation.
func (ee EventPayloads) String() string {
	return ee.Bases().String()
}

// Add appends hash to the slice.
func (ee *EventPayloads) Add(e ...*EventPayload) {
	*ee = append(*ee, e...)
}

func (ee EventPayloads) IDs() hash.Events {
	res := make(hash.Events, 0, len(ee))
	for _, e := range ee {
		res.Add(e.ID())
	}
	return res
}

func (ee EventPayloads) Bases() dag.Events {
	res := make(dag.Events, 0, ee.Len())
	for _, e := range ee {
		res = append(res, e)
	}
	return res
}

func (ee EventPayloads) Len() int      { return len(ee) }
func (ee EventPayloads) Swap(i, j int) { ee[i], ee[j] = ee[j], ee[i] }
func (ee EventPayloads) Less(i, j int) bool {
	return bytes.Compare(ee[i].ID().Bytes(), ee[j].ID().Bytes()) < 0
}

// EventIs is a ordered slice of events.
type EventIs []EventI

// String returns human readable representation.
func (ee EventIs) String() string {
	return ee.Bases().String()
}

// Add appends hash to the slice.
func (ee *EventIs) Add(e ...EventI) {
	*ee = append(*ee, e...)
}

func (ee EventIs) IDs() hash.Events {
	res := make(hash.Events, 0, len(ee))
	for _, e := range ee {
		res.Add(e.ID())
	}
	return res
}

func (ee EventIs) Bases() dag.Events {
	res := make(dag.Events, 0, ee.Len())
	for _, e := range ee {
		res = append(res, e)
	}
	return res
}

func (ee EventIs) Len() int      { return len(ee) }
func (ee EventIs) Swap(i, j int) { ee[i], ee[j] = ee[j], ee[i] }
func (ee EventIs) Less(i, j int) bool {
	return bytes.Compare(ee[i].ID().Bytes(), ee[j].ID().Bytes()) < 0
}
