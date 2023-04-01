package transaction

import (
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
)

type TransactionBuilder struct {
	RpcClient    *rpc.RpcClient
	RollupConfig *rollup.Config
}

func NewTransactionBuilder(rpcClient *rpc.RpcClient, rollupConfig *rollup.Config) *TransactionBuilder {
	return &TransactionBuilder{
		RpcClient:    rpcClient,
		RollupConfig: rollupConfig,
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
	dep.To = &to
	dep.Value = amount
	dep.Mint = amount
	dep.Gas = 10000000
	dep.Data = []byte{}
	dep.IsSystemTransaction = false

	return &dep, nil
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
	default:
		return nil, fmt.Errorf("Unknown transaction type")
	}
}
