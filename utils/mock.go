package utils

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

func MockL1Hash(num uint64) (out common.Hash) {
	out[31] = 1
	binary.BigEndian.PutUint64(out[:], num)
	return
}

func MockL2Hash(num uint64) (out common.Hash) {
	out[31] = 2
	binary.BigEndian.PutUint64(out[:], num)
	return
}
