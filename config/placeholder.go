package config

import (
	"fmt"
	"rpctesting/types"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Placeholder string

const (
	SIGNER           Placeholder = "signer"
	NONCE            Placeholder = "nonce"
	TX_HASH          Placeholder = "tx.hash"
	CONTRACT_ADDRESS Placeholder = "contract.address"
	TX_BLOCK_NUMBER  Placeholder = "tx.blocknumber"
	TX_BLOCK_HASH    Placeholder = "tx.blockhash"
)

func NewPlaceholder(str string) Placeholder {
	return Placeholder(strings.ToLower(str))
}

func ConvertArgumentsWithTXReceipt(arguments any, txCall *types.ExecutedCall) error {

	switch v := arguments.(type) {
	case []interface{}:
		replacePlaceholdersInArray(v, txCall)

	case map[string]interface{}:
		replacePlaceholdersInObject(v, txCall)

	default:
		return fmt.Errorf("unsupported method argument type: %T", v)
	}
	return nil
}

func replacePlaceholdersInObject(obj map[string]interface{}, txCall *types.ExecutedCall) {
	for k, v := range obj {

		switch val := v.(type) {
		case string:
			obj[k] = replacePlaceholder(val, txCall)

		case map[string]interface{}:
			replacePlaceholdersInObject(v.(map[string]interface{}), txCall)

		case []interface{}:
			ConvertArgumentsWithTXReceipt(v.([]interface{}), txCall)
		}
	}
}

func replacePlaceholdersInArray(arr []interface{}, txCall *types.ExecutedCall) {
	for k, v := range arr {

		switch val := v.(type) {
		case string:
			arr[k] = replacePlaceholder(val, txCall)

		case map[string]interface{}:
			replacePlaceholdersInObject(v.(map[string]interface{}), txCall)

		case []interface{}:
			ConvertArgumentsWithTXReceipt(v.([]interface{}), txCall)
		}
	}
}

func replacePlaceholder(str string, txCall *types.ExecutedCall) string {
	switch NewPlaceholder(str) {
	case SIGNER:
		return txCall.From.String()
	case NONCE:
		return "0x" + strconv.FormatUint(txCall.Nonce, 16)
	case CONTRACT_ADDRESS:
		return txCall.ContractAddress.String()
	case TX_HASH:
		return txCall.TxReceipt.TxHash.String()
	case TX_BLOCK_NUMBER:
		return hexutil.EncodeBig(txCall.TxReceipt.BlockNumber)
	case TX_BLOCK_HASH:
		return txCall.TxReceipt.BlockHash.String()
	default:
		return str
	}
}
