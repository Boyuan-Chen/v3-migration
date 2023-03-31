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
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// for local testing
	endpoint         = "http://localhost:8551"
	secretConfigPath = "./static/test-jwt-secret.txt"
	rollupConfigPath = "./static/rollup.json"
	key              = "8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61"
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

	// Get nonce
	ecskey, _ := crypto.HexToECDSA(key)
	address := crypto.PubkeyToAddress(ecskey.PublicKey)
	nonce, err := rpcClient.GetNextNonce(&address)
	if err != nil {
		Exit(err)
	}

	// Submit transaction
	transactionBuilder := transaction.NewTransactionBuilder(rpcClient, rollupConfig)
	err = transactionBuilder.SubmitTransaction(key)
	if err != nil {
		Exit(err)
	}

	// Check nonce
	newNonce, err := rpcClient.GetNextNonce(&address)
	if err != nil {
		Exit(err)
	}
	if newNonce != nonce+1 {
		fmt.Println("Nonce is not bumped... Maybe need to wait for a while? The logic is not added yet.")
		os.Exit(1)
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

	var transactions []eth.Data
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

			if len(newBlock.Transactions) != 1 {
				fmt.Println("Invalid new block - block should have 1 transaction")
				os.Exit(1)
			}

			fmt.Println("-> Successfully Submit Block to L2")
			break
		}
		time.Sleep(time.Second * 1)
	}
}
