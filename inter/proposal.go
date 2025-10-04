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
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/panoptisDev/pano/gossip/randao"
	"github.com/panoptisDev/pano/inter/pb"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/protobuf/proto"
)

// Proposal represents a block proposal in the consensus protocol.
// It contains the block number, parent hash, timestamp, previous randao,
// and a list of transactions.
//
// A proposal is a candidate for inclusion in the blockchain and is
// created by a proposer. It is signed by the proposer and sent to
// validators for validation and inclusion in the blockchain.
type Proposal struct {
	Number       idx.Block
	ParentHash   common.Hash
	RandaoReveal randao.RandaoReveal
	Transactions []*types.Transaction
}

// Hash computes a cryptographic hash of the proposal. The hash can be used to
// sign and verify the proposal.
func (p *Proposal) Hash() hash.Hash {
	data := []byte{}
	data = binary.BigEndian.AppendUint64(data, uint64(p.Number))
	data = append(data, p.ParentHash[:]...)
	data = append(data, p.RandaoReveal[:]...)
	for _, tx := range p.Transactions {
		txHash := tx.Hash()
		data = append(data, txHash[:]...)
	}
	return sha256.Sum256(data)
}

func (p *Proposal) Serialize() ([]byte, error) {
	pb, err := p.toProto()
	if err != nil {
		return nil, err
	}
	return proto.Marshal(pb)
}

func (p *Proposal) Deserialize(data []byte) error {
	var pb pb.Proposal
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	return p.fromProto(&pb)
}

func (p *Proposal) toProto() (*pb.Proposal, error) {
	transactions := make([]*pb.Transaction, 0, len(p.Transactions))
	for _, tx := range p.Transactions {
		data, err := tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &pb.Transaction{
			Encoded: data,
		})
	}

	return &pb.Proposal{
		Number:       uint64(p.Number),
		ParentHash:   p.ParentHash[:],
		RandaoReveal: p.RandaoReveal[:],
		Transactions: transactions,
	}, nil
}

func (p *Proposal) fromProto(pb *pb.Proposal) error {
	// Restore individual fields.
	p.Number = idx.Block(pb.Number)
	copy(p.ParentHash[:], pb.ParentHash)
	copy(p.RandaoReveal[:], pb.RandaoReveal)
	for _, tx := range pb.Transactions {
		if tx == nil {
			return fmt.Errorf("nil transaction in proposal")
		}
		var transaction types.Transaction
		if err := transaction.UnmarshalBinary(tx.Encoded); err != nil {
			return err
		}
		p.Transactions = append(p.Transactions, &transaction)
	}

	return nil
}
