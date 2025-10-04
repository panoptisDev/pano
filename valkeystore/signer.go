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

package valkeystore

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/panoptisDev/pano/inter/validatorpk"
	"github.com/panoptisDev/pano/valkeystore/encryption"
)

//go:generate mockgen -source=signer.go -destination=signer_mock.go  -package=valkeystore

// SignerAuthority is an interface for signing operations with a validator's key.
type SignerAuthority interface {
	// Sign signs a given digest using the validator's private key.
	Sign(digest common.Hash) ([]byte, error)
	// Returns the public key of the validator.
	PublicKey() validatorpk.PubKey
}

type signerAuthorityImpl struct {
	backend KeystoreI
	pubkey  validatorpk.PubKey
}

// NewSignerAuthority constructs a new SignerAuthority using the provided keystore and public key.
// The validator's private key is expected to be stored in the keystore
func NewSignerAuthority(store KeystoreI, pubkey validatorpk.PubKey) SignerAuthority {
	return &signerAuthorityImpl{
		backend: store,
		pubkey:  pubkey,
	}
}

func (s *signerAuthorityImpl) Sign(digest common.Hash) ([]byte, error) {
	if s.pubkey.Type != validatorpk.Types.Secp256k1 {
		return nil, encryption.ErrNotSupportedType
	}
	key, err := s.backend.GetUnlocked(s.pubkey)
	if err != nil {
		return nil, err
	}

	secp256k1Key := key.Decoded.(*ecdsa.PrivateKey)

	sigRSV, err := crypto.Sign(digest[:], secp256k1Key)
	if err != nil {
		return nil, err
	}
	sigRS := sigRSV[:64]
	return sigRS, err
}

func (s *signerAuthorityImpl) PublicKey() validatorpk.PubKey {
	return s.pubkey.Copy()
}

// VerifySignature verifies that the provided signature is valid for the given digest
// using the provided public key. It returns true if the signature is valid, false otherwise.
func VerifySignature(digest common.Hash, signature []byte, pubkey validatorpk.PubKey) bool {
	if pubkey.Type != validatorpk.Types.Secp256k1 {
		return false
	}
	return crypto.VerifySignature(pubkey.Raw, digest[:], signature)
}
