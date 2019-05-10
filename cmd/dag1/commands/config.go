package commands

import (
	"github.com/SamuelMarks/dag1/src/dag1"
	"os"
	"path/filepath"
)

//CLIConfig contains configuration for the Run command
type CLIConfig struct {
	DAG1   dag1.DAG1Config `mapstructure:",squash"`
	ProxyAddr  string                  `mapstructure:"proxy-listen"`
	ClientAddr string                  `mapstructure:"client-connect"`
	Standalone bool                    `mapstructure:"standalone"`
	Log2file   bool                    `mapstructure:"log2file"`
	Pidfile    string                  `mapstructure:"pidfile"`
	Syslog     bool                    `mapstructure:"syslog"`
}

//NewDefaultCLIConfig creates a CLIConfig with default values
func NewDefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		DAG1:   *dag1.NewDefaultConfig(),
		ProxyAddr:  "127.0.0.1:1338",
		ClientAddr: "127.0.0.1:1339",
		Standalone: false,
		Log2file:   false,
		Pidfile:    filepath.Join(os.TempDir(), "dag1.pid"),
		Syslog:     false,
	}
}
