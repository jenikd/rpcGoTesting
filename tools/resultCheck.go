package tools

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ResultType string

const (
	NOT_AVAILABLE ResultType = "NOT_AVAILABLE"
	HEX_NUMBER    ResultType = "HEX_NUMBER"
	ARRAY         ResultType = "ARRAY"
	HEX_BYTES     ResultType = "HEX_BYTES"
	STRING        ResultType = "STRING"
)

func NewResultType(s string) ResultType {
	return ResultType(s)
}

func CheckResult(expected any, got any, logger *Logger, ignoreFields ...string) error {

	if expected == nil {
		return nil
	}

	if _, ok := expected.(string); !ok {
		if reflect.TypeOf(expected).Kind() != reflect.TypeOf(got).Kind() {
			return fmt.Errorf("differet type of expected result, want: %s, got: %s", reflect.TypeOf(expected).Kind(), reflect.TypeOf(got).Kind())
		}
	}

	switch expected.(type) {
	case string:
		want := NewResultType(expected.(string))
		switch want {
		case HEX_NUMBER:
			if have, ok := got.(string); ok {
				_, err := hexutil.DecodeBig(have)
				if err != nil {
					return err
				}
				return nil
			} else {
				return fmt.Errorf("result is not a hex number")
			}
		case ARRAY:
			if have, ok := got.([]any); ok && len(have) > 0 {
				return nil
			} else {
				return fmt.Errorf("result is not an array with value")
			}
		case HEX_BYTES:
			if have, ok := got.(string); ok {
				_, err := hexutil.Decode(have)
				if err != nil {
					return fmt.Errorf("failed to decode hex bytes: %s, got: %s", err, got)
				}
				if len(have) == 0 {
					return fmt.Errorf("result is an empty hex bytes, got: %s", got)
				}
				return nil
			} else {
				return fmt.Errorf("result is not a hex bytes, got: %v", got)
			}
		case STRING:
			if have, ok := got.(string); ok {
				if len(have) == 0 {
					return fmt.Errorf("result is an empty string, got: %s", got)
				}
				return nil
			} else {
				return fmt.Errorf("result is not a string, got: %v", got)
			}
		case NOT_AVAILABLE:
			return fmt.Errorf("result should not be available, got: %v", got)
		}
	case map[string]interface{}:
		if err := IsEqualJson(expected, got, logger, ignoreFields...); err != nil {
			printInterface(expected, logger, "expected:")
			printInterface(got, logger, "got     :")
			return err
		}
	case []interface{}:
		for i := range expected.([]interface{}) {
			if err := CheckResult(expected.([]interface{})[i], got.([]interface{})[i], logger, ignoreFields...); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported result type: %T", expected)
	}

	return nil
}

func printInterface(obj interface{}, logger *Logger, v ...interface{}) {
	r, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		logger.Fatalf("failed to marshal object: %s", err)
	}
	logger.Println(v, ":", string(r))
}
