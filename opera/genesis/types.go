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

package genesis

import (
	"io"

	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/panoptisDev/pano/inter/ibr"
	"github.com/panoptisDev/pano/inter/ier"
	"github.com/panoptisDev/pano/scc/cert"
)

type (
	Hashes map[string]hash.Hash
	Header struct {
		GenesisID   hash.Hash
		NetworkID   uint64
		NetworkName string
	}
	Blocks interface {
		ForEach(fn func(ibr.LlrIdxFullBlockRecord) bool)
	}
	Epochs interface {
		ForEach(fn func(ier.LlrIdxFullEpochRecord) bool)
	}
	EvmItems interface {
		ForEach(fn func(key, value []byte) bool)
	}
	SccCommitteeCertificates interface {
		ForEach(fn func(cert.Certificate[cert.CommitteeStatement]) bool)
	}
	SccBlockCertificates interface {
		ForEach(fn func(cert.Certificate[cert.BlockStatement]) bool)
	}
	FwsLiveSection interface {
		GetReader() (io.Reader, error)
	}
	FwsArchiveSection interface {
		GetReader() (io.Reader, error)
	}
	SignatureSection interface {
		GetSignature() ([]byte, error)
	}
	SignedMetadata struct {
		Signature []byte
		Hashes    []byte
	}
	Genesis struct {
		Header

		Blocks                Blocks
		Epochs                Epochs
		RawEvmItems           EvmItems
		CommitteeCertificates SccCommitteeCertificates
		BlockCertificates     SccBlockCertificates
		FwsLiveSection
		FwsArchiveSection
		SignatureSection
	}
)

func (hh Hashes) Includes(hh2 Hashes) bool {
	for n, h := range hh {
		if hh2[n] != h {
			return false
		}
	}
	return true
}

func (hh Hashes) Equal(hh2 Hashes) bool {
	return hh.Includes(hh2) && hh2.Includes(hh)
}

func (h Header) Equal(h2 Header) bool {
	return h == h2
}
