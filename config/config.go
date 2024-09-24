// env.go

package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ClientConfig struct {
	ProviderUrl string
	Account     string
	Pk          string
	GasLimit    uint64
}

func GetClientConfig() (*ClientConfig, error) {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	gasLimit, err := strconv.ParseUint(os.Getenv("GAS_LIMIT"), 10, 64)
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		ProviderUrl: os.Getenv("PROVIDER_URL"),
		Account:     os.Getenv("ACCOUNT"),
		Pk:          os.Getenv("PK"),
		GasLimit:    gasLimit,
	}, nil
}
