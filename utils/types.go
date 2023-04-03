package utils

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	Uint256Type, _ = abi.NewType("uint256", "", nil)
	Uint64Type, _  = abi.NewType("uint64", "", nil)
	BytesType, _   = abi.NewType("bytes", "", nil)
	BoolType, _    = abi.NewType("bool", "", nil)
	AddressType, _ = abi.NewType("address", "", nil)
)

// abi.encodePacked()
// It can not be replaced by the abi.Pack()
// You have to use the following code to build the data
func EncodePacked(input ...[]byte) []byte {
	return bytes.Join(input, nil)
}

func EncodeBytesString(v string) []byte {
	decoded, err := hex.DecodeString(v)
	if err != nil {
		panic(err)
	}
	return decoded
}

func EncodeUint256(v string) []byte {
	bn := new(big.Int)
	bn.SetString(v, 10)
	return math.U256Bytes(bn)
}

func EncodeUint256Array(arr []string) []byte {
	var res [][]byte
	for _, v := range arr {
		b := EncodeUint256(v)
		res = append(res, b)
	}
	return bytes.Join(res, nil)
}

func EncodeBool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
}
