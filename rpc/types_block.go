package rpc

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/NethermindEth/juno/core/felt"
)

var ErrInvalidBlockID = errors.New("invalid blockid")

// BlockHashAndNumberOutput is a struct that is returned by BlockHashAndNumber.
type BlockHashAndNumberOutput struct {
	BlockNumber uint64     `json:"block_number,omitempty"`
	BlockHash   *felt.Felt `json:"block_hash,omitempty"`
}

// BlockID is a struct that is used to choose between different
// search types.
type BlockID struct {
	Number *uint64    `json:"block_number,omitempty"`
	Hash   *felt.Felt `json:"block_hash,omitempty"`
	Tag    string     `json:"block_tag,omitempty"`
}

func (b BlockID) MarshalJSON() ([]byte, error) {
	if b.Tag == "pending" || b.Tag == "latest" {
		return []byte(strconv.Quote(b.Tag)), nil
	}

	if b.Tag != "" {
		return nil, ErrInvalidBlockID
	}

	if b.Number != nil {
		return []byte(fmt.Sprintf(`{"block_number":%d}`, *b.Number)), nil
	}

	if b.Hash.BigInt(big.NewInt(0)).BitLen() != 0 {
		return []byte(fmt.Sprintf(`{"block_hash":"%s"}`, b.Hash.String())), nil
	}

	return nil, ErrInvalidBlockID

}

type BlockStatus string

const (
	BlockStatus_Pending      BlockStatus = "PENDING"
	BlockStatus_AcceptedOnL2 BlockStatus = "ACCEPTED_ON_L2"
	BlockStatus_AcceptedOnL1 BlockStatus = "ACCEPTED_ON_L1"
	BlockStatus_Rejected     BlockStatus = "REJECTED"
)

func (bs *BlockStatus) UnmarshalJSON(data []byte) error {
	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	switch unquoted {
	case "PENDING":
		*bs = BlockStatus_Pending
	case "ACCEPTED_ON_L2":
		*bs = BlockStatus_AcceptedOnL2
	case "ACCEPTED_ON_L1":
		*bs = BlockStatus_AcceptedOnL1
	case "REJECTED":
		*bs = BlockStatus_Rejected
	default:
		return fmt.Errorf("unsupported status: %s", data)
	}

	return nil
}

func (bs BlockStatus) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(string(bs))), nil
}

type Block struct {
	BlockHeader
	Status BlockStatus `json:"status"`
	// Transactions The hashes of the transactions included in this block
	Transactions Transactions `json:"transactions"`
}

type PendingBlock struct {
	// ParentHash The hash of this block's parent
	ParentHash *felt.Felt `json:"parent_hash"`
	// Timestamp the time in which the block was created, encoded in Unix time
	Timestamp uint64 `json:"timestamp"`
	// SequencerAddress the StarkNet identity of the sequencer submitting this block
	SequencerAddress *felt.Felt `json:"sequencer_address"`
	// Transactions The hashes of the transactions included in this block
	Transactions Transactions `json:"transactions"`
}

type BlockHeader struct {
	// BlockHash The hash of this block
	BlockHash *felt.Felt `json:"block_hash"`
	// ParentHash The hash of this block's parent
	ParentHash *felt.Felt `json:"parent_hash"`
	// BlockNumber the block number (its height)
	BlockNumber uint64 `json:"block_number"`
	// NewRoot The new global state root
	NewRoot *felt.Felt `json:"new_root"`
	// Timestamp the time in which the block was created, encoded in Unix time
	Timestamp uint64 `json:"timestamp"`
	// SequencerAddress the StarkNet identity of the sequencer submitting this block
	SequencerAddress *felt.Felt `json:"sequencer_address"`
}
