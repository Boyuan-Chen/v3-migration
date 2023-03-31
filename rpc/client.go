package rpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type RpcClient struct {
	Client client.RPC
}

func NewRpcClient(endpoint string, secret [32]byte) (*RpcClient, error) {
	l2EndPointConfig := &node.L2EndpointConfig{
		L2EngineAddr:      endpoint,
		L2EngineJWTSecret: secret,
	}
	logger := log.New("hash")
	client, err := l2EndPointConfig.Setup(context.Background(), logger)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize RPC Client: %v", err)
	}
	return &RpcClient{Client: client}, nil
}

func (rpc *RpcClient) GetBlock() (*Block, error) {
	var block *Block
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := rpc.Client.CallContext(ctx, &block, "eth_getBlockByNumber", "latest", false); err != nil {
		return nil, errors.New("Failed to obtain head block")
	}
	return block, nil
}

func (rpc *RpcClient) GetNextNonce(account *common.Address) (uint64, error) {
	var nonce hexutil.Uint64
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := rpc.Client.CallContext(ctx, &nonce, "eth_getTransactionCount", account, "pending"); err != nil {
		fmt.Println(err)
		return 0, errors.New("Failed to obtain nonce")
	}
	return uint64(nonce), nil
}

func (rpc *RpcClient) GetGasPrice() (*big.Int, error) {
	var hex hexutil.Big
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := rpc.Client.CallContext(ctx, &hex, "eth_gasPrice"); err != nil {
		return nil, errors.New("Failed to obtain gas price")
	}
	return (*big.Int)(&hex), nil
}

// func (rpc *RpcClient) GetPendingTransaction() (*[]interface{}, error) {
// 	var pendingTransactions *[]interface{}
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	if err := rpc.Client.CallContext(ctx, pendingTransactions, "txpool_status"); err != nil {
// 		fmt.Println(err)
// 		return nil, errors.New("Failed to obtain pending transactions")
// 	}
// 	return pendingTransactions, nil
// }

func (rpc *RpcClient) SendRawTransaction(tx *types.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	data, err := tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("Failed to marshal transaction: %v", err)
	}
	err = rpc.Client.CallContext(ctx, nil, "eth_sendRawTransaction", hexutil.Encode(data))
	if err != nil {
		return fmt.Errorf("Failed to send transaction: %v", err)
	}
	return nil
}
