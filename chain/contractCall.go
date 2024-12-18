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

	// Convert arguments
	args, err := ConvertArgumentsWithAbi(&decodedAbi, call.MethodName, call.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %s", err)
	}

	// Get the contract instance
	boundContract := bind.NewBoundContract(contract.Address, decodedAbi, client, client, client)

	// Call the contract
	tx, err := boundContract.Transact(signer, call.MethodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to make transaction: %s", err)
	}

	// Wait for the transaction to be mined
	txReceipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract address: %s", err)
	}

	nonce, err := client.NonceAt(ctx, signer.From, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %s", err)
	}

	return &t.ExecutedCall{
		CallID:          call.CallID,
		ContractAddress: contract.Address,
		TxReceipt:       txReceipt,
		From:            signer.From,
		Nonce:           nonce,
	}, nil
}
