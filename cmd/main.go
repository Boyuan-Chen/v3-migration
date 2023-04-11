package main

import (
	"fmt"
	"math/big"
	"os"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/engineapi"
	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/Boyuan-Chen/v3-migration/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
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

	// logger
	logger := log.New()

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

	// Verify that legacy transaction has the same txHash
	if legacyTransaction.Hash() != txHash {
		Exit(fmt.Errorf("legacy transaction hash does not match"))
	}

	// Build binary legacy transaction
	binaryLegacyTx, err := transactionBuilder.MarshalBinary(legacyTransaction)
	if err != nil {
		Exit(err)
	}
	transactions := make([]engineapi.Data, 1)
	transactions[0] = binaryLegacyTx

	// Create engine
	engine, err := engineapi.NewEngineAPI(rpcClient, logger)
	if err != nil {
		Exit(err)
	}

	// Step 1: Get payloadID
	// engine_forkchoiceUpdatedV1 -> Get payloadID
	fc := &engineapi.ForkchoiceState{
		HeadBlockHash:      latestBlock.Hash,
		SafeBlockHash:      latestBlock.Hash,
		FinalizedBlockHash: latestBlock.Hash,
	}
	attributes := &engineapi.PayloadAttributes{
		Timestamp:             hexutil.Uint64(legacyBlock.Time),
		PrevRandao:            [32]byte{},
		SuggestedFeeRecipient: common.HexToAddress("0x4200000000000000000000000000000000000011"),
		Transactions:          transactions,
		NoTxPool:              false,
		GasLimit:              &gasLimit,
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
	if len(executionRes.Transactions) != 1 {
		logger.Warn("Pending transaction length is not 1")
		Exit(fmt.Errorf("pending transaction length is not 1"))
	}
	err = txType.UnmarshalBinary(executionRes.Transactions[0])
	if err != nil {
		Exit(fmt.Errorf("failed to unmarshal transaction: %w", err))
	}
	if txType.Hash() != txHash {
		logger.Warn("Pending transaction hash is not correct", "pending", txType.Hash(), "latest", txHash)
		Exit(fmt.Errorf("Pending transaction hash is correct"))
	}
	if executionRes.BlockHash != legacyBlock.Hash {
		logger.Warn("Pending block hash is not correct", "pending", executionRes.BlockHash, "latest", legacyBlock.Hash)
		Exit(fmt.Errorf("Pending block hash is correct"))
	}

	logger.Info("Execution block", "blockNumber", executionRes.BlockNumber)

	// Step 3: Execute payload
	// engine_newPayloadV1 -> Execute payload
	res, err := engine.ExecutePayload(executionRes)
	if err != nil {
		Exit(err)
	}
	if res.Status != "VALID" {
		logger.Warn("Payload is invalid", "status", res.Status)
		Exit(fmt.Errorf("payload is invalid"))
	}
	if *res.LatestValidHash != executionRes.BlockHash {
		logger.Warn("Latest valid hash is not correct", "pending", executionRes.BlockHash, "latest", res.LatestValidHash)
		Exit(fmt.Errorf("Latest valid hash is not correct"))
	}

	// Step 4: Submit block
	// engine_executePayloadV1 -> Submit block
	newfc := &engineapi.ForkchoiceState{
		HeadBlockHash:      executionRes.BlockHash,
		SafeBlockHash:      executionRes.BlockHash,
		FinalizedBlockHash: executionRes.BlockHash,
	}
	finalRes, err := engine.ForkchoiceUpdate(newfc, nil)
	if err != nil {
		Exit(err)
	}
	if finalRes.PayloadStatus.Status != "VALID" {
		logger.Warn("Payload is invalid", "status", finalRes.PayloadStatus.Status)
		Exit(fmt.Errorf("payload is invalid"))
	}

	logger.Info("Block submitted", "blockNumber", executionRes.BlockNumber)
}
