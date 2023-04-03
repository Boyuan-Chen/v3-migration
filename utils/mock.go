package utils

import (
	"encoding/binary"
	"math/big"

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

// optimism/packages/contracts-bedrock/contracts/vendor/AddressAliasHelper.sol
func ApplyL1ToL2Alias(address common.Address) common.Address {
	offset := common.HexToAddress("0x1111000000000000000000000000000000001111")
	sum := common.BigToAddress(new(big.Int).Add(new(big.Int).SetBytes(address[:]), new(big.Int).SetBytes(offset[:])))
	return sum
}

// optimism/packages/contracts-bedrock/contracts/libraries/Encoding.sol
// assembly {
// nonce := or(shl(240, _version), _nonce)
// }
func EncodeNonce(nonce *big.Int) *big.Int {
	nonceShift := nonce.Lsh(nonce, 240)
	return nonceShift.Or(nonceShift, common.Big1)
}
