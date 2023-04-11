package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/engineapi"
	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/Boyuan-Chen/v3-migration/transaction"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	l2PublicEndpoint = "http://localhost:9545"
	l2Endpoint       = "http://localhost:8551"
	l2LegacyEndpoint = "https://replica.goerli.boba.network"

	secretConfigPath = "./static/test-jwt-secret.txt"
	rollupConfigPath = "./static/rollup.json"
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
	rpcClient, err := rpc.NewRpcClient(l2Endpoint, *secret)
	if err != nil {
		Exit(err)
	}
	l2EthClient, err := ethclient.Dial(l2PublicEndpoint)
	if err != nil {
		Exit(err)
	}

	smartContractViewer := transaction.NewSmartContractViewer(nil, l2EthClient)
	transactionBuilder := transaction.NewTransactionBuilder(smartContractViewer, rpcClient, rollupConfig)
	fmt.Println("Starting to submit transaction...", transactionBuilder)

	// Get latest block
	latestBlock, err := rpcClient.GetLatestBlock()
	if err != nil {
		Exit(err)
	}
	nextBlockNumber := uint64(latestBlock.Number) + 1

	legacyRpcClient, err := rpc.NewRpcClient(l2LegacyEndpoint, *secret)
	legacyBlock, err := legacyRpcClient.GetLegacyBlock(big.NewInt(int64(nextBlockNumber)))
	if err != nil {
		Exit(err)
	}
	gasLimit := legacyBlock.GasLimit
	txHash := legacyBlock.Transactions[0].Hash()
	legacyTransaction, err := legacyRpcClient.GetLegacyTransaction(txHash)

	// fmt.Println("-> Legacy transaction FROM: ", legacyTransaction.GetSender())

	// Verify that legacy transaction has the same txHash
	if legacyTransaction.Hash() != txHash {
		Exit(fmt.Errorf("legacy transaction hash does not match"))
	}

	// Build binary legacy transaction
	binaryLegacyTx, err := transactionBuilder.MarshalBinary(legacyTransaction)
	if err != nil {
		Exit(err)
	}
	transactions := make([]eth.Data, 1)
	transactions[0] = binaryLegacyTx

	// Create engine
	engine, err := engineapi.NewEngineAPI(rpcClient, rollupConfig)
	if err != nil {
		Exit(err)
	}

	// Step 1: Get payloadID
	// engine_forkchoiceUpdatedV1 -> Get payloadID
	fc := &eth.ForkchoiceState{
		HeadBlockHash:      latestBlock.Hash,
		SafeBlockHash:      latestBlock.Hash,
		FinalizedBlockHash: latestBlock.Hash,
	}
	attributes := &eth.PayloadAttributes{
		Timestamp:             hexutil.Uint64(legacyBlock.Time),
		PrevRandao:            [32]byte{},
		SuggestedFeeRecipient: common.HexToAddress("0x4200000000000000000000000000000000000011"),
		Transactions:          transactions,
		NoTxPool:              false,
		GasLimit:              (*eth.Uint64Quantity)(&gasLimit),
	}

	// engine_forkchoiceUpdatedV1
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
	var txType types.Transaction
	fmt.Println("-> Execution payload: ", executionRes)
	pendingTransactionsNum := len(executionRes.Transactions)
	fmt.Println("-> Pending transaction number: ", pendingTransactionsNum)
	// fmt.Println("-> Pending transaction hash: ", executionRes.Transactions[0])
	err = txType.UnmarshalBinary(executionRes.Transactions[0])
	if err != nil {
		Exit(fmt.Errorf("failed to unmarshal transaction: %w", err))
	}
	fmt.Println("-> Pending transaction type: ", txType.Hash().Hex())
	if txType.Hash() != txHash {
		fmt.Println("-> Pending transaction hash is correct")
	}
	fmt.Println("-> Pending block hash: ", executionRes.BlockHash)
	fmt.Println("-> Pending StateRoot: ", executionRes.StateRoot)
	fmt.Println("-> Pending FeeRecipient: ", executionRes.FeeRecipient)
	fmt.Println("-> Pending BlockNumber: ", executionRes.BlockNumber)
	fmt.Println("-> Pending GasLimit: ", executionRes.GasLimit)
	fmt.Println("-> Pending GasUsed: ", executionRes.GasUsed)
	fmt.Println("-> Pending Timestamp: ", executionRes.Timestamp)
	fmt.Println("-> Pending Difficulty: ", executionRes.ExtraData)
	fmt.Println("-> Pending ExtraData: ", executionRes.ExtraData)

	// Step 3: Execute payload
	// engine_newPayloadV1 -> Execute payload
	res, err := engine.ExecutePayload(executionRes)
	if err != nil {
		Exit(err)
	}

	fmt.Println("-> Execution result: ", res)

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
}
