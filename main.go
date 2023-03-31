package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/engineapi"
	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/Boyuan-Chen/v3-migration/transaction"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	// for local testing
	endpoint         = "http://localhost:8551"
	secretConfigPath = "./static/test-jwt-secret.txt"
	rollupConfigPath = "./static/rollup.json"
	// 0xcd3B766CCDd6AE721141F452C550Ca635964ce71
	key1 = "8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61"
	// 0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097
	key2 = "c526ee95bf44d8fc405a158bb884d9d1238d99f0612e9f33d006bb0789009aaa"
)

func Exit(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	// Create config
	config := config.NewConfig(secretConfigPath, rollupConfigPath)

	// Get rollup config
	rollupConfig, secret, err := config.GetConfig()
	if err != nil {
		Exit(err)
	}

	// create Rpc client
	rpcClient, err := rpc.NewRpcClient(endpoint, *secret)
	if err != nil {
		Exit(err)
	}

	// Submit transaction
	transactionBuilder := transaction.NewTransactionBuilder(rpcClient, rollupConfig)
	err = transactionBuilder.SubmitTransaction(key1)
	if err != nil {
		Exit(err)
	}

	// Build another transaction and add it in attributes
	// This is just like the system transaction, but it is a standard transaction
	tx, err := transactionBuilder.BuildTestTransaction(key2)
	if err != nil {
		Exit(err)
	}
	binaryTx, err := tx.MarshalBinary()
	if err != nil {
		Exit(err)
	}

	// Play with engine_api
	// Get latest block
	block, err := rpcClient.GetBlock()
	if err != nil {
		Exit(err)
	}

	// Create engine
	engine, err := engineapi.NewEngineAPI(rpcClient, rollupConfig)
	if err != nil {
		Exit(err)
	}

	// Step 1: Get payloadID
	// engine_forkchoiceUpdatedV1 -> Get payloadID
	fc := &eth.ForkchoiceState{
		HeadBlockHash:      block.Hash,
		SafeBlockHash:      block.Hash,
		FinalizedBlockHash: block.Hash,
	}

	transactions := make([]eth.Data, 1)
	// add transaction
	// This is just like the system transaction, but it is a standard transaction
	// This should be included first
	transactions[0] = binaryTx

	futureTimeStamp := block.Time + 1000
	gasLimit := hexutil.Uint64(15000000)
	attributes := &eth.PayloadAttributes{
		Timestamp:             hexutil.Uint64(futureTimeStamp),
		PrevRandao:            [32]byte{},
		SuggestedFeeRecipient: common.HexToAddress("0x4200000000000000000000000000000000000011"),
		Transactions:          transactions,
		NoTxPool:              true,
		GasLimit:              (*eth.Uint64Quantity)(&gasLimit),
	}

	fcUpdateRes, err := engine.ForkchoiceUpdate(fc, attributes)
	if err != nil {
		Exit(err)
	}

	// Step 2: Get executionPayload
	// engine_getPayloadV1 -> Get executionPayload
	executionRes, err := engine.GetPayload(fcUpdateRes.PayloadID)
	if err != nil {
		Exit(err)
	}
	// verify transaction number
	if len(executionRes.Transactions) != 2 {
		fmt.Println("Invalid execution payload - should have 2 transaction")
		os.Exit(1)
	}

	// Step 3: Execute payload
	// engine_newPayloadV1 -> Execute payload
	_, err = engine.ExecutePayload(executionRes)
	if err != nil {
		Exit(err)
	}

	// Step 4: Submit block
	// engine_executePayloadV1 -> Submit block
	newfc := &eth.ForkchoiceState{
		HeadBlockHash:      executionRes.BlockHash,
		SafeBlockHash:      executionRes.BlockHash,
		FinalizedBlockHash: executionRes.BlockHash,
	}
	_, err = engine.ForkchoiceUpdate(newfc, nil)
	if err != nil {
		Exit(err)
	}

	// check block number
	newBlock, err := rpcClient.GetBlock()
	if err != nil {
		Exit(err)
	}

	// security check
	for i := 0; i < 10; i++ {
		if newBlock.Number == block.Number+1 {

			if newBlock.ParentHash != block.Hash {
				fmt.Println("Invalid new block - block is equal to executionPayloadRes.BlockHash")
				os.Exit(1)
			}
			if newBlock.Hash != executionRes.BlockHash {
				fmt.Println("Invalid new block - block is equal to executionPayloadRes.BlockHash")
				os.Exit(1)
			}

			if len(newBlock.Transactions) != 2 {
				fmt.Println("Invalid new block - block should have 2 transaction")
				os.Exit(1)
			}

			fmt.Println("-> Successfully Submit Block to L2")
			break
		}
		time.Sleep(time.Second * 1)
	}
}
