package main

import (
	"os"
	"rpctesting/types"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
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
