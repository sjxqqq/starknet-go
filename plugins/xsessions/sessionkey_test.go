package xsessions

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	_ "embed"

	"github.com/NethermindEth/juno/core/felt"
	starknetgo "github.com/NethermindEth/starknet.go"
	"github.com/NethermindEth/starknet.go/artifacts"
	"github.com/NethermindEth/starknet.go/types"
	"github.com/NethermindEth/starknet.go/utils"
)

var sessionPluginCompiled = artifacts.PluginV0Compiled

func sessionToken(privateKey, accountAddress, sessionPublicKey string) *SessionKeyToken {
	token, _ := SignToken(
		privateKey,
		types.UTF8StrToBig("SN_GOERLI").Text(16),
		sessionPublicKey,
		accountAddress,
		2*time.Hour,
		[]Policy{{ContractAddress: counterAddress, Selector: "increment"}},
	)
	return token
}

// TestSessionKey_RegisterPlugin
func TestSessionKey_RegisterPlugin(t *testing.T) {
	pluginHash := RegisterClass(t, sessionPluginCompiled)
	v := &accountPlugin{
		PluginHash: pluginHash,
	}
	err := v.Write(".sessionkey.json")
	if err != nil {
		t.Fatal("should be able to save pluginHash, instead:", err)
	}
}

// TestSessionKey_DeployAccount
func TestSessionKey_DeployAccount(t *testing.T) {
	pk, ok := big.NewInt(0).SetString(privateKey, 0)
	if !ok {
		t.Fatal("could not match *big.Int private key with current value")
	}
	publicKey, _, err := starknetgo.Curve.PrivateToPoint(pk)
	if err != nil {
		t.Fatal(err)
	}
	publicKeyString := fmt.Sprintf("0x%s", publicKey.Text(16))
	v := &accountPlugin{}
	err = v.Read(".sessionkey.json")
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{
		publicKeyString,
		v.PluginHash,
	}
	accountAddress := DeployContract(t, accountCompiled, inputs)
	v.AccountAddress = accountAddress
	err = v.Write(".sessionkey.json")
	if err != nil {
		t.Fatal(err)
	}
}

// TestSessionKey_MintEth
func TestSessionKey_MintEth(t *testing.T) {
	v := &accountPlugin{}
	err := v.Read(".sessionkey.json")
	if err != nil {
		t.Fatal(err)
	}
	MintEth(t, v.AccountAddress)
}

// TestSessionKey_CheckEth
func TestSessionKey_CheckEth(t *testing.T) {
	v := &accountPlugin{}
	err := v.Read(".sessionkey.json")
	if err != nil {
		t.Fatal(err)
	}
	CheckEth(t, v.AccountAddress)
}

// IncrementWithSessionKeyPlugin
func IncrementWithSessionKeyPlugin(t *testing.T, accountAddress string, pluginClass string, token *SessionKeyToken, counterAddress string) {
	provider := beforeEachRPC(t)
	// shim a keystore into existing tests.
	// use a string representation of the PK as a fake sender address for the keystore
	ks := starknetgo.NewMemKeystore()

	fakeSenderAddress := sessionPrivateKey
	k := types.SNValToBN(sessionPrivateKey)
	ks.Put(fakeSenderAddress, k)
	account, err := starknetgo.NewRPCAccount(
		utils.TestHexToFelt(t, fakeSenderAddress),
		utils.TestHexToFelt(t, accountAddress),
		ks,
		provider,
		WithSessionKeyPlugin(
			pluginClass,
			token,
		))
	if err != nil {
		t.Fatal("deploy should succeed, instead:", err)
	}
	calls := []types.FunctionCall{
		{
			ContractAddress:    utils.TestHexToFelt(t, counterAddress),
			EntryPointSelector: types.GetSelectorFromNameFelt("increment"),
			Calldata:           []*felt.Felt{},
		},
	}
	ctx := context.Background()
	tx, err := account.Execute(ctx, calls, types.ExecuteDetails{})
	if err != nil {
		t.Fatal("execute should succeed, instead:", err)
	}
	if !strings.HasPrefix(tx.TransactionHash.String(), "0x") {
		t.Fatal("execute should return transaction hash, instead:", tx.TransactionHash)
	}
	status, err := provider.WaitForTransaction(ctx, tx.TransactionHash, 8*time.Second)
	if err != nil {
		t.Fatal("declare should succeed, instead:", err)
	}
	if status != types.TransactionAcceptedOnL2 {
		t.Log("unexpected status transaction status, check:", status)
		t.Log("...")
		t.Log("   verify transaction")
		t.Log("...")
		t.Log("export STARKNET_WALLET=starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount")
		t.Log("export STARKNET_NETWORK=alpha-goerli")
		t.Logf("export HASH=%s\n", tx.TransactionHash)
		t.Log("starknet get_transaction --hash $HASH --feeder_gateway http://localhost:5050/feeder_gateway")
		t.Log("...")
		t.Fail()
	}
	fmt.Printf("tx hash: %s\n", tx.TransactionHash)
}

// TestCounter_IncrementWithSessionKeyPlugin
func TestCounter_IncrementWithSessionKeyPlugin(t *testing.T) {
	v := &accountPlugin{}
	err := v.Read(".sessionkey.json")
	if err != nil {
		t.Fatal(err)
	}
	sessionPrivateKeyInt, ok := big.NewInt(0).SetString(sessionPrivateKey, 0)
	if !ok {
		t.Fatal("could not match *big.Int private key with current value")
	}
	sessionPublicKeyInt, _, err := starknetgo.Curve.PrivateToPoint(sessionPrivateKeyInt)
	if err != nil {
		t.Fatal(err)
	}
	sessionPublicKey := fmt.Sprintf("0x%s", sessionPublicKeyInt.Text(16))
	token := sessionToken(privateKey, v.AccountAddress, sessionPublicKey)
	IncrementWithSessionKeyPlugin(t, v.AccountAddress, v.PluginHash, token, counterAddress)
}
