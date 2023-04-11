package engineapi

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/beacon"
)

type Data = hexutil.Bytes

type PayloadID = beacon.PayloadID

type ExecutionPayload struct {
	ParentHash    common.Hash    `json:"parentHash"`
	FeeRecipient  common.Address `json:"feeRecipient"`
	StateRoot     common.Hash    `json:"stateRoot"`
	ReceiptsRoot  common.Hash    `json:"receiptsRoot"`
	LogsBloom     hexutil.Bytes  `json:"logsBloom"`
	PrevRandao    common.Hash    `json:"prevRandao"`
	BlockNumber   hexutil.Uint64 `json:"blockNumber"`
	GasLimit      hexutil.Uint64 `json:"gasLimit"`
	GasUsed       hexutil.Uint64 `json:"gasUsed"`
	Timestamp     hexutil.Uint64 `json:"timestamp"`
	ExtraData     hexutil.Bytes  `json:"extraData"`
	BaseFeePerGas *hexutil.Big   `json:"baseFeePerGas"`
	BlockHash     common.Hash    `json:"blockHash"`
	// Array of transaction objects, each object is a byte list (DATA) representing
	// TransactionType || TransactionPayload or LegacyTransaction as defined in EIP-2718
	Transactions []Data `json:"transactions"`
}

type PayloadAttributes struct {
	Timestamp             hexutil.Uint64  `json:"timestamp"`
	PrevRandao            common.Hash     `json:"prevRandao"`
	SuggestedFeeRecipient common.Address  `json:"suggestedFeeRecipient"`
	Transactions          []hexutil.Bytes `json:"transactions"`
	NoTxPool              bool            `json:"noTxPool"`
	GasLimit              *hexutil.Uint64 `json:"gasLimit,omitempty"`
}

type ExecutePayloadStatus string

const (
	ExecutionValid                ExecutePayloadStatus = "VALID"
	ExecutionInvalid              ExecutePayloadStatus = "INVALID"
	ExecutionSyncing              ExecutePayloadStatus = "SYNCING"
	ExecutionAccepted             ExecutePayloadStatus = "ACCEPTED"
	ExecutionInvalidBlockHash     ExecutePayloadStatus = "INVALID_BLOCK_HASH"
	ExecutionInvalidTerminalBlock ExecutePayloadStatus = "INVALID_TERMINAL_BLOCK"
)

type PayloadStatusV1 struct {
	Status          ExecutePayloadStatus `json:"status"`
	LatestValidHash *common.Hash         `json:"latestValidHash,omitempty"`
	ValidationError *string              `json:"validationError,omitempty"`
}

type ForkchoiceState struct {
	HeadBlockHash      common.Hash `json:"headBlockHash"`
	SafeBlockHash      common.Hash `json:"safeBlockHash"`
	FinalizedBlockHash common.Hash `json:"finalizedBlockHash"`
}

type ForkchoiceUpdatedResult struct {
	PayloadStatus PayloadStatusV1 `json:"payloadStatus"`
	PayloadID     *PayloadID      `json:"payloadId"`
}

type BlockID struct {
	Hash   common.Hash `json:"hash"`
	Number uint64      `json:"number"`
}
