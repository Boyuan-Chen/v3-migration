package transaction

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Boyuan-Chen/v3-migration/abi"
	"github.com/Boyuan-Chen/v3-migration/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SmartContractViewer struct {
	L1EthClient *ethclient.Client
	L2EthClient *ethclient.Client
}

func NewSmartContractViewer(l1Client *ethclient.Client, l2Client *ethclient.Client) *SmartContractViewer {
	return &SmartContractViewer{
		L1EthClient: l1Client,
		L2EthClient: l2Client,
	}
}

func (t *SmartContractViewer) GetNonceFromL1Messenger() (*big.Int, error) {
	abiSelector, err := abi.GetABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %s", err.Error())
	}
	callData, err := abiSelector.Pack("messageNonce")
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %s", err.Error())
	}
	callMsg := ethereum.CallMsg{
		To:   &utils.L1MessengerAddress,
		Data: callData,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := t.L1EthClient.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %s", err.Error())
	}
	unpackRes, err := abiSelector.Unpack("messageNonce", res)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %s", err.Error())
	}
	nonce, ok := unpackRes[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to *big.Int")
	}
	return nonce, nil
}

func (t *SmartContractViewer) GetBOBABalance(address *common.Address) (*big.Int, error) {
	abiSelector, err := abi.GetABI()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %s", err.Error())
	}
	callData, err := abiSelector.Pack("balanceOf", address)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %s", err.Error())
	}
	callMsg := ethereum.CallMsg{
		To:   &utils.L2BobaAddress,
		Data: callData,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := t.L2EthClient.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %s", err.Error())
	}
	unpackRes, err := abiSelector.Unpack("balanceOf", res)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %s", err.Error())
	}
	balance, ok := unpackRes[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to *big.Int")
	}
	return balance, nil
}
