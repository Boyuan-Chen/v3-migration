package engineapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/ethereum/go-ethereum/log"
)

type EngineAPI struct {
	Engine sources.EngineClient
}

func NewEngineAPI(rpcClient *rpc.RpcClient, rollupConfig *rollup.Config) (*EngineAPI, error) {
	if rpcClient == nil {
		return nil, fmt.Errorf("Failed to create NewEngineAPI, rpcClient is nil")
	}
	logger := log.New("hash")
	engineAPI, err := sources.NewEngineClient(
		rpcClient.Client,
		logger,
		nil,
		sources.EngineClientDefaultConfig(rollupConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create NewEngineAPI: %s", err.Error())
	}
	return &EngineAPI{
		Engine: *engineAPI,
	}, nil
}

func (e *EngineAPI) ForkchoiceUpdate(fc *eth.ForkchoiceState, attributes *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	fmt.Println("-> Getting New PayloadId... (ForkchoiceUpdate)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := e.Engine.ForkchoiceUpdate(ctx, fc, attributes)
	if err != nil {
		return nil, fmt.Errorf("Failed to update forkchoice: %s", err.Error())
	}
	fmt.Println("-> Forkchoice Updated")
	fmt.Println("-> New PayloadId: ", res.PayloadID)
	fmt.Println("--------------------------------------------")
	return res, nil
}

func (e *EngineAPI) GetPayload(payloadID *beacon.PayloadID) (*eth.ExecutionPayload, error) {
	fmt.Println("-> Getting New Payload... (GetPayload)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := e.Engine.GetPayload(ctx, *payloadID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get payload: %s", err.Error())
	}
	fmt.Println("-> GetPayload Success")
	fmt.Println("-> New Block Hash: ", res.BlockHash)
	fmt.Println("-> New Block Transaction: ", res.Transactions)
	fmt.Println("--------------------------------------------")
	return res, nil
}

func (e *EngineAPI) ExecutePayload(executionPayload *eth.ExecutionPayload) (*eth.PayloadStatusV1, error) {
	fmt.Println("-> Executing New Block... (ExecutePayload)")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res, err := e.Engine.NewPayload(ctx, executionPayload)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute payload: %s", err.Error())
	}
	fmt.Println("-> ExecutePayload Success")
	fmt.Println("-> New Latest Valid Hash: ", res.LatestValidHash)
	fmt.Println("--------------------------------------------")
	return res, nil
}
