package rpc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/NethermindEth/juno/core/felt"
)

type ResultPageRequest struct {
	// a pointer to the last element of the delivered page, use this token in a subsequent query to obtain the next page
	ContinuationToken string `json:"continuation_token,omitempty"`
	ChunkSize         int    `json:"chunk_size"`
}

type StorageEntry struct {
	Key   *felt.Felt `json:"key"`
	Value *felt.Felt `json:"value"`
}

// type StorageEntries struct {
// 	StorageEntry []StorageEntry
// }

// ContractStorageDiffItem is a change in a single storage item
type ContractStorageDiffItem struct {
	// ContractAddress is the contract address for which the state changed
	Address        *felt.Felt     `json:"address"`
	StorageEntries []StorageEntry `json:"storage_entries"`
}

// DeclaredClassesItem is an object with class_hash and compiled_class_hash
type DeclaredClassesItem struct {
	//The hash of the declared class
	ClassHash *felt.Felt `json:"class_hash"`
	//The Cairo assembly hash corresponding to the declared class
	CompiledClassHash *felt.Felt `json:"compiled_class_hash"`
}

// DeployedContractItem A new contract deployed as part of the new state
type DeployedContractItem struct {
	// ContractAddress is the address of the contract
	Address *felt.Felt `json:"address"`
	// ClassHash is the hash of the contract code
	ClassHash *felt.Felt `json:"class_hash"`
}

// contracts whose class was replaced
type ReplacedClassesItem struct {
	//The address of the contract whose class was replaced
	ContractClass *felt.Felt `json:"contract_address"`
	//The new class hash
	ClassHash *felt.Felt `json:"class_hash"`
}

// ContractNonce is a the updated nonce per contract address
type ContractNonce struct {
	// ContractAddress is the address of the contract
	ContractAddress *felt.Felt `json:"contract_address"`
	// Nonce is the nonce for the given address at the end of the block"
	Nonce *felt.Felt `json:"nonce"`
}

// StateDiff is the change in state applied in this block, given as a
// mapping of addresses to the new values and/or new contracts.
type StateDiff struct {
	// list storage changes
	StorageDiffs []ContractStorageDiffItem `json:"storage_diffs"`
	// a list of Deprecated declared classes
	DeprecatedDeclaredClasses []*felt.Felt `json:"deprecated_declared_classes"`
	// list of DeclaredClassesItems objects
	DeclaredClasses []DeclaredClassesItem `json:"declared_classes"`
	// list of new contract deployed as part of the state update
	DeployedContracts []DeployedContractItem `json:"deployed_contracts"`
	// list of contracts whose class was replaced
	ReplacedClasses []ReplacedClassesItem `json:"replaced_classes"`
	// Nonces provides the updated nonces per contract addresses
	Nonces []ContractNonce `json:"nonces"`
}

// STATE_UPDATE in spec
type StateUpdateOutput struct {
	// BlockHash is the block identifier,
	BlockHash *felt.Felt `json:"block_hash"`
	// NewRoot is the new global state root.
	NewRoot *felt.Felt `json:"new_root"`
	// Pending
	PendingStateUpdate
}

// PENDING_STATE_UPDATE in spec
type PendingStateUpdate struct {
	// OldRoot is the previous global state root.
	OldRoot *felt.Felt `json:"old_root"`
	// AcceptedTime is when the block was accepted on L1.
	StateDiff StateDiff `json:"state_diff"`
}

// SyncStatus is An object describing the node synchronization status
type SyncStatus struct {
	SyncStatus        bool       // todo(remove? not in spec)
	StartingBlockHash *felt.Felt `json:"starting_block_hash,omitempty"`
	StartingBlockNum  NumAsHex   `json:"starting_block_num,omitempty"`
	CurrentBlockHash  *felt.Felt `json:"current_block_hash,omitempty"`
	CurrentBlockNum   NumAsHex   `json:"current_block_num,omitempty"`
	HighestBlockHash  *felt.Felt `json:"highest_block_hash,omitempty"`
	HighestBlockNum   NumAsHex   `json:"highest_block_num,omitempty"`
}

func (s SyncStatus) MarshalJSON() ([]byte, error) {
	if !s.SyncStatus {
		return []byte("false"), nil
	}
	output := map[string]interface{}{}
	output["starting_block_hash"] = s.StartingBlockHash
	output["starting_block_num"] = s.StartingBlockNum
	output["current_block_hash"] = s.CurrentBlockHash
	output["current_block_num"] = s.CurrentBlockNum
	output["highest_block_hash"] = s.HighestBlockHash
	output["highest_block_num"] = s.HighestBlockNum
	return json.Marshal(output)
}

func (s *SyncStatus) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, s)

	// if string(data) == "false" {
	// 	s.SyncStatus = false
	// 	return nil
	// }
	// s.SyncStatus = true
	// output := map[string]interface{}{}
	// err := json.Unmarshal(data, &output)
	// if err != nil {
	// 	return err
	// }
	// s.StartingBlockHash = output["starting_block_hash"].(string)
	// s.StartingBlockNum = types.NumAsHex(output["starting_block_num"].(string))
	// s.CurrentBlockHash = output["current_block_hash"].(string)
	// s.CurrentBlockNum = types.NumAsHex(output["current_block_num"].(string))
	// s.HighestBlockHash = output["highest_block_hash"].(string)
	// s.HighestBlockNum = types.NumAsHex(output["highest_block_num"].(string))
	// return nil
}

// AddDeclareTransactionOutput provides the output for AddDeclareTransaction.
type AddDeclareTransactionOutput struct {
	TransactionHash *felt.Felt `json:"transaction_hash"`
	ClassHash       *felt.Felt `json:"class_hash"`
}

// FunctionCall function call information
type FunctionCall struct {
	ContractAddress    *felt.Felt `json:"contract_address"`
	EntryPointSelector *felt.Felt `json:"entry_point_selector,omitempty"`

	// Calldata The parameters passed to the function
	Calldata []*felt.Felt `json:"calldata"`
}

type FeeEstimate struct {
	// GasConsumed the Ethereum gas cost of the transaction (see https://docs.starknet.io/docs/Fees/fee-mechanism for more info)
	GasConsumed NumAsHex `json:"gas_consumed"`

	// GasPrice the gas price (in gwei) that was used in the cost estimation
	GasPrice NumAsHex `json:"gas_price"`

	// OverallFee the estimated fee for the transaction (in gwei), product of gas_consumed and gas_price
	OverallFee NumAsHex `json:"overall_fee"`
}

type TxnExecutionStatus string

const (
	TxnExecutionStatusSUCCEEDED TxnExecutionStatus = "SUCCEEDED"
	TxnExecutionStatusREVERTED  TxnExecutionStatus = "REVERTED"
)

func (ts *TxnExecutionStatus) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	switch unquoted {
	case "SUCCEEDED":
		*ts = TxnExecutionStatusSUCCEEDED
	case "REVERTED":
		*ts = TxnExecutionStatusREVERTED
	default:
		return fmt.Errorf("unsupported status: %s", data)
	}
	return nil
}

func (ts TxnExecutionStatus) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(ts))), nil
}

func (s TxnExecutionStatus) String() string {
	return string(s)
}

type TxnFinalityStatus string

const (
	TxnFinalityStatusAcceptedOnL1 TxnFinalityStatus = "ACCEPTED_ON_L1"
	TxnFinalityStatusAcceptedOnL2 TxnFinalityStatus = "ACCEPTED_ON_L2"
)

func (ts *TxnFinalityStatus) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	switch unquoted {
	case "ACCEPTED_ON_L1":
		*ts = TxnFinalityStatusAcceptedOnL1
	case "ACCEPTED_ON_L2":
		*ts = TxnFinalityStatusAcceptedOnL2
	default:
		return fmt.Errorf("unsupported status: %s", data)
	}
	return nil
}

func (ts TxnFinalityStatus) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(ts))), nil
}

func (s TxnFinalityStatus) String() string {
	return string(s)
}
