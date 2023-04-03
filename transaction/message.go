package transaction

import (
	"fmt"
	"math/big"

	"github.com/Boyuan-Chen/v3-migration/abi"
	"github.com/Boyuan-Chen/v3-migration/utils"
	"github.com/ethereum/go-ethereum/common"
)

func BuildBobaDepositFromL1ToL2(from *common.Address, amount *big.Int, nonce *big.Int) ([]byte, error) {
	// Build data

	// message from L1StandardBridge to L2CrossDomainMessenger
	abiSelector, err := abi.GetABI()
	if err != nil {
		return nil, fmt.Errorf("Failed to get abi: %s", err.Error())
	}

	msgFromBridgeToCDM, err := abiSelector.Pack(
		"finalizeBridgeERC20",
		utils.L2BobaAddress,
		utils.L1BobaAddress,
		from,
		from,
		amount,
		[]byte{},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to pack msgFromBridgeToCDM data: %s", err.Error())
	}

	msgFromCDMToPort, err := abiSelector.Pack(
		"relayMessage",
		nonce,
		utils.L1StandardBridgeAddress,
		utils.L2StandardBridgeAddress,
		big.NewInt(0),
		big.NewInt(1000000),
		msgFromBridgeToCDM,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to pack msgFromCDMToPort data: %s", err.Error())
	}
	return msgFromCDMToPort, nil
}
