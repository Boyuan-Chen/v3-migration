package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/Boyuan-Chen/v3-migration/config"
	"github.com/Boyuan-Chen/v3-migration/engineapi"
	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/Boyuan-Chen/v3-migration/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

var (
	l2PublicEndpoint = "http://localhost:9545"
	l2Endpoint       = "http://localhost:8551"
	l2LegacyEndpoint = "https://replica.goerli.boba.network"

	secretConfigPath = "./static/test-jwt-secret.txt"
	rollupConfigPath = "./static/rollup.json"
)

type ExecutionReport struct {
	Error   error
	Success bool
}

func mineBlock(rpcClient *rpc.RpcClient, secret *[32]byte, logger log.Logger, executionReport chan ExecutionReport) {
	for {
		// Get latest block
		latestBlock, err := rpcClient.GetLatestBlock()
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		nextBlockNumber := uint64(latestBlock.Number) + 1

		legacyRpcClient, err := rpc.NewRpcClient(l2LegacyEndpoint, *secret)
		legacyBlock, err := legacyRpcClient.GetLegacyBlock(big.NewInt(int64(nextBlockNumber)))
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		gasLimit := legacyBlock.GasLimit
		txHash := legacyBlock.Transactions[0].Hash()
		legacyTransaction, err := legacyRpcClient.GetLegacyTransaction(txHash)

		// Verify that legacy transaction has the same txHash
		if legacyTransaction.Hash() != txHash {
			executionReport <- ExecutionReport{Error: fmt.Errorf("legacy transaction hash does not match"), Success: false}
		}

		// Build binary legacy transaction
		binaryLegacyTx, err := transaction.MarshalBinary(legacyTransaction)
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		transactions := make([]engineapi.Data, 1)
		transactions[0] = binaryLegacyTx

		// Create engine
		engine, err := engineapi.NewEngineAPI(rpcClient, logger)
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
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
			NoTxPool:              true,
			GasLimit:              &gasLimit,
		}

		// engine_forkchoiceUpdatedV1
		fcUpdateRes, err := engine.ForkchoiceUpdate(fc, attributes)
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}

		// Step 2: Get executionPayload
		// engine_getPayloadV1 -> Get executionPayload
		executionRes, err := engine.GetPayload(fcUpdateRes.PayloadID)
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		var txType types.Transaction
		if len(executionRes.Transactions) != 1 {
			logger.Warn("Pending transaction length is not 1")
			executionReport <- ExecutionReport{Error: fmt.Errorf("pending transaction length is not 1"), Success: false}
		}
		err = txType.UnmarshalBinary(executionRes.Transactions[0])
		if err != nil {
			executionReport <- ExecutionReport{Error: fmt.Errorf("failed to unmarshal transaction: %w", err), Success: false}
		}
		if txType.Hash() != txHash {
			logger.Warn("Pending transaction hash is not correct", "pending", txType.Hash(), "latest", txHash)
			executionReport <- ExecutionReport{Error: fmt.Errorf("pending transaction hash is not correct"), Success: false}
		}
		if executionRes.BlockHash != legacyBlock.Hash {
			logger.Warn("Pending block hash is not correct", "pending", executionRes.BlockHash, "latest", legacyBlock.Hash)
			executionReport <- ExecutionReport{Error: fmt.Errorf("pending block hash is not correct"), Success: false}
		}

		logger.Info("Execution block", "blockNumber", uint64(executionRes.BlockNumber))

		// Step 3: Execute payload
		// engine_newPayloadV1 -> Execute payload
		res, err := engine.ExecutePayload(executionRes)
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		if res.Status != "VALID" {
			logger.Warn("Payload is invalid", "status", res.Status)
			executionReport <- ExecutionReport{Error: fmt.Errorf("payload is invalid"), Success: false}
		}
		if *res.LatestValidHash != executionRes.BlockHash {
			logger.Warn("Latest valid hash is not correct", "pending", executionRes.BlockHash, "latest", res.LatestValidHash)
			executionReport <- ExecutionReport{Error: fmt.Errorf("Latest valid hash is not correct"), Success: false}
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
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		if finalRes.PayloadStatus.Status != "VALID" {
			logger.Warn("Payload is invalid", "status", finalRes.PayloadStatus.Status)
			executionReport <- ExecutionReport{Error: fmt.Errorf("payload is invalid"), Success: false}
		}

		// verify some information before going on
		latestBlock, err = rpcClient.GetLatestBlock()
		if err != nil {
			executionReport <- ExecutionReport{Error: err, Success: false}
		}
		if latestBlock.Root != legacyBlock.Root {
			logger.Warn("Block root is not correct", "pending", legacyBlock.Root, "latest", latestBlock.Root)
			executionReport <- ExecutionReport{Error: fmt.Errorf("Block root is not correct"), Success: false}
		}
		if latestBlock.ReceiptHash != legacyBlock.ReceiptHash {
			logger.Warn("Receipt hash is not correct", "pending", legacyBlock.ReceiptHash, "latest", latestBlock.ReceiptHash)
			executionReport <- ExecutionReport{Error: fmt.Errorf("Receipt hash not correct"), Success: false}
		}

		logger.Info("Block mined", "blockNumber", uint64(executionRes.BlockNumber))
		time.Sleep(1 * time.Second)
		logger.Info("Waiting for next block to be mined", "blockNumber", uint64(executionRes.BlockNumber+1))
		executionReport <- ExecutionReport{Error: nil, Success: true}
	}
}

func main() {

	executionReport := make(chan ExecutionReport, 1)

	// logger
	logger := log.New()

	// Create config
	config := config.NewConfig(secretConfigPath, rollupConfigPath)

	// Get rollup config
	_, secret, _ := config.GetConfig()

	// create Rpc client
	rpcClient, _ := rpc.NewRpcClient(l2Endpoint, *secret)

	go mineBlock(rpcClient, secret, logger, executionReport)

	for {
		select {
		case executionReport := <-executionReport:
			if !executionReport.Success {
				logger.Error("Failed to execute payload", "error", executionReport.Error)
			}
		}
	}
}
