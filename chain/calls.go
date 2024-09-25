package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	t "rpctesting/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func MakeCalls(ctx context.Context, signer *bind.TransactOpts, client *ethclient.Client, callConfig []t.CallConfig, contracts []*t.DeployedContract) ([]*t.ExecutedCall, error) {

	executedCalls := make([]*t.ExecutedCall, len(callConfig))

	for i, call := range callConfig {

		if call.ContractID >= 0 {

			contract := contracts[call.ContractID]

			executedCall, err := makeCall(ctx, client, &call, contract, signer)
			if err != nil {
				return nil, fmt.Errorf("failed to call contract id: %d,  %s", call.ContractID, err)
			}

			executedCalls[i] = executedCall
		}
	}
	return executedCalls, nil
}

func makeCall(ctx context.Context, client *ethclient.Client, call *t.CallConfig, contract *t.DeployedContract, signer *bind.TransactOpts) (*t.ExecutedCall, error) {

	// Unmarshal the contract ABI
	var decodedAbi abi.ABI
	if err := json.Unmarshal([]byte(contract.ABI), &decodedAbi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract ABI: %s", err)
	}

	args, err := convertArguments(&decodedAbi, call.MethodName, call.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %s", err)
	}

	data, err := decodedAbi.Pack(call.MethodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments: %s", err)
	}

	log.Println("Packed data:", common.Bytes2Hex(data))

	// Get the contract instance
	boundContract := bind.NewBoundContract(contract.Address, decodedAbi, client, client, client)

	//if abi.Methods[methodName].IsConstant() {

	//args := []interface{}{big.NewInt(42)}
	tx, err := boundContract.Transact(signer, "store", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to make transaction: %s", err)
	}
	// Wait for the transaction to be mined
	txReceipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract address: %s", err)
	}

	return &t.ExecutedCall{
		CallID: call.CallID,
		TxHash: txReceipt.TxHash,
	}, nil
}

func convertArguments(contractABI *abi.ABI, methodName string, arguments []any) ([]interface{}, error) {
	// Convert arguments
	method := contractABI.Methods[methodName]
	args := make([]interface{}, len(method.Inputs))
	for i, input := range method.Inputs {
		switch input.Type.String() {
		case "uint256":
			args[i] = big.NewInt(int64(arguments[i].(int)))
		case "string":
			args[i] = arguments[i].(string)
		case "bool":
			args[i] = arguments[i].(bool)
		case "address":
			args[i] = common.HexToAddress(arguments[i].(string))
			// Add more cases as needed for other types
		default:
			return nil, fmt.Errorf("unsupported type: %s", input.Type.String())
		}
	}
	return args, nil
}

func makeSimpleCall(ctx context.Context, client *ethclient.Client, methodName string, arguments []any) (map[string]interface{}, error) {

	var result map[string]interface{}
	err := client.Client().CallContext(ctx, &result, methodName, arguments...)
	if err != nil {
		return nil, fmt.Errorf("failed to call method : %s", err)
	}
	return result, nil
}
