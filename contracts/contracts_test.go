package contracts

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/joho/godotenv"
	starknetgo "github.com/sjxqqq/starknet-go"
	"github.com/sjxqqq/starknet-go/artifacts"
	devtest "github.com/sjxqqq/starknet-go/test"
	"github.com/sjxqqq/starknet-go/types"
	"github.com/sjxqqq/starknet-go/utils"
)

func TestGateway_InstallCounter(t *testing.T) {
	godotenv.Load()
	testConfiguration := beforeEach(t)

	type TestCase struct {
		providerType  starknetgo.ProviderType
		CompiledClass []byte
		Salt          string
		Inputs        []string
	}

	TestCases := map[string][]TestCase{
		"devnet": {
			{
				providerType:  starknetgo.ProviderGateway,
				CompiledClass: artifacts.CounterCompiled,
				Salt:          "0x0",
				Inputs:        []string{},
			},
		},
	}[testEnv]
	for _, test := range TestCases {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*60)
		defer cancel()
		var err error
		var tx *DeployOutput
		switch test.providerType {
		case starknetgo.ProviderGateway:
			provider := GatewayProvider(*testConfiguration.gateway)
			tx, err = provider.deployAndWaitNoWallet(ctx, test.CompiledClass, test.Salt, test.Inputs)
		default:
			t.Fatal("unsupported client type", test.providerType)
		}
		if err != nil {
			t.Fatal("should succeed, instead", err)
		}
		fmt.Println("deployment transaction", tx.TransactionHash)
	}
}

func TestRPC_InstallCounter(t *testing.T) {
	godotenv.Load()
	testConfiguration := beforeEach(t)

	type TestCase struct {
		providerType  starknetgo.ProviderType
		CompiledClass []byte
		Salt          string
		Inputs        []string
	}

	TestCases := map[string][]TestCase{
		"devnet": {
			{
				providerType:  starknetgo.ProviderRPC,
				CompiledClass: artifacts.CounterCompiled,
				Salt:          "0x01",
				Inputs:        []string{},
			},
		},
	}[testEnv]
	for _, test := range TestCases {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*60)
		defer cancel()
		var err error
		var tx *DeployOutput
		switch test.providerType {
		case starknetgo.ProviderRPC:
			provider := RPCProvider(*testConfiguration.rpc)
			tx, err = provider.deployAndWaitWithWallet(ctx, test.CompiledClass, test.Salt, test.Inputs)
		default:
			t.Fatal("unsupported client type", test.providerType)
		}
		if err != nil {
			t.Fatal("should succeed, instead", err)
		}
		fmt.Println("deployment transaction", tx.TransactionHash)
	}
}

func TestGateway_LoadAndExecuteCounter(t *testing.T) {
	godotenv.Load()
	testConfiguration := beforeEach(t)

	type TestCase struct {
		privateKey      string
		providerType    starknetgo.ProviderType
		accountContract artifacts.CompiledContract
	}

	TestCases := map[string][]TestCase{
		"devnet": {
			{
				privateKey:      "0x01",
				providerType:    starknetgo.ProviderGateway,
				accountContract: artifacts.AccountContracts[ACCOUNT_VERSION1][false][false],
			},
		},
	}[testEnv]
	for _, test := range TestCases {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*120)
		defer cancel()
		var err error
		var counterTransaction *DeployOutput
		var account *starknetgo.Account
		// shim a keystore into existing tests.
		// use string representation of the PK as a fake sender address for the keystore
		ks := starknetgo.NewMemKeystore()

		fakeSenderAddress := test.privateKey
		k := types.SNValToBN(test.privateKey)
		ks.Put(fakeSenderAddress, k)
		switch test.providerType {
		case starknetgo.ProviderGateway:
			pk, _ := big.NewInt(0).SetString(test.privateKey, 0)
			accountManager, err := InstallAndWaitForAccount(
				ctx,
				testConfiguration.gateway,
				pk,
				test.accountContract,
			)
			if err != nil {
				t.Fatal("error deploying account", err)
			}
			mint, err := devtest.NewDevNet().Mint(utils.TestHexToFelt(t, accountManager.AccountAddress), big.NewInt(int64(1000000000000000000)))
			if err != nil {
				t.Fatal("error deploying account", err)
			}
			fmt.Printf("current balance is %d\n", mint.NewBalance)
			provider := GatewayProvider(*testConfiguration.gateway)
			counterTransaction, err = provider.deployAndWaitNoWallet(ctx, artifacts.CounterCompiled, "0x0", []string{})
			if err != nil {
				t.Fatal("should succeed, instead", err)
			}
			fmt.Println("deployment transaction", counterTransaction.TransactionHash)
			account, err = starknetgo.NewGatewayAccount(utils.TestHexToFelt(t, fakeSenderAddress), utils.TestHexToFelt(t, accountManager.AccountAddress), ks, testConfiguration.gateway, starknetgo.AccountVersion1)
			if err != nil {
				t.Fatal("should succeed, instead", err)
			}
		default:
			t.Fatal("unsupported client type", test.providerType)
		}
		tx, err := account.Execute(ctx, []types.FunctionCall{{ContractAddress: utils.TestHexToFelt(t, counterTransaction.ContractAddress), EntryPointSelector: types.GetSelectorFromNameFelt("increment"), Calldata: []*felt.Felt{}}}, types.ExecuteDetails{})
		if err != nil {
			t.Fatal("should succeed, instead", err)
		}
		fmt.Println("increment transaction", tx.TransactionHash)
	}
}

func TestRPC_LoadAndExecuteCounter(t *testing.T) {
	godotenv.Load()
	testConfiguration := beforeEach(t)

	type TestCase struct {
		privateKey      string
		providerType    starknetgo.ProviderType
		accountContract artifacts.CompiledContract
	}

	TestCases := map[string][]TestCase{
		"devnet": {
			{
				privateKey:      "0xe3e70682c2094cac629f6fbed82c07cd",
				providerType:    starknetgo.ProviderRPC,
				accountContract: artifacts.AccountContracts[ACCOUNT_VERSION1][false][false],
			},
		},
	}[testEnv]
	for _, test := range TestCases {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*120)
		defer cancel()
		var err error
		var counterTransaction *DeployOutput
		var account *starknetgo.Account
		ks := starknetgo.NewMemKeystore()

		fakeSenderAddress := test.privateKey
		k := types.SNValToBN(test.privateKey)
		ks.Put(fakeSenderAddress, k)
		switch test.providerType {
		case starknetgo.ProviderRPC:
			pk, _ := big.NewInt(0).SetString(test.privateKey, 0)
			fmt.Println("befor")
			accountManager := &AccountManager{}
			accountManager, err := InstallAndWaitForAccount(
				ctx,
				testConfiguration.rpc,
				pk,
				test.accountContract,
			)
			if err != nil {
				t.Fatal("error deploying account", err)
			}
			fmt.Println("after")
			mint, err := devtest.NewDevNet().Mint(utils.TestHexToFelt(t, accountManager.AccountAddress), big.NewInt(int64(1000000000000000000)))
			if err != nil {
				t.Fatal("error deploying account", err)
			}
			fmt.Printf("current balance is %d\n", mint.NewBalance)
			provider := RPCProvider(*testConfiguration.rpc)
			counterTransaction, err = provider.deployAndWaitWithWallet(ctx, artifacts.CounterCompiled, "0x0", []string{})
			if err != nil {
				t.Fatal("should succeed, instead", err)
			}
			fmt.Println("deployment transaction", counterTransaction.TransactionHash)
			account, err = starknetgo.NewRPCAccount(utils.TestHexToFelt(t, fakeSenderAddress), utils.TestHexToFelt(t, accountManager.AccountAddress), ks, testConfiguration.rpc, starknetgo.AccountVersion1)
			if err != nil {
				t.Fatal("should succeed, instead", err)
			}
		default:
			t.Fatal("unsupported client type", test.providerType)
		}
		tx, err := account.Execute(ctx, []types.FunctionCall{{ContractAddress: utils.TestHexToFelt(t, counterTransaction.ContractAddress), EntryPointSelector: types.GetSelectorFromNameFelt("increment"), Calldata: []*felt.Felt{}}}, types.ExecuteDetails{})
		if err != nil {
			t.Fatal("should succeed, instead", err)
		}
		fmt.Println("increment transaction", tx.TransactionHash)
	}
}
