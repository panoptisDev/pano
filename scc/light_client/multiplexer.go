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

package light_client

import (
	"errors"
	"fmt"

	"github.com/panoptisDev/carmen/go/carmen"
	"github.com/panoptisDev/pano/scc"
	"github.com/panoptisDev/pano/scc/cert"
	"github.com/panoptisDev/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
)

// multiplexer is a provider that distributes requests across multiple providers.
// It attempts to retrieve data by calling each provider in sequence until one
// returns a nil error.
type multiplexer struct {
	providers []provider
}

// newMultiplexer creates a new multiplexer instance.
//
// Parameters:
//   - providers: A list of provider instances to be used by the multiplexer.
//
// Returns:
//   - *multiplexer: A pointer to a new multiplexer instance.
//   - error: Returns an error if no providers are given.
func newMultiplexer(providers ...provider) (*multiplexer, error) {
	if len(providers) == 0 {
		return nil, errors.New("no providers provided")
	}
	return &multiplexer{providers: providers}, nil
}

// Close closes all the providers in the Multiplexer.
// Closing an already closed provider has no effect.
func (m *multiplexer) close() {
	for _, p := range m.providers {
		p.close()
	}
}

// getCommitteeCertificates returns up to `maxResults` consecutive committee
// certificates starting from the given period.
//
// Parameters:
// - first: The starting period for which to retrieve committee certificates.
// - maxResults: The maximum number of committee certificates to retrieve.
//
// Returns:
//   - []cert.CommitteeCertificate: A slice of committee certificates.
//   - error: Not nil if the provider failed to obtain the requested certificates.
func (m *multiplexer) getCommitteeCertificates(first scc.Period, maxResults uint64) ([]cert.CommitteeCertificate, error) {
	return tryAll(m.providers, func(p provider) ([]cert.CommitteeCertificate, error) {
		return p.getCommitteeCertificates(first, maxResults)
	})
}

// getBlockCertificates returns up to `maxResults` consecutive block
// certificates starting from the given block number.
//
// Parameters:
//   - number: The starting block number for which to retrieve the block certificate.
//     Can be LatestBlock to retrieve the latest certificates.
//   - maxResults: The maximum number of block certificates to retrieve.
//
// Returns:
//   - cert.BlockCertificate: The block certificates for the given block number
//     and the following blocks.
//   - error: Not nil if the provider failed to obtain the requested certificates.
func (m *multiplexer) getBlockCertificates(first idx.Block, maxResults uint64) ([]cert.BlockCertificate, error) {
	return tryAll(m.providers, func(p provider) ([]cert.BlockCertificate, error) {
		return p.getBlockCertificates(first, maxResults)
	})
}

// getAccountProof returns the account proof corresponding to the
// given address at the given height.
//
// Parameters:
//   - address: The address of the account.
//   - height: The block height of the state.
//
// Returns:
//   - WitnessProof: witness proof for the given account.
//   - error: Not nil if the provider failed to obtain the requested account proof.
func (m *multiplexer) getAccountProof(address common.Address, height idx.Block) (carmen.WitnessProof, error) {
	return tryAll(m.providers, func(p provider) (carmen.WitnessProof, error) {
		return p.getAccountProof(address, height)
	})
}

// tryAll executes a function on each provider in sequence until one returns a nil error.
//
// This function is a generic helper that iterates over a list of providers,
// calling the given function on each. If any function call succeeds, it immediately
// returns the result. If all calls fail, it aggregates the errors and returns them.
//
// Type Parameters:
//   - C: The type of the result returned by the function.
//
// Parameters:
//   - ps: A slice of provider instances to be tried.
//   - fn: A function that takes a provider and returns a result of type C and an error.
//
// Returns:
//   - C: The result of the first successful function execution.
//   - error: Nil if at least one execution of fn returned without error.
//     The joined error of all failed attempts if all attempts failed.
func tryAll[C any](ps []provider, fn func(provider) (C, error)) (C, error) {
	var errs []error
	for _, p := range ps {
		res, err := fn(p)
		if err == nil {
			return res, nil
		}
		errs = append(errs, err)
	}
	var c C
	return c, errors.Join(fmt.Errorf("all providers failed: "), errors.Join(errs...))
}
