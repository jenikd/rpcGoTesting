package chain

import (
	"context"
	"encoding/json"
	"fmt"
	t "rpctesting/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ContractsDeployer interface {
	DeployContracts(deployConfig []t.DeployConfig) (map[int]*t.DeployedContract, error)
}

type Deployer struct {
	ctx    context.Context
	client *ethclient.Client
	signer *bind.TransactOpts
}

func NewDeployer(ctx context.Context, client *ethclient.Client, signer *bind.TransactOpts) ContractsDeployer {
	deployer := &Deployer{
		ctx:    ctx,
		client: client,
		signer: signer,
	}
	return deployer
}

func (d *Deployer) DeployContracts(deployConfig []t.DeployConfig) (map[int]*t.DeployedContract, error) {
	deployedContracts := map[int]*t.DeployedContract{}

	for _, deploy := range deployConfig {

		deployedContract, err := d.deployContract(deploy.ContractID, deploy.ABI, deploy.Bytecode)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy contract id: %d,  %s", deploy.ContractID, err)
		}

		deployedContracts[deploy.ContractID] = deployedContract
	}

	return deployedContracts, nil
}

func (d *Deployer) deployContract(id int, contractABI string, contractBin string) (*t.DeployedContract, error) {
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
	_, tx, _, err := bind.DeployContract(d.signer, abi, bytecode, d.client)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contract: %s", err)
	}

	txHash := tx.Hash()

	// Wait for the transaction to be mined
	contractAddress, err := bind.WaitDeployed(d.ctx, d.client, tx)
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
