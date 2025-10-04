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
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/panoptisDev/pano/utils/cser"
)

var ErrUnknownTxType = errors.New("unknown tx type")

func encodeSig(r, s *big.Int) (sig [64]byte) {
	copy(sig[0:], cser.PaddedBytes(r.Bytes(), 32)[:32])
	copy(sig[32:], cser.PaddedBytes(s.Bytes(), 32)[:32])
	return sig
}

func decodeSig(sig [64]byte) (r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	return
}

func TransactionMarshalCSER(w *cser.Writer, tx *types.Transaction) error {
	if tx.Type() != types.LegacyTxType && tx.Type() != types.AccessListTxType && tx.Type() != types.DynamicFeeTxType {
		return ErrUnknownTxType
	}
	if tx.Type() != types.LegacyTxType {
		// marker of a non-standard tx
		w.BitsW.Write(6, 0)
		// tx type
		w.U8(tx.Type())
	} else if tx.Gas() <= 0xff {
		return errors.New("cannot serialize legacy tx with gasLimit <= 256")
	}
	w.U64(tx.Nonce())
	w.U64(tx.Gas())
	if tx.Type() == types.DynamicFeeTxType {
		w.BigInt(tx.GasTipCap())
		w.BigInt(tx.GasFeeCap())
	} else {
		w.BigInt(tx.GasPrice())
	}
	w.BigInt(tx.Value())
	w.Bool(tx.To() != nil)
	if tx.To() != nil {
		w.FixedBytes(tx.To().Bytes())
	}
	w.SliceBytes(tx.Data())
	v, r, s := tx.RawSignatureValues()
	w.BigInt(v)
	sig := encodeSig(r, s)
	w.FixedBytes(sig[:])
	if tx.Type() == types.AccessListTxType || tx.Type() == types.DynamicFeeTxType {
		w.BigInt(tx.ChainId())
		w.U32(uint32(len(tx.AccessList())))
		for _, tuple := range tx.AccessList() {
			w.FixedBytes(tuple.Address.Bytes())
			w.U32(uint32(len(tuple.StorageKeys)))
			for _, h := range tuple.StorageKeys {
				w.FixedBytes(h.Bytes())
			}
		}
	}
	return nil
}

func TransactionUnmarshalCSER(r *cser.Reader) (*types.Transaction, error) {
	txType := uint8(types.LegacyTxType)
	if r.BitsR.View(6) == 0 {
		r.BitsR.Read(6)
		txType = r.U8()
	}

	nonce := r.U64()
	gasLimit := r.U64()
	var gasPrice *big.Int
	var gasTipCap *big.Int
	var gasFeeCap *big.Int
	if txType == types.DynamicFeeTxType {
		gasTipCap = r.BigInt()
		gasFeeCap = r.BigInt()
	} else {
		gasPrice = r.BigInt()
	}
	amount := r.BigInt()
	toExists := r.Bool()
	var to *common.Address
	if toExists {
		var _to common.Address
		r.FixedBytes(_to[:])
		to = &_to
	}
	data := r.SliceBytes(ProtocolMaxMsgSize)
	// sig
	v := r.BigInt()
	var sig [64]byte
	r.FixedBytes(sig[:])
	_r, s := decodeSig(sig)

	switch txType {
	case types.LegacyTxType:
		return types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       to,
			Value:    amount,
			Data:     data,
			V:        v,
			R:        _r,
			S:        s,
		}), nil
	case types.AccessListTxType, types.DynamicFeeTxType:
		chainID := r.BigInt()
		accessListLen := r.U32()
		if accessListLen > ProtocolMaxMsgSize/24 {
			return nil, cser.ErrTooLargeAlloc
		}
		accessList := make(types.AccessList, accessListLen)
		for i := range accessList {
			r.FixedBytes(accessList[i].Address[:])
			keysLen := r.U32()
			if keysLen > ProtocolMaxMsgSize/32 {
				return nil, cser.ErrTooLargeAlloc
			}
			accessList[i].StorageKeys = make([]common.Hash, keysLen)
			for j := range accessList[i].StorageKeys {
				r.FixedBytes(accessList[i].StorageKeys[j][:])
			}
		}
		if txType == types.AccessListTxType {
			return types.NewTx(&types.AccessListTx{
				ChainID:    chainID,
				Nonce:      nonce,
				GasPrice:   gasPrice,
				Gas:        gasLimit,
				To:         to,
				Value:      amount,
				Data:       data,
				AccessList: accessList,
				V:          v,
				R:          _r,
				S:          s,
			}), nil
		} else {
			return types.NewTx(&types.DynamicFeeTx{
				ChainID:    chainID,
				Nonce:      nonce,
				GasTipCap:  gasTipCap,
				GasFeeCap:  gasFeeCap,
				Gas:        gasLimit,
				To:         to,
				Value:      amount,
				Data:       data,
				AccessList: accessList,
				V:          v,
				R:          _r,
				S:          s,
			}), nil
		}
	}
	return nil, ErrUnknownTxType
}
