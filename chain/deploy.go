package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	t "rpctesting/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

func DeployContracts(ctx context.Context, client *ethclient.Client, signer *bind.TransactOpts, deployConfig []t.DeployConfig) (map[int]*t.DeployedContract, error) {

	deployedContracts := map[int]*t.DeployedContract{}

	for _, deploy := range deployConfig {
		log.Printf("Deploying contract %d\n", deploy.ContractID)

		deployedContract, err := deployContract(ctx, client, deploy.ContractID, deploy.ABI, deploy.Bytecode, signer)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy contract id: %d,  %s", deploy.ContractID, err)
		}

		deployedContracts[deploy.ContractID] = deployedContract
	}

	return deployedContracts, nil
}

func deployContract(ctx context.Context, client *ethclient.Client, id int, contractABI string, contractBin string, signer *bind.TransactOpts) (*t.DeployedContract, error) {
	// Unmarshal the contract ABI
	var abi abi.ABI
	if err := json.Unmarshal([]byte(contractABI), &abi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract ABI: %s", err)
	}

	bytecode, err := hexutil.Decode(contractBin)
	if err != nil {
		return nil, fmt.Errorf("failed to decode contract bytecode: %s", err)
	}

	// Create a new instance of the contract
	_, tx, _, err := bind.DeployContract(signer, abi, bytecode, client)
	if err != nil {
		return nil, err
	}

	txHash := tx.Hash()

	// Wait for the transaction to be mined
	contractAddress, err := bind.WaitDeployed(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployed contract address: %s", err)
	}

	return &t.DeployedContract{
		ContractID: id,
		ABI:        contractABI,
		Address:    contractAddress,
		TxHash:     txHash,
	}, nil
}
