package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"rpctesting/chain"
	"rpctesting/types"

	"gopkg.in/yaml.v3"
)

func main() {
	// Define flags
	testDir := flag.String("testDirectory", "./testfiles", "a directory, where test config files are located")
	if testDir == nil {
		fmt.Println("testDirectory is required")
		return
	}

	// Parse flags
	flag.Parse()

	testConfig, err := loadAllConfigs(*testDir)
	if err != nil {
		fmt.Printf("Failed to load test configs: %s\n", err)
	}

	tstJson, _ := json.MarshalIndent(testConfig, "", " ")

	fmt.Println(string(tstJson))

	// fmt.Println("Deploying contracts...")
	// var contracts []types.DeployedContract
	// for _, test := range testConfig {
	// 	contracts, err = chain.DeployContracts(context.Background(), test.Deploy)
	// 	if err != nil {
	// 		fmt.Printf("Failed to deploy contracts: %s\n", err)
	// 	}
	// }
	// fmt.Printf("deployed %v contracts", len(contracts))

	fmt.Println("Calling contracts...")

	fmt.Println("Running tests...")

	err = chain.Call()
	if err != nil {
		fmt.Printf("Failed to call contracts: %s\n", err)
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
			fmt.Println("file name: ", name)
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
