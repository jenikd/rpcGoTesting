package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"rpctesting/chain"
	"rpctesting/config"
	"rpctesting/tools"
	"rpctesting/types"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
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

	logger := tools.NewLogger(tools.InfoLevel)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if testDir == nil {
		t.Fatalf("testDirectory is required")
		return
	}

	// Parse flags
	flag.Parse()

	logger.Debugln("Loading configuration...")

	client, signer, err := chain.GetSignerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to get client: %s\n", err)
	}
	defer client.Close()

	testConfigFiles, err := config.LoadAllConfigs(*testDir)
	if err != nil {
		t.Fatalf("Failed to load test configs: %s\n", err)
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

					err := config.ConvertArgumentsWithTXReceipt(testCall.Arguments, contractCalls[testCall.CallID])
					if err != nil {
						t.Fatalf("failed to convert arguments: %s", err)
					}
				}

				logger.Debugln("testing method", testCall.MethodName, " :", testCall.Arguments)

				res, err := chain.MakeSimpleCall(ctx, client, testCall.MethodName, testCall.Arguments)
				if err != nil {
					if want, ok := testCall.Result.(string); ok &&
						(tools.NewResultType(want) == tools.NOT_AVAILABLE || tools.NewResultType(want) == tools.ERROR) {
						return
					}

					t.Fatalf("failed to call method : %s", err)
				}

				if err = tools.CheckResult(testCall.Result, res, logger, testCall.IgnoreFields...); err != nil {
					t.Fatalf("failed to check result: %s", err)
				}
			})
		}
	}
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
