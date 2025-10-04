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

package ethapi

import (
	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/pano/scc/bls"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
)

// BlockCertificate is a JSON representation of a block certificate
// as returned by the Pano API. This type provides a conversion between the
// internal certificate representation and the JSON representation provided to
// the API clients. The external API is expected to be stable over time and
// should only be updated in a backward compatible way.
type BlockCertificate struct {
	ChainId   uint64                    `json:"chainId"`
	Number    uint64                    `json:"number"`
	Hash      common.Hash               `json:"hash"`
	StateRoot common.Hash               `json:"stateRoot"`
	Signers   cert.BitSet[scc.MemberId] `json:"signers"`
	Signature bls.Signature             `json:"signature"`
}

func (b BlockCertificate) ToCertificate() cert.BlockCertificate {
	aggregatedSignature := cert.NewAggregatedSignature[cert.BlockStatement](
		b.Signers, b.Signature)

	newCert := cert.NewCertificateWithSignature(
		cert.NewBlockStatement(
			b.ChainId,
			idx.Block(b.Number),
			b.Hash,
			b.StateRoot),
		aggregatedSignature)

	return newCert
}

func toJsonBlockCertificate(b cert.BlockCertificate) BlockCertificate {
	sub := b.Subject()
	agg := b.Signature()
	return BlockCertificate{
		ChainId:   sub.ChainId,
		Number:    uint64(sub.Number),
		Hash:      sub.Hash,
		StateRoot: sub.StateRoot,
		Signers:   agg.Signers(),
		Signature: agg.Signature(),
	}
}
