package chain

import (
	"context"
	"encoding/json"
	"fmt"
	t "rpctesting/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

func MakeContractCalls(ctx context.Context, signer *bind.TransactOpts, client *ethclient.Client, callConfig []t.CallConfig, contracts map[int]*t.DeployedContract) (map[int]*t.ExecutedCall, error) {

	executedCalls := map[int]*t.ExecutedCall{}

	for _, call := range callConfig {

		if call.ContractID >= 0 {

			contract := contracts[call.ContractID]

			executedCall, err := makeContractCall(ctx, client, &call, contract, signer)
			if err != nil {
				return nil, fmt.Errorf("failed to call contract id: %d,  %s", call.ContractID, err)
			}

			executedCalls[call.ContractID] = executedCall
		}
	}
	return executedCalls, nil
}

func makeContractCall(ctx context.Context, client *ethclient.Client, call *t.CallConfig, contract *t.DeployedContract, signer *bind.TransactOpts) (*t.ExecutedCall, error) {

	// Unmarshal the contract ABI
	var decodedAbi abi.ABI
	if err := json.Unmarshal([]byte(contract.ABI), &decodedAbi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract ABI: %s", err)
	}

	args, err := ConvertArgumentsWithAbi(&decodedAbi, call.MethodName, call.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %s", err)
	}

	// data, err := decodedAbi.Pack(call.MethodName, args...)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to pack arguments: %s", err)
	// }

	// log.Println("Packed data:", common.Bytes2Hex(data))

	// Get the contract instance
	boundContract := bind.NewBoundContract(contract.Address, decodedAbi, client, client, client)

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
		CallID:    call.CallID,
		TxReceipt: txReceipt,
	}, nil
}
