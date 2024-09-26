package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"rpctesting/chain"
	"rpctesting/config"
	"rpctesting/types"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if testDir == nil {
		t.Fatalf("testDirectory is required")
		return
	}

	// Parse flags
	flag.Parse()

	testConfigFiles, err := loadAllConfigs(*testDir)
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

	log.Println("Deploying contracts...")
	for fileName, test := range testConfigFiles {

		_, contractCalls, err := prepareTestData(ctx, client, signer, test)
		if err != nil {
			t.Errorf("Failed to prepare test data: %s", err)
			return
		}

		for _, testCall := range test.Test {
			t.Run(fileName+" - "+testCall.TestName, func(t *testing.T) {

				if testCall.CallID >= 0 {

					//callTxHash := calls[testCall.CallID].TxHash

					err := chain.ConvertArgumentsWithTXReceipt(testCall.Arguments, contractCalls[testCall.CallID].TxReceipt)
					if err != nil {
						t.Fatalf("failed to convert arguments: %s", err)
					}
					log.Println("method", testCall.MethodName, " :", testCall.Arguments)

					res, err := chain.MakeSimpleCall(ctx, client, testCall.MethodName, testCall.Arguments)
					if err != nil {
						t.Fatalf("failed to call method : %s", err)
					}

					r, err := json.MarshalIndent(res, "", "  ")
					if err != nil {
						t.Fatalf("failed to marshal result: %s", err)
					}
					log.Println("method result", testCall.MethodName, " :", string(r))

					expected, err := json.Marshal(testCall.Result)
					if err != nil {
						// handle the error
						log.Println("Error marshaling JSON:", err)
						return
					}

					got, err := json.Marshal(res)
					if err != nil {
						// handle the error
						log.Println("Error marshaling JSON:", err)
						return
					}

					patch, err := jsondiff.CompareJSON(
						expected,
						got,
						jsondiff.Ignores("/city/name", "/D"),
					)
					if err != nil {
						log.Fatal(err)
					}
					for _, op := range patch {
						log.Printf("Difference in result: %s\n", op)
					}
				}

			})
		}
	}
}

func prepareTestData(ctx context.Context, client *ethclient.Client, signer *bind.TransactOpts, test types.TestConfig) (map[int]*types.DeployedContract, map[int]*types.ExecutedCall, error) {

	// Deploy contracts
	contracts, err := chain.DeployContracts(ctx, client, signer, test.Deploy)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deploy contracts: %s", err)
	}

	log.Printf("deployed %v contracts", len(contracts))

	c, _ := json.MarshalIndent(contracts, "", " ")
	log.Println("New contracts:", string(c))

	log.Println("Calling contracts...")
	contractCalls, err := chain.MakeContractCalls(ctx, signer, client, test.Call, contracts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to call contracts: %s", err)
	}
	return contracts, contractCalls, nil
}
