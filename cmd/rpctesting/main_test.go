package main

import (
	"os"
	"rpctesting/types"
	"testing"

	"gopkg.in/yaml.v3"
)

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
