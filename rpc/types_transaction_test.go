package rpc

import (
	"encoding/json"
	"testing"

	"github.com/sjxqqq/starknet-go/utils"
	"github.com/test-go/testify/require"
)

func TestTransaction(t *testing.T) {
	f := utils.TestHexToFelt(t, "0xdead")
	th := TransactionHash{f}
	b, err := json.Marshal(th)
	if err != nil {
		t.Fatalf("marshalling transaction hash: %v", err)
	}

	marshalled, err := f.MarshalJSON()
	if err != nil {
		t.Fatalf("marshalling transaction hash: %v", err)
	}

	require.Equal(t, b, marshalled)

}
