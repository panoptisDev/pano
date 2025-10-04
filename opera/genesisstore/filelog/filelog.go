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

package filelog

import (
	"fmt"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/panoptisDev/pano/utils"
)

type Filelog struct {
	io.Reader
	name     string
	size     uint64
	period   time.Duration
	consumed uint64
	prevLog  time.Time
	start    time.Time
}

func (f *Filelog) Read(p []byte) (n int, err error) {
	n, err = f.Reader.Read(p)
	f.consumed += uint64(n)
	if f.prevLog.IsZero() {
		log.Info(fmt.Sprintf("- Reading %s", f.name))
		f.prevLog = time.Now()
		f.start = time.Now()
	} else if f.consumed > 0 && f.consumed < f.size && time.Since(f.prevLog) >= f.period {
		elapsed := time.Since(f.start)
		eta := float64(f.size-f.consumed) / float64(f.consumed) * float64(elapsed)
		progress := float64(f.consumed) / float64(f.size)
		eta *= 1.0 + (1.0-progress)/2.0 // show slightly higher ETA as performance degrades over larger volumes of data
		progressStr := fmt.Sprintf("%.2f%%", 100*progress)
		log.Info(fmt.Sprintf("- Reading %s", f.name), "progress", progressStr, "elapsed", utils.PrettyDuration(elapsed), "eta", utils.PrettyDuration(eta))
		f.prevLog = time.Now()
	}
	return
}

func Wrap(r io.Reader, name string, size uint64, period time.Duration) *Filelog {
	return &Filelog{
		Reader: r,
		name:   name,
		size:   size,
		period: period,
	}
}
