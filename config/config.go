// env.go

package config

import (
	"os"
	"rpctesting/types"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type ClientConfig struct {
	ProviderUrl string
	Pk          string
	GasLimit    uint64
}

func GetClientConfig() (*ClientConfig, error) {

	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		return nil, err
	}

	gasLimit, err := strconv.ParseUint(os.Getenv("GAS_LIMIT"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		ProviderUrl: os.Getenv("PROVIDER_URL"),
		Pk:          os.Getenv("PK"),
		GasLimit:    gasLimit,
	}, nil
}

func LoadAllConfigs(testDir string) (map[string]types.TestConfig, error) {
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
