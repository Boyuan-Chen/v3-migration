package transaction

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/Boyuan-Chen/v3-migration/rpc"
	"github.com/Boyuan-Chen/v3-migration/utils"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
)

type TransactionBuilder struct {
	SmartContractViewer *SmartContractViewer
	RpcClient           *rpc.RpcClient
	RollupConfig        *rollup.Config
}

type rpcTransaction struct {
	types.Transaction
	rpcTransactionMeta
	txExtraInfo
}

type rpcTransactionMeta struct {
	L1BlockNumber   *big.Int        `json:"l1BlockNumber"`
	L1Timestamp     uint64          `json:"l1Timestamp"`
	L1Turing        []byte          `json:"l1Turing"`
	L1MessageSender *common.Address `json:"l1MessageSender"`
	QueueOrigin     uint8           `json:"queueOrigin"`
	Index           *uint64         `json:"index"`
	QueueIndex      *uint64         `json:"queueIndex"`
	RawTransaction  []byte          `json:"rawTransaction"`
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func NewTransactionBuilder(smartContractViewer *SmartContractViewer, rpcClient *rpc.RpcClient, rollupConfig *rollup.Config) *TransactionBuilder {
	return &TransactionBuilder{
		SmartContractViewer: smartContractViewer,
		RpcClient:           rpcClient,
		RollupConfig:        rollupConfig,
	}
}

func (t *TransactionBuilder) BuildTestTransaction(key string) (*types.Transaction, error) {
	ecskey, _ := crypto.HexToECDSA(key)
	address := crypto.PubkeyToAddress(ecskey.PublicKey)
	signer := types.NewEIP155Signer(t.RollupConfig.L2ChainID)
	nonce, err := t.RpcClient.GetNextNonce(&address)
	if err != nil {
		return nil, fmt.Errorf("Failed to get nonce: %s", err.Error())
	}
	gasPrice, err := t.RpcClient.GetGasPrice()
	if err != nil {
		return nil, fmt.Errorf("Failed to get gas price: %s", err.Error())
	}
	unsignedTx := types.NewTransaction(
		nonce,
		common.HexToAddress("0x00000000000000000000000000000000deadbeef"),
		new(big.Int),
		5000000,
		gasPrice,
		[]byte{},
	)
	tx, err := types.SignTx(unsignedTx, signer, ecskey)
	if err != nil {
		return nil, fmt.Errorf("Failed to sign transaction: %s", err.Error())
	}
	return tx, nil
}

func (t *TransactionBuilder) BuildTestDepositETHTransaction(from common.Address, to common.Address, amount *big.Int) (*types.DepositTx, error) {
	// SourceHash is random generated
	// The sourceHash of a deposit transaction is computed based on the origin:
	// User-deposited: keccak256(bytes32(uint256(0)), keccak256(l1BlockHash, bytes32(uint256(l1LogIndex)))). Where the l1BlockHash, and l1LogIndex all refer to the inclusion of the deposit log event on L1. l1LogIndex is the index of the deposit event log in the combined list of log events of the block.
	// op-node/rollup/derive/deposit_source.go
	var dep types.DepositTx
	source := derive.UserDepositSource{
		L1BlockHash: utils.MockL1Hash(rand.Uint64()),
		LogIndex:    rand.Uint64(),
	}
	dep.SourceHash = source.SourceHash()
	dep.From = from
	dep.To = &utils.L2MessengerAddress
	dep.Value = amount
	dep.Mint = amount
	dep.Gas = 15000000
	// build payload
	dep.Data = []byte{}
	dep.IsSystemTransaction = false

	return &dep, nil
}

func (t *TransactionBuilder) BuildTestDepositBOBATransaction(from common.Address, to common.Address, amount *big.Int) (*types.DepositTx, error) {
	// SourceHash is random generated
	// The sourceHash of a deposit transaction is computed based on the origin:
	// User-deposited: keccak256(bytes32(uint256(0)), keccak256(l1BlockHash, bytes32(uint256(l1LogIndex)))). Where the l1BlockHash, and l1LogIndex all refer to the inclusion of the deposit log event on L1. l1LogIndex is the index of the deposit event log in the combined list of log events of the block.
	// op-node/rollup/derive/deposit_source.go
	var dep types.DepositTx
	source := derive.UserDepositSource{
		L1BlockHash: utils.MockL1Hash(rand.Uint64()),
		LogIndex:    rand.Uint64(),
	}

	dep.SourceHash = source.SourceHash()
	dep.From = utils.ApplyL1ToL2Alias(utils.L1MessengerAddress)
	dep.To = &to
	dep.Value = big.NewInt(0)
	dep.Mint = nil
	dep.Gas = 15000000
	dep.IsSystemTransaction = false

	nonce, err := t.SmartContractViewer.GetNonceFromL1Messenger()
	if err != nil {
		return nil, fmt.Errorf("Failed to get nonce from L1 Messenger: %s", err.Error())
	}

	/** CAREFUL: This is a hack to make sure the nonce is unique on L2 **/
	// Since nonce doesn't change on L1, we need to add a random number to it
	// to make sure the nonce is unique on L2
	nonce = nonce.Add(nonce, big.NewInt(int64(rand.Uint64())))

	data, err := BuildBobaDepositFromL1ToL2(&from, amount, nonce)
	if err != nil {
		return nil, fmt.Errorf("Failed to build Boba payload: %s", err.Error())
	}
	dep.Data = data
	return &dep, nil
}

func (t *TransactionBuilder) GetPastTransaction() (*types.Transaction, error) {
	client, err := ethRpc.Dial("https://goerli.boba.network")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RPC: %s", err.Error())
	}

	var r *rpcTransaction
	err = client.CallContext(context.Background(), &r, "eth_getTransactionByHash", "0x75eda4a9208188f16824d28a3615baff87c5a3ad83dc433f66a29e5e4558722a")
	if err != nil {
		return nil, fmt.Errorf("Failed to get transaction: %s", err.Error())
	}

	return &r.Transaction, nil
}

// func (t *TransactionBuilder) BuildSystemTransaction() {
// 	var seqNum uint64 = 1
// 	l1InfoTx, err := derive.L1InfoDepositBytes(seqNum, l1Info, sysConfig)
// }

func (t *TransactionBuilder) SubmitTransaction(key string) error {
	fmt.Println("Building and Submitting Test Transaction...")
	tx, err := t.BuildTestTransaction(key)
	if err != nil {
		return fmt.Errorf("Failed to build transaction: %s", err.Error())
	}
	err = t.RpcClient.SendRawTransaction(tx)
	if err != nil {
		return fmt.Errorf("Failed to send transaction: %s", err.Error())
	}
	fmt.Println("-> Test Transaction Submitted")
	fmt.Println("--------------------------------------------")
	return nil
}

func (t *TransactionBuilder) MarshalBinary(tx interface{}) ([]byte, error) {
	switch tx.(type) {
	case *types.Transaction:
		binaryTx, err := tx.(*types.Transaction).MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal transaction: %s", err.Error())
		}
		return binaryTx, nil
	case *types.DepositTx:
		binaryTx, err := types.NewTx(tx.(*types.DepositTx)).MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal deposit transaction: %s", err.Error())
		}
		return binaryTx, nil
	case *rpc.LegacyTransaction:
		binaryTx, err := tx.(*rpc.LegacyTransaction).MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal legacy transaction: %s", err.Error())
		}
		fmt.Println("Got legacy transaction: ", binaryTx)
		return binaryTx, nil
	default:
		return nil, fmt.Errorf("Unknown transaction type")
	}
}
