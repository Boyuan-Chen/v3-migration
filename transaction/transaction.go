package transaction

import (
	"fmt"
	"math/big"

	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type TransactionBuilder struct {
	RpcClient    *rpc.RpcClient
	RollupConfig *rollup.Config
}

func NewTransactionBuilder(rpcClient *rpc.RpcClient, rollupConfig *rollup.Config) *TransactionBuilder {
	return &TransactionBuilder{
		RpcClient:    rpcClient,
		RollupConfig: rollupConfig,
	}
}

func (t *TransactionBuilder) BuildTransaction(key string) (*types.Transaction, error) {
	ecskey, _ := crypto.HexToECDSA(key)
	address := crypto.PubkeyToAddress(ecskey.PublicKey)
	signer := types.NewEIP155Signer(t.RollupConfig.L2ChainID)
	nonce, err := t.RpcClient.GetNextNonce(&address)
	if err != nil {
		return nil, fmt.Errorf("Failed to get nonce: %s", err.Error())
	}
	gasPrice, err := t.RpcClient.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("Failed to get gas price: %s", err.Error())
	}
	unsignedTx := types.NewTransaction(
		nonce,
		common.HexToAddress("0x00000000000000000000000000000000deadbeef"),
		new(big.Int),
		5000000,
		gasPrice,
		[]byte{},
	)
	tx, err := types.SignTx(unsignedTx, signer, ecskey)
	if err != nil {
		return nil, fmt.Errorf("Failed to sign transaction: %s", err.Error())
	}
	return tx, nil
}

func (t *TransactionBuilder) SubmitTransaction(key string) error {
	tx, err := t.BuildTransaction(key)
	if err != nil {
		return fmt.Errorf("Failed to build transaction: %s", err.Error())
	}
	err = t.RpcClient.SendRawTransaction(tx)
	if err != nil {
		return fmt.Errorf("Failed to send transaction: %s", err.Error())
	}
	fmt.Println("-> Test Transaction Submitted")
	return nil
}
