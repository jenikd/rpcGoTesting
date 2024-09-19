package chain

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

func getChainId(ctx context.Context, client *rpc.Client) (string, error) {
	var result string
	err := client.CallContext(context.Background(), &result, "eth_chainId")
	if err != nil {
		return "", err
	}

	chainidInt, err := hexutil.DecodeUint64(result)
	log.Printf("chain id: %s, decoded cahin id: %d", result, chainidInt)

	return result, nil
}

func getGasPrice(ctx context.Context, client *rpc.Client) (*big.Int, error) {
	var gasPrice string
	err := rpcCall(ctx, client, "eth_gasPrice", nil, &gasPrice)
	if err != nil {
		return nil, err
	}
	log.Printf("gas price: %v", hexutil.MustDecodeBig(gasPrice))

	return hexutil.MustDecodeBig(gasPrice), nil
}

func getNonce(ctx context.Context, client *rpc.Client, address *common.Address) (*big.Int, error) {
	var nonce string
	err := rpcCall(ctx, client, "eth_getTransactionCount", []interface{}{address.String(), "pending"}, &nonce)
	if err != nil {
		return nil, err
	}
	log.Printf("nonce: %v", hexutil.MustDecodeBig(nonce))

	return hexutil.MustDecodeBig(nonce), nil
}

func estimateGas(ctx context.Context, client *rpc.Client, args []interface{}) (*big.Int, error) {
	var gas string
	err := rpcCall(ctx, client, "eth_estimateGas", args, &gas)
	if err != nil {
		return nil, err
	}
	log.Printf("est gas: %v", hexutil.MustDecodeBig(gas))

	return hexutil.MustDecodeBig(gas), nil
}

func signTx(ctx context.Context, client *rpc.Client, args []interface{}) (string, error) {
	type SignedResult struct {
		Raw string `json:"raw"`
	}
	var result SignedResult
	err := rpcCall(ctx, client, "eth_signTransaction", args, &result)
	if err != nil {
		return "", err
	}
	log.Printf("signed tx: %v", result)

	return result.Raw, nil
}

func sendRawTx(ctx context.Context, client *rpc.Client, args []interface{}) (string, error) {

	var txHash string
	err := rpcCall(ctx, client, "eth_sendRawTransaction", args, &txHash)
	if err != nil {
		return "", err
	}
	log.Printf("tx hash: %v", txHash)

	return txHash, nil
}

func getTxByHash(ctx context.Context, client *rpc.Client, args []interface{}) (string, error) {

	var result map[string]interface{}
	err := rpcCall(ctx, client, "eth_getTransactionByHash", args, &result)
	if err != nil {
		return "", err
	}
	s, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	log.Printf("tx: %v", string(s))

	return string(s), nil
}

func rpcCall(ctx context.Context, client *rpc.Client, method string, args []interface{}, result interface{}) error {

	err := client.CallContext(context.Background(), &result, method, args...)
	if err != nil {
		return err
	}

	return nil
}

func getTransactionReceipt(ctx context.Context, client *rpc.Client, txHash common.Hash) (interface{}, error) {
	var r interface{}
	err := client.CallContext(ctx, &r, "eth_getTransactionReceipt", txHash)
	if err == nil && r == nil {
		return nil, ethereum.NotFound
	}
	return r, err
}

func waitMined(ctx context.Context, client *rpc.Client, txHash common.Hash) (interface{}, error) {

	queryTicker := time.NewTicker(10 * time.Millisecond)
	defer queryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
			receipt, err := getTransactionReceipt(ctx, client, txHash)
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					continue
				}
				return nil, err
			}

			if receipt == nil {
				return nil, errors.New("transaction receipt is nil")
			}

			return receipt, nil
		}
	}
}
