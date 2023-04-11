package engineapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/ethereum/go-ethereum/log"
)

type EngineAPI struct {
	Engine client.RPC
	logger log.Logger
}

func NewEngineAPI(rpcClient *rpc.RpcClient, logger log.Logger) (*EngineAPI, error) {
	if rpcClient == nil {
		return nil, fmt.Errorf("Failed to create NewEngineAPI, rpcClient is nil")
	}
	return &EngineAPI{
		Engine: rpcClient.Client,
		logger: logger,
	}, nil
}

func (e *EngineAPI) ForkchoiceUpdate(fc *ForkchoiceState, attributes *PayloadAttributes) (*ForkchoiceUpdatedResult, error) {
	// e.logger.Info("ForkchoiceUpdate... (engine_forkchoiceUpdatedV1)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var result ForkchoiceUpdatedResult
	if err := e.Engine.CallContext(ctx, &result, "engine_forkchoiceUpdatedV1", fc, attributes); err != nil {
		return nil, fmt.Errorf("Failed to obtain new payloadId: %v", err)
	}
	e.logger.Info("ForkchoiceUpdate Success", "PayloadStatus", result.PayloadStatus.Status, "LatestValidHash", result.PayloadStatus.LatestValidHash)
	return &result, nil
}

func (e *EngineAPI) GetPayload(payloadID *beacon.PayloadID) (*ExecutionPayload, error) {
	// e.logger.Info("GetPayload... (engine_getPayloadV1)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var result ExecutionPayload
	if err := e.Engine.CallContext(ctx, &result, "engine_getPayloadV1", payloadID); err != nil {
		return nil, fmt.Errorf("Failed to obtain new payloadId: %v", err)
	}
	e.logger.Info("GetPayload Success", "PayloadID", payloadID, "BlockHash", result.BlockHash)
	return &result, nil
}

func (e *EngineAPI) ExecutePayload(executionPayload *ExecutionPayload) (*PayloadStatusV1, error) {
	// e.logger.Info("ExecutePayload... (engine_newPayloadV1)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var result PayloadStatusV1
	if err := e.Engine.CallContext(ctx, &result, "engine_newPayloadV1", executionPayload); err != nil {
		return nil, fmt.Errorf("Failed to execute new payloadId: %v", err)
	}
	e.logger.Info("ExecutePayload Result", "PayloadStatus", result.Status, "LatestValidHash", result.LatestValidHash)
	return &result, nil
}
