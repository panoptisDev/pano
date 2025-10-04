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

package gossip

import (
	"testing"

	"github.com/panoptisDev/pano/eventcheck"
	"github.com/panoptisDev/pano/eventcheck/gaspowercheck"
	"github.com/panoptisDev/pano/eventcheck/parentscheck"
	"github.com/panoptisDev/pano/eventcheck/proposalcheck"
	"github.com/panoptisDev/pano/inter"
	parentscheckbase "github.com/panoptisDev/lachesis-base/eventcheck/parentscheck"
	"github.com/panoptisDev/lachesis-base/hash"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/panoptisDev/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/p2p/discover/discfilter"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestValidateEventPropertiesDependingOnParents(t *testing.T) {

	tests := map[string]struct {
		modify   func(*inter.MutableEventPayload)
		expected error
	}{
		"valid event": {
			modify: func(event *inter.MutableEventPayload) {},
		},
		"parents check violation": {
			modify: func(event *inter.MutableEventPayload) {
				event.SetLamport(2)
			},
			expected: parentscheckbase.ErrWrongLamport,
		},
		"gas power check violation": {
			modify: func(event *inter.MutableEventPayload) {
				event.SetGasPowerLeft(inter.GasPowerLeft{
					Gas: [inter.GasPowerConfigs]uint64{1000, 2000},
				})
			},
			expected: gaspowercheck.ErrWrongGasPowerLeft,
		},
		"proposal check violation": {
			modify: func(event *inter.MutableEventPayload) {
				event.SetVersion(3)
				event.SetPayload(inter.Payload{
					ProposalSyncState: inter.ProposalSyncState{
						LastSeenProposalTurn: 75,
					},
				})
			},
			expected: proposalcheck.ErrSyncStateProgressionWithoutProposal,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			gasPowerCheckReader := gaspowercheck.NewMockReader(ctrl)
			proposalCheckReader := proposalcheck.NewMockReader(ctrl)

			checkers := &eventcheck.Checkers{
				Parentscheck:  parentscheck.New(),
				Gaspowercheck: gaspowercheck.New(gasPowerCheckReader),
				Proposalcheck: proposalcheck.New(proposalCheckReader),
			}

			epoch := idx.Epoch(12)

			creator := idx.ValidatorID(1)
			validatorsBuilder := pos.ValidatorsBuilder{}
			validatorsBuilder.Set(creator, pos.Weight(100))
			validators := validatorsBuilder.Build()

			// Create a parent event.
			builder := inter.MutableEventPayload{}
			builder.SetEpoch(epoch)
			builder.SetCreator(creator)
			builder.SetSeq(1)
			parent := builder.Build()

			// Create the event to be tested.
			builder = inter.MutableEventPayload{}
			builder.SetEpoch(epoch)
			builder.SetCreator(creator)
			builder.SetLamport(1)
			builder.SetSeq(2)
			builder.SetCreationTime(1)
			builder.SetParents([]hash.Event{parent.ID()})

			test.modify(&builder)

			event := builder.Build()

			// Set up the validation context.
			validationContext := &gaspowercheck.ValidationContext{
				Epoch:           epoch,
				Validators:      validators,
				ValidatorStates: []gaspowercheck.ValidatorState{{}},
			}
			gasPowerCheckReader.EXPECT().GetValidationContext().Return(validationContext).AnyTimes()

			proposalCheckReader.EXPECT().GetEventPayload(gomock.Any()).Return(inter.Payload{}).AnyTimes()

			// Run the actual check.
			require.ErrorIs(t, validateEventPropertiesDependingOnParents(
				checkers,
				event,
				[]inter.EventI{parent},
			), test.expected)
		})
	}
}

func TestIsUseless(t *testing.T) {
	validEnode := enode.MustParse("enode://3f4306c065eaa5d8079e17feb56c03a97577e67af3c9c17496bb8916f102f1ff603e87d2a4ebfa0a2f70b780b85db212618857ea4e9627b24a9b0dd2faeb826e@127.0.0.1:5050")
	panoName := "Pano/v1.0.0-a-61af51c2-1715085138/linux-amd64/go1.21.7"
	operaName := "go-opera/v1.1.2-rc.6-8e84c9dc-1688013329/linux-amd64/go1.19.11"
	invalidName := "bot"

	discfilter.Enable()
	if isUseless(validEnode, panoName) {
		t.Errorf("pano peer reported as useless")
	}
	if isUseless(validEnode, operaName) {
		t.Errorf("opera peer reported as useless")
	}
	if !isUseless(validEnode, invalidName) {
		t.Errorf("invalid peer not reported as useless")
	}
	if !isUseless(validEnode, operaName) {
		t.Errorf("peer not banned after marking as useless")
	}
}
