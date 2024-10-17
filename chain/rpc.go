package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"rpctesting/config"
	t "rpctesting/types"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetSigner(ctx context.Context, clientConfig *config.ClientConfig, client *rpc.Client) (*bind.TransactOpts, error) {
	privateKey, err := getPrivateKey(clientConfig.Pk)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %s", err)
	}

	chainId, err := getChainId(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %s", err)
	}

	deployer, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployer: %s", err)
	}

	deployer.GasPrice, err = getGasPrice(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %s", err)
	}

	return deployer, nil
}

func ConvertArgumentsWithAbi(contractABI *abi.ABI, methodName string, arguments []any) ([]interface{}, error) {

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

func ConvertArgumentsWithTXReceipt(arguments []interface{}, txCall *t.ExecutedCall) error {

	for i, arg := range arguments {
		switch v := arg.(type) {
		case int:
			fmt.Printf("Integer: %d\n", v)
		case string:
			switch {
			case strings.Contains(arg.(string), "tx.hash"):
				arguments[i] = txCall.TxReceipt.TxHash.String()
			case strings.Contains(arg.(string), "contract.address"):
				arguments[i] = txCall.ContractAddress.String()
			case strings.Contains(arg.(string), "tx.blockNumber"):
				arguments[i] = hexutil.EncodeBig(txCall.TxReceipt.BlockNumber)
			default:
			}
		case bool:
			fmt.Printf("Boolean: %t\n", v)

		case map[string]interface{}, []interface{}:
			// do nothing
		default:
			fmt.Printf("Unknown type: arg: %v type %T\n", arg, v)
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
	case "bytes32":
		arg = common.HexToHash(argument.(string))

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

func getChainId(ctx context.Context, client *rpc.Client) (*big.Int, error) {
	var result *hexutil.Big
	err := client.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return result.ToInt(), nil
}

func getGasPrice(ctx context.Context, client *rpc.Client) (*big.Int, error) {
	var result string
	err := client.CallContext(ctx, &result, "eth_gasPrice")
	if err != nil {
		return nil, err
	}
	return hexutil.DecodeBig(result)
}
