// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package tests

import (
	"encoding/json"
	"math/big"

	"github.com/iceming123/go-ice/common"
	"github.com/iceming123/go-ice/common/hexutil"
	"github.com/iceming123/go-ice/common/math"
	"github.com/iceming123/go-ice/core/types"
)

var _ = (*btHeaderMarshaling)(nil)

func (b btHeader) MarshalJSON() ([]byte, error) {
	type btHeader struct {
		CommitteeHash		common.Hash
		SnailHash			common.Hash
		SnailNumber			 *math.HexOrDecimal256
		Bloom            types.Bloom
		Number           *math.HexOrDecimal256
		Hash             common.Hash
		ParentHash       common.Hash
		ReceiptsRoot      common.Hash
		StateRoot        common.Hash
		TransactionsRoot common.Hash
		ExtraData        hexutil.Bytes
		GasLimit         math.HexOrDecimal64
		GasUsed          math.HexOrDecimal64
		Timestamp        *math.HexOrDecimal256
	}
	var enc btHeader
	enc.CommitteeHash = b.CommitteeHash
	enc.SnailHash = b.SnailHash
	enc.SnailNumber =  (*math.HexOrDecimal256)(b.SnailNumber)
	enc.Bloom = b.Bloom
	enc.Number = (*math.HexOrDecimal256)(b.Number)
	enc.Hash = b.Hash
	enc.ParentHash = b.ParentHash
	enc.ReceiptsRoot = b.ReceiptsRoot
	enc.StateRoot = b.StateRoot
	enc.TransactionsRoot = b.TransactionsRoot
	enc.ExtraData = b.ExtraData
	enc.GasLimit = math.HexOrDecimal64(b.GasLimit)
	enc.GasUsed = math.HexOrDecimal64(b.GasUsed)
	enc.Timestamp = (*math.HexOrDecimal256)(b.Timestamp)
	return json.Marshal(&enc)
}

func (b *btHeader) UnmarshalJSON(input []byte) error {
	type btHeader struct {
		CommitteeHash		*common.Hash
		SnailHash			*common.Hash
		SnailNumber			*math.HexOrDecimal256
		Bloom            *types.Bloom
		Number           *math.HexOrDecimal256
		Hash             *common.Hash
		ParentHash       *common.Hash
		ReceiptsRoot      *common.Hash
		StateRoot        *common.Hash
		TransactionsRoot *common.Hash
		ExtraData        *hexutil.Bytes
		GasLimit         *math.HexOrDecimal64
		GasUsed          *math.HexOrDecimal64
		Timestamp        *math.HexOrDecimal256
	}
	var dec btHeader
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	if dec.CommitteeHash != nil {
		b.CommitteeHash = *dec.CommitteeHash
	}
	if dec.SnailHash != nil {
		b.SnailHash = *dec.SnailHash
	}
	if dec.SnailNumber != nil {
		b.SnailNumber = (*big.Int)(dec.SnailNumber)
	}
	if dec.Bloom != nil {
		b.Bloom = *dec.Bloom
	}
	if dec.Number != nil {
		b.Number = (*big.Int)(dec.Number)
	}
	if dec.Hash != nil {
		b.Hash = *dec.Hash
	}
	if dec.ParentHash != nil {
		b.ParentHash = *dec.ParentHash
	}
	if dec.ReceiptsRoot != nil {
		b.ReceiptsRoot = *dec.ReceiptsRoot
	}
	if dec.StateRoot != nil {
		b.StateRoot = *dec.StateRoot
	}
	if dec.TransactionsRoot != nil {
		b.TransactionsRoot = *dec.TransactionsRoot
	}
	if dec.ExtraData != nil {
		b.ExtraData = *dec.ExtraData
	}
	if dec.GasLimit != nil {
		b.GasLimit = uint64(*dec.GasLimit)
	}
	if dec.GasUsed != nil {
		b.GasUsed = uint64(*dec.GasUsed)
	}
	if dec.Timestamp != nil {
		b.Timestamp = (*big.Int)(dec.Timestamp)
	}
	return nil
}
