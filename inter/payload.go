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

	"github.com/panoptisDev/pano/inter/pb"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"google.golang.org/protobuf/proto"
)

const (
	// PayloadVersion is the version of the payload format.
	currentPayloadVersion = 1
)

// Payload is the content of an event of version 3. Unlike previous formats,
// defining new RLP encoded content, this payload uses protobuf encoding to
// standardize the serialization of the content and simplify portability.
type Payload struct {
	// ProposalSyncState keeps track of the turn-taking of proposers and enables
	// the decision of whose turn it is to propose a block. This information is
	// present in all events.
	ProposalSyncState
	// Proposal is an optional proposal for a new block that can be included in
	// the payload of an event only by a producer who is allowed to do so based
	// on the tracked sync state.
	Proposal *Proposal
}

// Hash computes a secure hash of the payload that can be used for signing and
// verifying the payload.
func (e *Payload) Hash() hash.Hash {
	data := []byte{currentPayloadVersion}
	data = binary.BigEndian.AppendUint32(data, uint32(e.LastSeenProposalTurn))
	data = binary.BigEndian.AppendUint32(data, uint32(e.LastSeenProposalFrame))
	if e.Proposal != nil {
		hash := e.Proposal.Hash()
		data = append(data, hash[:]...)
	}
	return sha256.Sum256(data)
}

func (e *Payload) Serialize() ([]byte, error) {
	var proposal *pb.Proposal
	if e.Proposal != nil {
		p, err := e.Proposal.toProto()
		if err != nil {
			return nil, err
		}
		proposal = p
	}
	return proto.Marshal(&pb.Payload{
		Version:               currentPayloadVersion,
		LastSeenProposalTurn:  uint32(e.LastSeenProposalTurn),
		LastSeenProposalFrame: uint32(e.LastSeenProposalFrame),
		Proposal:              proposal,
	})
}

func (e *Payload) Deserialize(data []byte) error {
	var pb pb.Payload
	if err := proto.Unmarshal(data, &pb); err != nil {
		return err
	}
	if pb.Version != currentPayloadVersion {
		return fmt.Errorf("unsupported payload version: %d", pb.Version)
	}
	e.LastSeenProposalTurn = Turn(pb.LastSeenProposalTurn)
	e.LastSeenProposalFrame = idx.Frame(pb.LastSeenProposalFrame)
	if pb.Proposal != nil {
		p := &Proposal{}
		if err := p.fromProto(pb.Proposal); err != nil {
			return err
		}
		e.Proposal = p
	} else {
		e.Proposal = nil
	}
	return nil
}
