package chain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"rpctesting/config"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetSigner(ctx context.Context, clientConfig *config.ClientConfig, client *rpc.Client) (*bind.TransactOpts, error) {
	privateKey, err := getPrivateKey(clientConfig.Pk)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %s", err)
	}

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

	//deployer.GasLimit = uint64(clientConfig.GasLimit)

	deployer.GasPrice, err = getGasPrice(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %s", err)
	}

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

func ConvertArgumentsWithAbi(contractABI *abi.ABI, methodName string, arguments []any) ([]interface{}, error) {
	// Convert arguments
	method := contractABI.Methods[methodName]
	args := make([]interface{}, len(method.Inputs))
	for i, input := range method.Inputs {

		arg, err := convertArgument(input.Type.String(), arguments[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert argument: %s, %s", arguments[i], err)
		}
		args[i] = arg
	}
	return args, nil
}

func ConvertArgumentsWithTXReceipt(arguments []interface{}, txReceipt *types.Receipt) error {

	for i, arg := range arguments {
		switch v := arg.(type) {
		case int:
			fmt.Printf("Integer: %d\n", v)
		case string:
			switch {
			case strings.Contains(arg.(string), "tx.hash"):
				arguments[i] = txReceipt.TxHash.String()
			case strings.Contains(arg.(string), "tx.blockNumber"):
				arguments[i] = hexutil.EncodeBig(txReceipt.BlockNumber)
			default:
			}
		case bool:
			fmt.Printf("Boolean: %t\n", v)

		case map[string]interface{}:
			// do nothing
		default:
			fmt.Printf("Unknown type: %T\n", v)
		}
	}
	return nil
}

func convertArgument(input string, argument interface{}) (arg interface{}, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	switch input {
	case "uint256":
		arg = big.NewInt(int64(argument.(int)))
	case "string":
		arg = argument.(string)
	case "bool":
		arg = argument.(bool)
	case "address":
		arg = common.HexToAddress(argument.(string))

	default:
		return nil, fmt.Errorf("unsupported type: %s", input)
	}
	return arg, nil
}

func MakeSimpleCall(ctx context.Context, client *ethclient.Client, methodName string, arguments []any) (interface{}, error) {

	var result interface{}
	err := client.Client().CallContext(ctx, &result, methodName, arguments...)
	if err != nil {
		return nil, fmt.Errorf("failed to call method %s, %s", methodName, err)
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
