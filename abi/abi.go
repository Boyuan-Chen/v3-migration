package abi

import (
	"fmt"
	"strings"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
)

const jsondata = `
[
  {"inputs": [], "name": "messageNonce", "outputs": [{"internalType": "uint256","name": "","type": "uint256"}],"stateMutability": "view","type": "function"},
  {"inputs": [{"internalType": "address","name": "_localToken","type": "address"},{"internalType": "address","name": "_remoteToken","type": "address"},{"internalType": "address","name": "_from","type": "address"},{"internalType": "address","name": "_to","type": "address"},{"internalType": "uint256","name": "_amount","type": "uint256"},{"internalType": "bytes","name": "_extraData","type": "bytes"}],"name": "finalizeBridgeERC20","outputs": [],"stateMutability": "nonpayable","type": "function"},
  {"inputs": [{"internalType": "uint256","name": "_nonce","type": "uint256"},{"internalType": "address","name": "_sender","type": "address"},{"internalType": "address","name": "_target","type": "address"},{"internalType": "uint256","name": "_value","type": "uint256"},{"internalType": "uint256","name": "_minGasLimit","type": "uint256"},{"internalType": "bytes","name": "_message","type": "bytes"}],"name": "relayMessage","outputs": [],"stateMutability": "payable","type": "function"},
	{"inputs": [{"internalType": "address","name": "account","type": "address"}],"name": "balanceOf","outputs": [{"internalType": "uint256","name": "","type": "uint256"}],"stateMutability": "view","type": "function"}
]`

func GetABI() (*ethabi.ABI, error) {
	abi, err := ethabi.JSON(strings.NewReader(jsondata))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %s", err.Error())
	}
	return &abi, nil
}
