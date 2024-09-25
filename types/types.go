package types

import "github.com/ethereum/go-ethereum/common"

type TestConfig struct {
	Deploy []DeployConfig   `yaml:"deploy"`
	Call   []CallConfig     `yaml:"contract_call"`
	Test   []TestCallConfig `yaml:"test_call"`
}

type DeployConfig struct {
	ContractID int    `yaml:"contract_id"`
	ABI        string `yaml:"abi"`
	Bytecode   string `yaml:"bytecode"`
}

type CallConfig struct {
	CallID     int    `yaml:"call_id"`
	ContractID int    `yaml:"contract_id"`
	MethodName string `yaml:"method_name"`
	Arguments  []any  `yaml:"arguments"`
}

type TestCallConfig struct {
	TestID     int    `yaml:"test_id"`
	TestName   string `yaml:"test_name"`
	CallID     int    `yaml:"call_id,omitempty"`
	MethodName string `yaml:"method_name"`
	Arguments  []any  `yaml:"arguments"`
}

type DeployedContract struct {
	ContractID int
	ABI        string
	Address    common.Address
	TxHash     common.Hash
}

type ExecutedCall struct {
	CallID int
	TxHash common.Hash
}
