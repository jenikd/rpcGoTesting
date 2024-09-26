package chain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"rpctesting/config"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetSigner(ctx context.Context, clientConfig *config.ClientConfig) (*bind.TransactOpts, error) {
	privateKey, err := getPrivateKey(clientConfig.Pk)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %s", err)
	}

	client, err := rpc.Dial(clientConfig.ProviderUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rpc provider: %s", err)
	}
	defer client.Close()

	chainIdString, err := getChainId(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %s", err)
	}

	chainId, err := hexutil.DecodeBig(chainIdString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode chain id: %s", err)
	}

	deployer, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %s", err)
	}

	deployer.GasLimit = uint64(clientConfig.GasLimit)

	return deployer, nil
}

func Call() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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

func addressFromPrivateKey(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

func GetClient(provider string) (*ethclient.Client, error) {

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
