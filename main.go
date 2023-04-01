package main

import (
	"fmt"
	"math/big"
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
	// 0xcd3B766CCDd6AE721141F452C550Ca635964ce71
	key1 = "8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61"
	// 0xdF3e18d64BC6A983f673Ab319CCaE4f1a57C7097
	key2 = "c526ee95bf44d8fc405a158bb884d9d1238d99f0612e9f33d006bb0789009aaa"
	// 0xFABB0ac9d68B0B445fB7357272Ff202C5651694a
	key3          = "a267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1"
	mintETHAmount = big.NewInt(1000000000)
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

	// Build and submit transaction to tx pool
	transactionBuilder := transaction.NewTransactionBuilder(rpcClient, rollupConfig)
	err = transactionBuilder.SubmitTransaction(key1)
	if err != nil {
		Exit(err)
	}

	// Build a deposit transaction that mints ETH on L2
	ecskey, _ := crypto.HexToECDSA(key3)
	depositAddress := crypto.PubkeyToAddress(ecskey.PublicKey)
	depositETHTx, _ := transactionBuilder.BuildTestDepositETHTransaction(depositAddress, depositAddress, mintETHAmount)
	binaryDepositTx, err := transactionBuilder.MarshalBinary(depositETHTx)
	if err != nil {
		Exit(err)
	}

	// Build another transaction and add it in attributes
	// This is just like the system transaction, but it is a standard transaction
	tx, err := transactionBuilder.BuildTestTransaction(key2)
	if err != nil {
		Exit(err)
	}
	binaryTx, err := transactionBuilder.MarshalBinary(tx)
	if err != nil {
		Exit(err)
	}

	// Play with engine_api
	// Get latest block
	block, err := rpcClient.GetBlock()
	if err != nil {
		Exit(err)
	}

	// Get balance
	depositAddressBalance, err := rpcClient.GetBalance(depositAddress)
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

	transactions := make([]eth.Data, 2)
	// add transaction
	// This is just like the system transaction, but it is a standard transaction
	// This should be included first
	transactions[0] = binaryDepositTx
	transactions[1] = binaryTx

	futureTimeStamp := block.Time + 1000
	gasLimit := hexutil.Uint64(15000000)
	attributes := &eth.PayloadAttributes{
		Timestamp:             hexutil.Uint64(futureTimeStamp),
		PrevRandao:            [32]byte{},
		SuggestedFeeRecipient: common.HexToAddress("0x4200000000000000000000000000000000000011"),
		Transactions:          transactions,
		NoTxPool:              false,
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
	pendingTransactionsNum := len(executionRes.Transactions)
	fmt.Println("-> Pending transaction number: ", pendingTransactionsNum)

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

	// Check block number
	newBlock, err := rpcClient.GetBlock()
	if err != nil {
		Exit(err)
	}

	// Check balance
	postDepositAddressBalance, err := rpcClient.GetBalance(depositAddress)
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

			if len(newBlock.Transactions) != pendingTransactionsNum {
				fmt.Println("Invalid new block - block should have 3 transaction")
				os.Exit(1)
			}

			postBalance := depositAddressBalance.Add(depositAddressBalance, mintETHAmount)
			if (postDepositAddressBalance.Cmp(postBalance)) != 0 {
				fmt.Println("Invalid new block - depositAddressBalance should be increased")
				os.Exit(1)
			}

			fmt.Println("-> All checks passed!!!!")
			fmt.Println("-> Successfully Submit Block to L2")
			break
		}
		time.Sleep(time.Second * 1)
	}
}
