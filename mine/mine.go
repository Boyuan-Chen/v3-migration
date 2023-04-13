package mine

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
)

type Miner struct {
	l2PublicRpc  *rpc.RpcClient
	l2LegacyRpc  *rpc.RpcClient
	l2PrivateRpc *engineapi.EngineAPI
	config       *config.Config
}

func NewMiner(l2PublicRpc *rpc.RpcClient, l2LegacyRpc *rpc.RpcClient, l2PrivateRpc *engineapi.EngineAPI, cfg *config.Config) *Miner {
	return &Miner{
		l2PublicRpc:  l2PublicRpc,
		l2LegacyRpc:  l2LegacyRpc,
		l2PrivateRpc: l2PrivateRpc,
		config:       cfg,
	}
}

func (m *Miner) MineBlock() error {
	for {
		// Get latest block
		latestBlock, err := m.l2PublicRpc.GetLatestBlock()
		if err != nil {
			return err
		}
		nextBlockNumber := uint64(latestBlock.Number) + 1

		legacyBlock, err := m.l2LegacyRpc.GetLegacyBlock(big.NewInt(int64(nextBlockNumber)))
		if err != nil {
			return err
		}
		gasLimit := legacyBlock.GasLimit
		txHash := legacyBlock.Transactions[0].Hash()
		legacyTransaction, err := m.l2LegacyRpc.GetLegacyTransaction(txHash)

		// Verify that legacy transaction has the same txHash
		if legacyTransaction.Hash() != txHash {
			return fmt.Errorf("legacy transaction hash does not match")
		}

		// Build binary legacy transaction
		binaryLegacyTx, err := transaction.MarshalBinary(legacyTransaction)
		if err != nil {
			return err
		}
		transactions := make([]engineapi.Data, 1)
		transactions[0] = binaryLegacyTx

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
		fcUpdateRes, err := m.l2PrivateRpc.ForkchoiceUpdate(fc, attributes)
		if err != nil {
			return err
		}

		// Step 2: Get executionPayload
		// engine_getPayloadV1 -> Get executionPayload
		executionRes, err := m.l2PrivateRpc.GetPayload(fcUpdateRes.PayloadID)
		if err != nil {
			return err
		}
		var txType types.Transaction
		if len(executionRes.Transactions) != 1 {
			log.Warn("Pending transaction length is not 1")
			return fmt.Errorf("pending transaction length is not 1")
		}
		err = txType.UnmarshalBinary(executionRes.Transactions[0])
		if err != nil {
			return fmt.Errorf("failed to unmarshal transaction: %w", err)
		}
		if txType.Hash() != txHash {
			log.Warn("Pending transaction hash is not correct", "pending", txType.Hash(), "latest", txHash)
			return fmt.Errorf("pending transaction hash is not correct")
		}
		if executionRes.BlockHash != legacyBlock.Hash {
			log.Warn("Pending block hash is not correct", "pending", executionRes.BlockHash, "latest", legacyBlock.Hash)
			return fmt.Errorf("pending block hash is not correct")
		}

		log.Info("Execution block", "blockNumber", uint64(executionRes.BlockNumber))

		// Step 3: Execute payload
		// engine_newPayloadV1 -> Execute payload
		res, err := m.l2PrivateRpc.ExecutePayload(executionRes)
		if err != nil {
			return err
		}
		if res.Status != "VALID" {
			log.Warn("Payload is invalid", "status", res.Status)
			return fmt.Errorf("payload is invalid")

		}
		if *res.LatestValidHash != executionRes.BlockHash {
			log.Warn("Latest valid hash is not correct", "pending", executionRes.BlockHash, "latest", res.LatestValidHash)
			return fmt.Errorf("Latest valid hash is not correct")
		}

		// Step 4: Submit block
		// engine_executePayloadV1 -> Submit block
		newfc := &engineapi.ForkchoiceState{
			HeadBlockHash:      executionRes.BlockHash,
			SafeBlockHash:      executionRes.BlockHash,
			FinalizedBlockHash: executionRes.BlockHash,
		}
		finalRes, err := m.l2PrivateRpc.ForkchoiceUpdate(newfc, nil)
		if err != nil {
			return err
		}
		if finalRes.PayloadStatus.Status != "VALID" {
			log.Warn("Payload is invalid", "status", finalRes.PayloadStatus.Status)
			return fmt.Errorf("payload is invalid")
		}

		// verify some information before going on
		latestBlock, err = m.l2PublicRpc.GetLatestBlock()
		if err != nil {
			log.Warn("Failed to get latest block", "error", err)

		}
		if latestBlock.Root != legacyBlock.Root {
			log.Warn("Block root is not correct", "pending", legacyBlock.Root, "latest", latestBlock.Root)
			return fmt.Errorf("Block root is not correct")
		}
		if latestBlock.ReceiptHash != legacyBlock.ReceiptHash {
			log.Warn("Receipt hash is not correct", "pending", legacyBlock.ReceiptHash, "latest", latestBlock.ReceiptHash)
			return fmt.Errorf("Receipt hash is not correct")
		}

		log.Info("Block mined", "blockNumber", uint64(executionRes.BlockNumber))
		time.Sleep(1 * time.Second)
		log.Info("Waiting for next block to be mined", "blockNumber", uint64(executionRes.BlockNumber+1))
		return nil
	}
}
