package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"rpctesting/chain"
	"rpctesting/config"
	"rpctesting/types"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define flags
	testDir := flag.String("testDirectory", "./testfiles", "a directory, where test config files are located")
	if testDir == nil {
		log.Println("testDirectory is required")
		return
	}

	// Parse flags
	flag.Parse()

	testConfigFileMap, err := loadAllConfigs(*testDir)
	if err != nil {
		log.Printf("Failed to load test configs: %s\n", err)
	}

	tstJson, _ := json.MarshalIndent(testConfigFileMap, "", " ")
	log.Println(string(tstJson))

	clientConfig, err := config.GetClientConfig()
	if err != nil {
		log.Printf("Failed to get client config: %s\n", err)
		return
	}

	signer, err := chain.GetSigner(ctx, clientConfig)
	if err != nil {
		log.Printf("Failed to get signer: %s\n", err)
		return
	}

	log.Println("Deploying contracts...")
	var contracts []*types.DeployedContract
	for _, test := range testConfigFileMap {
		contracts, err = chain.DeployContracts(ctx, signer, clientConfig, test.Deploy)
		if err != nil {
			log.Printf("Failed to deploy contracts: %s\n", err)
			return
		}

		log.Printf("deployed %v contracts", len(contracts))

		c, _ := json.MarshalIndent(contracts, "", " ")
		log.Println("New contracts:", string(c))

		log.Println("Calling contracts...")

		log.Println("Running tests...")
	}

	err = chain.Call()
	if err != nil {
		log.Printf("Failed to call contracts: %s\n", err)
	}
}

func loadAllConfigs(testDir string) (map[string]types.TestConfig, error) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		return nil, err
	}

	configs := make(map[string]types.TestConfig)
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if !strings.HasSuffix(name, ".yaml") {
				continue
			}
			data, err := os.ReadFile(testDir + "/" + name)
			if err != nil {
				return nil, err
			}

			var config types.TestConfig
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				return nil, err
			}

			configs[name] = config
		}
	}

	return configs, nil
}
