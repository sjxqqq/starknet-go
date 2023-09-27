package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/google/go-querystring/query"
	"github.com/sjxqqq/starknet-go/rpc"
	"github.com/sjxqqq/starknet-go/types"
)

type StarkResp struct {
	Result []string `json:"result"`
}

type StateUpdate struct {
	BlockHash string `json:"block_hash"`
	NewRoot   string `json:"new_root"`
	OldRoot   string `json:"old_root"`
	StateDiff struct {
		StorageDiffs      map[string]interface{} `json:"storage_diffs"`
		DeployedContracts []struct {
			Address   string `json:"address"`
			ClassHash string `json:"class_hash"`
		} `json:"deployed_contracts"`
	} `json:"state_diff"`
}

func (sg *Gateway) ChainID(context.Context) (string, error) {
	return sg.ChainId, nil
}

type GatewayFunctionCall struct {
	FunctionCall
	Signature []string `json:"signature"`
}

type FunctionCall types.FunctionCall

func (f FunctionCall) MarshalJSON() ([]byte, error) {
	return json.Marshal(f)
	// output := map[string]interface{}{}
	// output["contract_address"] = f.ContractAddress.String()
	// if f.EntryPointSelector != "" {
	// 	output["entry_point_selector"] = f.EntryPointSelector
	// }
	// calldata := []string{}
	// for _, v := range f.Calldata {
	// 	data, _ := big.NewInt(0).SetString(v, 0)
	// 	calldata = append(calldata, data.Text(10))
	// }
	// output["calldata"] = calldata
	// return json.Marshal(output)
}

/*
'call_contract' wrapper and can accept a blockId in the hash or height format
*/
func (sg *Gateway) Call(ctx context.Context, call rpc.FunctionCall, blockHashOrTag string) ([]string, error) {
	gc := GatewayFunctionCall{
		FunctionCall: FunctionCall(call),
	}
	gc.EntryPointSelector = types.GetSelectorFromNameFelt(gc.EntryPointSelector.String())
	if len(gc.Calldata) == 0 {
		gc.Calldata = []*felt.Felt{}
	}

	if len(gc.Signature) == 0 {
		gc.Signature = []string{"0", "0"} // allows rpc and http clients to implement(has to be a better way)
	}

	req, err := sg.newRequest(ctx, http.MethodPost, "/call_contract", gc)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(blockHashOrTag, "0x"):
		appendQueryValues(req, url.Values{
			"blockHash": []string{blockHashOrTag},
		})
	case blockHashOrTag == "":
		appendQueryValues(req, url.Values{
			"blockNumber": []string{"pending"},
		})
	default:
		appendQueryValues(req, url.Values{
			"blockNumber": []string{blockHashOrTag},
		})
	}

	var resp StarkResp
	return resp.Result, sg.do(req, &resp)
}

/*
'add_transaction' wrapper for invokation requests
*/
func (sg *Gateway) Invoke(ctx context.Context, invoke types.FunctionInvoke) (*types.AddInvokeTransactionOutput, error) {
	tx := Transaction{
		Type:          INVOKE,
		SenderAddress: invoke.SenderAddress.String(),
		Version:       fmt.Sprintf("0x%x", invoke.Version),
		MaxFee:        fmt.Sprintf("0x%x", invoke.MaxFee),
	}
	if invoke.EntryPointSelector != "" {
		tx.EntryPointSelector = types.GetSelectorFromName(invoke.EntryPointSelector).String()
	}
	if invoke.Nonce != nil {
		tx.Nonce = fmt.Sprintf("0x%x", invoke.Nonce)
	}

	calldata := []string{}
	for _, v := range invoke.Calldata {
		bv, _ := big.NewInt(0).SetString(v, 0)
		calldata = append(calldata, bv.Text(10))
	}
	tx.Calldata = calldata

	if len(invoke.Signature) == 0 {
		tx.Signature = []string{}
	} else {
		// stop-gap before full *felt.Felt cutover
		tx.Signature = []string{invoke.Signature[0].String(), invoke.Signature[1].String()}
	}

	req, err := sg.newRequest(ctx, http.MethodPost, "/add_transaction", tx)
	if err != nil {
		return nil, err
	}
	var resp types.AddInvokeTransactionOutput
	return &resp, sg.do(req, &resp)
}

/*
'add_transaction' wrapper for compressing and deploying a compiled StarkNet contract
*/
func (sg *Gateway) Deploy(ctx context.Context, contract rpc.DeprecatedContractClass, deployRequest rpc.DeployAccountTxn) (resp types.AddDeployResponse, err error) {
	panic("deploy transaction has been removed, use account.Deploy() instead")
}

type DeployAccountRequest types.DeployAccountRequest

func (d DeployAccountRequest) MarshalJSON() ([]byte, error) {
	if d.Type != "DEPLOY_ACCOUNT" {
		return nil, errors.New("wrong type")
	}
	output := map[string]interface{}{}
	constructorCalldata := []string{}
	for _, value := range d.ConstructorCalldata {
		constructorCalldata = append(constructorCalldata, types.SNValToBN(value).Text(10))
	}
	output["constructor_calldata"] = constructorCalldata
	output["max_fee"] = fmt.Sprintf("0x%x", d.MaxFee)
	output["version"] = fmt.Sprintf("0x%x", d.Version)
	signature := []string{}
	for _, value := range d.Signature {
		signature = append(signature, value.Text(10))
	}
	output["signature"] = signature
	if d.Nonce != nil {
		output["nonce"] = fmt.Sprintf("0x%x", d.Nonce)
	}
	output["type"] = "DEPLOY_ACCOUNT"
	if d.ContractAddressSalt == "" {
		d.ContractAddressSalt = "0x0"
	}
	contractAddressSalt := fmt.Sprintf("0x%x", types.SNValToBN(d.ContractAddressSalt))
	output["contract_address_salt"] = contractAddressSalt
	classHash := fmt.Sprintf("0x%x", types.SNValToBN(d.ClassHash))
	output["class_hash"] = classHash
	return json.Marshal(output)
}

/*
'add_transaction' wrapper for deploying a compiled StarkNet account
*/
func (sg *Gateway) DeployAccount(ctx context.Context, deployAccountRequest types.DeployAccountRequest) (resp types.AddDeployResponse, err error) {
	d := DeployAccountRequest(deployAccountRequest)
	d.Type = DEPLOY_ACCOUNT

	req, err := sg.newRequest(ctx, http.MethodPost, "/add_transaction", d)
	if err != nil {
		return resp, err
	}

	return resp, sg.do(req, &resp)
}

/*
'add_transaction' wrapper for compressing and declaring a contract class
*/
func (sg *Gateway) Declare(ctx context.Context, contract rpc.DeprecatedContractClass, declareRequest DeclareRequest) (resp types.AddDeclareResponse, err error) {
	declareRequest.Type = DECLARE

	req, err := sg.newRequest(ctx, http.MethodPost, "/add_transaction", declareRequest)
	if err != nil {
		return resp, err
	}

	return resp, sg.do(req, &resp)
}

// type DeployRequest rpc.DeployRequest

// func (d DeployRequest) MarshalJSON() ([]byte, error) {
// 	calldata := []string{}
// 	for _, value := range d.ConstructorCalldata {
// 		calldata = append(calldata, types.SNValToBN(value).Text(10))
// 	}
// 	d.ConstructorCalldata = calldata
// 	return json.Marshal(types.DeployRequest(d))
// }

type DeclareRequest struct {
	Type          string                      `json:"type"`
	SenderAddress *felt.Felt                  `json:"sender_address"`
	Version       string                      `json:"version"`
	MaxFee        string                      `json:"max_fee"`
	Nonce         string                      `json:"nonce"`
	Signature     []string                    `json:"signature"`
	ContractClass rpc.DeprecatedContractClass `json:"contract_class"`
}

func (sg *Gateway) StateUpdate(ctx context.Context, opts *BlockOptions) (*StateUpdate, error) {
	req, err := sg.newRequest(ctx, http.MethodGet, "/get_state_update", nil)
	if err != nil {
		return nil, err
	}

	if opts != nil {
		vs, err := query.Values(opts)
		if err != nil {
			return nil, err
		}
		appendQueryValues(req, vs)
	}

	var resp StateUpdate
	return &resp, sg.do(req, &resp)
}

func (sg *Gateway) ContractAddresses(ctx context.Context) (*ContractAddresses, error) {
	req, err := sg.newRequest(ctx, http.MethodGet, "/get_contract_addresses", nil)
	if err != nil {
		return nil, err
	}

	var resp ContractAddresses
	return &resp, sg.do(req, &resp)
}

type ContractAddresses struct {
	Starknet             string `json:"Starknet"`
	GpsStatementVerifier string `json:"GpsStatementVerifier"`
}
