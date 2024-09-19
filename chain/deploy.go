package chain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	t "rpctesting/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	PROVIDER    = "http://localhost:8545"
	PRIVATE_KEY = "0x67b1d87101671b127f5f8714789C7192f7ad340e"
	CHAIN_ID    = 4003
)

func DeployContracts(ctx context.Context, deployConfig []t.DeployConfig) ([]t.DeployedContract, error) {

	// chainId := big.NewInt(4003)
	// privateKey, err := getPrivateKey(PRIVATE_KEY)
	// if err != nil {
	// 	return nil, err
	// }
	// deployer, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	// if err != nil {
	// 	return nil, err
	// }

	deployer2 := &bind.TransactOpts{
		From:     common.HexToAddress("0x239fA7623354eC26520dE878B52f13Fe84b06971"),
		Signer:   nil,
		GasLimit: 1000000,

		// Nonce:    nil,
		// GasPrice: nil,
		// NoSend:   false,
		// ChainID:  big.NewInt(CHAIN_ID),
		// Value:    big.NewInt(0),
	}

	deployedContracts := make([]t.DeployedContract, len(deployConfig))

	for i, deploy := range deployConfig {
		fmt.Printf("Deploying contract %d\n", deploy.ContractID)

		client, err := getClient(PROVIDER)
		if err != nil {
			return nil, err
		}

		address, err := deployContract(ctx, client, deploy.ABI, deploy.Bytecode, deployer2)
		if err != nil {
			return nil, err
		}
		deployedContracts[i] = t.DeployedContract{
			ContractID: deploy.ContractID,
			ABI:        deploy.ABI,
			Address:    address,
		}
	}

	return deployedContracts, nil
}

func deployContract(ctx context.Context, client *ethclient.Client, contractABI string, contractBin string, deployer *bind.TransactOpts) (common.Address, error) {
	// Unmarshal the contract ABI
	var abi abi.ABI
	if err := json.Unmarshal([]byte(contractABI), &abi); err != nil {
		return common.Address{}, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(&ecdsa.PrivateKey{}, big.NewInt(4003))
	if err != nil {
		return common.Address{}, err
	}
	auth.From = common.HexToAddress("0x239fA7623354eC26520dE878B52f13Fe84b06971")

	// Create a new instance of the contract
	//contractAddr, tx, _, err := bind.DeployContract(deployer, abi, []byte(contractBin), client)
	contractAddr, tx, _, err := bind.DeployContract(auth, abi, []byte(contractBin), client)
	if err != nil {
		return common.Address{}, err
	}

	// Wait for the transaction to be mined
	txHash := tx.Hash()
	fmt.Printf("Contract deployed to address: %s, transaction hash: %s\n", contractAddr.Hex(), txHash.Hex())

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return common.Address{}, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, fmt.Errorf("contract deployment failed with status %v", receipt.Status)
	}

	return contractAddr, nil
}

func Call() error {
	ctx := context.Background()
	rpcURL := "http://localhost:8545"

	client, err := rpc.Dial(rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	gasPrice, err := getGasPrice(ctx, client)
	if err != nil {
		return err
	}

	from := common.HexToAddress("0x239fA7623354eC26520dE878B52f13Fe84b06971")
	to := common.HexToAddress("0x67b1d87101671b127f5f8714789C7192f7ad340e")
	value := "0xde0b6b3a7640000"

	nonce, err := getNonce(ctx, client, &from)
	if err != nil {
		return err
	}

	gas, err := estimateGas(ctx, client, []interface{}{
		map[string]interface{}{
			"from":  from.String(),
			"to":    to.String(),
			"value": value,
		},
	})
	if err != nil {
		return err
	}

	signedTX, err := signTx(ctx, client, []interface{}{
		map[string]interface{}{
			"from":     from.String(),
			"to":       to.String(),
			"value":    value,
			"gas":      hexutil.EncodeBig(gas),
			"gasPrice": hexutil.EncodeBig(gasPrice),
			"nonce":    hexutil.EncodeBig(nonce),
			"data":     "0x",
		},
	})
	if err != nil {
		return err
	}

	txHash, err := sendRawTx(ctx, client, []interface{}{signedTX})
	if err != nil {
		return err
	}

	// Wait for the transaction to be mined
	receipt, err := waitMined(ctx, client, common.HexToHash(txHash))
	if err != nil {
		return err
	}

	if receipt == nil {
		return fmt.Errorf("transaction failed with status %v", receipt)
	} else {
		s, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return err
		}

		log.Printf("tx: %v", string(s))
	}

	return nil
}

func addressFromPrivateKey(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

func getClient(provider string) (*ethclient.Client, error) {

	client, err := ethclient.Dial(provider)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getPrivateKey(key string) (*ecdsa.PrivateKey, error) {

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
