package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"rpctesting/chain"
	"rpctesting/config"
	"rpctesting/tools"
	"rpctesting/types"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/wI2L/jsondiff"
	"gopkg.in/yaml.v3"
)

var testDir = flag.String("testDirectory", "../../testfiles", "directory, where test config files are located")

func TestParseYaml(t *testing.T) {
	data, err := os.ReadFile("../../testfiles/tracing.yaml")
	if err != nil {
		panic(err)
	}

	var config types.TestConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		t.Errorf("Failed to parse YAML: %s", err)
	}
}

func TestAllConfigs(t *testing.T) {

	logger := Logger{Logger: log.Default(), logLevel: InfoLevel}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if testDir == nil {
		t.Fatalf("testDirectory is required")
		return
	}

	// Parse flags
	flag.Parse()

	logger.Debugln("Loading configuration...")

	testConfigFiles, err := config.LoadAllConfigs(*testDir)
	if err != nil {
		t.Fatalf("Failed to load test configs: %s\n", err)
	}

	clientConfig, err := config.GetClientConfig()
	if err != nil {
		t.Fatalf("Failed to get client config: %s\n", err)
	}

	client, err := chain.GetClient(clientConfig.ProviderUrl)
	if err != nil {
		t.Fatalf("Failed to get ethClient: %s", err)
	}
	defer client.Close()

	signer, err := chain.GetSigner(ctx, clientConfig, client.Client())
	if err != nil {
		t.Fatalf("Failed to get signer: %s", err)
		return
	}

	// For all test files
	for fileName, test := range testConfigFiles {

		if test.Ignore {
			continue
		}

		_, contractCalls, err := prepareTestData(ctx, client, signer, test)
		if err != nil {
			t.Errorf("Failed to prepare test data for file %s, error: %s", fileName, err)
			return
		}

		// For all test calls
		for _, testCall := range test.Test {

			// Ignore test
			if testCall.IgnoreTest {
				continue
			}

			t.Run(fileName+" - "+testCall.TestName, func(t *testing.T) {

				if testCall.CallID > 0 {

					if contractCalls[testCall.CallID] == nil {
						t.Fatalf("call id %d not found", testCall.CallID)
					}

					err := chain.ConvertArgumentsWithTXReceipt(testCall.Arguments, contractCalls[testCall.CallID])
					if err != nil {
						t.Fatalf("failed to convert arguments: %s", err)
					}
				}

				logger.Debugln("testing method", testCall.MethodName, " :", testCall.Arguments)

				res, err := chain.MakeSimpleCall(ctx, client, testCall.MethodName, testCall.Arguments)
				if err != nil {
					if want, ok := testCall.Result.(string); ok && newResultType(want) == NOT_AVAILABLE {
						return
					}
					t.Fatalf("failed to call method : %s", err)
				}

				if err = checkResult(testCall.Result, res, logger, testCall.IgnoreFields...); err != nil {
					t.Fatalf("failed to check result: %s", err)
				}
			})
		}
	}
}

type ResultType string

const (
	NOT_AVAILABLE ResultType = "NOT_AVAILABLE"
	HEX_NUMBER    ResultType = "HEX_NUMBER"
	ARRAY         ResultType = "ARRAY"
)

func newResultType(s string) ResultType {
	return ResultType(s)
}

func checkResult(expected any, got any, logger Logger, ignoreFields ...string) error {

	if expected == nil {
		return nil
	}

	if str, ok := expected.(string); ok {

		want := newResultType(str)
		switch want {
		case HEX_NUMBER:
			if have, ok := got.(string); ok {
				_, err := hexutil.DecodeUint64(have)
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
		}
	}

	if isOk := isEqualJson(expected, got, logger, ignoreFields...); !isOk {
		printInterface(expected, logger, "expected:")
		printInterface(got, logger, "got     :")
		return fmt.Errorf("result is not as expected")
	}
	return nil
}

func isEqualJson(expected any, res any, logger Logger, ignoreFields ...string) bool {

	if reflect.TypeOf(expected).Kind() != reflect.TypeOf(res).Kind() {
		logger.Printf("Differet type of expected result, want: %s, got: %s\n", reflect.TypeOf(expected).Kind(), reflect.TypeOf(res).Kind())
		return false
	}

	err := removeIgnoredFields(expected, res, ignoreFields...)
	if err != nil {
		logger.Debugln("Error removing fields:", err)
		return false
	}
	expectedJson, err := json.Marshal(expected)
	if err != nil {
		logger.Debugln("Error marshaling JSON:", err)
		return true
	}

	got, err := json.Marshal(res)
	if err != nil {
		logger.Debugln("Error marshaling JSON:", err)
		return true
	}

	patch, err := jsondiff.CompareJSON(expectedJson, got)
	if err != nil {
		log.Fatal(err)
	}
	if len(patch.String()) == 0 {
		return true
	}
	for _, op := range patch {
		logger.Printf("Difference in result: %s\n", op)
	}
	return false
}

func removeIgnoredFields(expected any, res any, ignoreFields ...string) error {

	// expected is string
	if reflect.TypeOf(expected).Kind() == reflect.String {
		return nil
	}

	if reflect.TypeOf(expected).Kind() == reflect.Slice {

		var m map[string]interface{}

		for i := range res.([]any) {

			if reflect.TypeOf(expected.([]any)[i]).Kind() != reflect.TypeOf(m).Kind() ||
				reflect.TypeOf(res.([]any)[i]).Kind() != reflect.TypeOf(m).Kind() {

				return fmt.Errorf("not comparable types")
			}
			// remove ignored fields
			tools.DeleteFields(res.([]any)[i].(map[string]interface{}), ignoreFields...)
			tools.DeleteFields(expected.([]any)[i].(map[string]interface{}), ignoreFields...)
		}
	} else if reflect.TypeOf(expected).Kind() == reflect.Map {

		tools.DeleteFields(res.(map[string]interface{}), ignoreFields...)
		tools.DeleteFields(expected.(map[string]interface{}), ignoreFields...)
	}

	return nil
}

func printInterface(obj interface{}, logger Logger, v ...interface{}) {
	r, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		logger.Fatalf("failed to marshal object: %s", err)
	}
	logger.Println(v, ":", string(r))
}

func prepareTestData(ctx context.Context, client *ethclient.Client, signer *bind.TransactOpts, test types.TestConfig) (map[int]*types.DeployedContract, map[int]*types.ExecutedCall, error) {

	// Deploy contracts
	contracts, err := chain.NewDeployer(ctx, client, signer).DeployContracts(test.Deploy)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deploy contracts: %s", err)
	}

	// Call contracts
	contractCalls, err := chain.MakeContractCalls(ctx, signer, client, test.Call, contracts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call contracts: %s", err)
	}
	return contracts, contractCalls, nil
}

type Logger struct {
	*log.Logger
	logLevel int
}

const (
	TraceLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l *Logger) Debugf(format string, v ...interface{}) {

	if l.logLevel <= DebugLevel {
		l.Printf(format, v...)
	}
}

func (l *Logger) Debugln(v ...any) {

	if l.logLevel <= DebugLevel {
		l.Println(v...)
	}
}
