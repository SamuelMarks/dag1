package commands

import "github.com/SamuelMarks/dag1/src/dag1"

//CLIConfig contains configuration for the Run command
type CLIConfig struct {
	DAG1 dag1.DAG1Config `mapstructure:",squash"`
	NbNodes  int                     `mapstructure:"nodes"`
	SendTxs  int                     `mapstructure:"send-txs"`
	Stdin    bool                    `mapstructure:"stdin"`
	Node     int                     `mapstructure:"node"`
}

//NewDefaultCLIConfig creates a CLIConfig with default values
func NewDefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		DAG1: *dag1.NewDefaultConfig(),
		NbNodes:  4,
		SendTxs:  0,
		Stdin:    false,
		Node:     0,
	}
}
