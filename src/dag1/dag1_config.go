package dag1

import (
	"crypto/ecdsa"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/SamuelMarks/dag1/src/log"
	"github.com/SamuelMarks/dag1/src/node"
	"github.com/SamuelMarks/dag1/src/peer"
	"github.com/SamuelMarks/dag1/src/pos"
	"github.com/SamuelMarks/dag1/src/proxy"
)

type DAG1Config struct {
	DataDir     string `mapstructure:"datadir"`
	BindAddr    string `mapstructure:"listen"`
	ServiceAddr string `mapstructure:"service-listen"`
	ServiceOnly bool   `mapstructure:"service-only"`
	MaxPool     int    `mapstructure:"max-pool"`
	Store       bool   `mapstructure:"store"`
	LogLevel    string `mapstructure:"log"`

	NodeConfig node.Config `mapstructure:",squash"`
	PoSConfig  pos.Config  `mapstructure:",squash"`

	LoadPeers bool
	Proxy     proxy.AppProxy
	Key       *ecdsa.PrivateKey
	Logger    *logrus.Logger

	ConnFunc peer.CreateNetConnFunc

	Test      bool   `mapstructure:"test"`
	TestN     uint64 `mapstructure:"test_n"`
	TestDelay uint64 `mapstructure:"test_delay"`
	PeerSelector string `mapstructure:"peer_selector"`
}

func NewDefaultConfig() *DAG1Config {
	config := &DAG1Config{
		DataDir:     DefaultDataDir(),
		BindAddr:    ":1337",
		ServiceAddr: ":8000",
		ServiceOnly: false,
		ConnFunc:    net.DialTimeout,
		MaxPool:     2,
		NodeConfig:  *node.DefaultConfig(),
		PoSConfig:   *pos.DefaultConfig(),
		Store:       false,
		LogLevel:    "info",
		Proxy:       nil,
		Logger:      logrus.New(),
		LoadPeers:   true,
		Key:         nil,
		Test:        false,
		TestN:       ^uint64(0),
		TestDelay:   1,
		PeerSelector: "smart",
	}

	config.Logger.Level = LogLevel(config.LogLevel)
	dag1_log.NewLocal(config.Logger, config.LogLevel)
	//config.Proxy = sproxy.NewInmemAppProxy(config.Logger)
	//config.Proxy, _ = sproxy.NewSocketAppProxy("127.0.0.1:1338", "127.0.0.1:1339", 1*time.Second, config.Logger)
	config.NodeConfig.Logger = config.Logger
	config.NodeConfig.TestDelay = config.TestDelay

	return config
}

func DefaultBadgerDir() string {
	dataDir := DefaultDataDir()
	if dataDir != "" {
		return filepath.Join(dataDir, "badger_db")
	}
	return ""
}

func (c *DAG1Config) BadgerDir() string {
	return filepath.Join(c.DataDir, "badger_db")
}

func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := HomeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".dag1")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "DAG1")
		} else {
			return filepath.Join(home, ".dag1")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func LogLevel(l string) logrus.Level {
	switch l {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.DebugLevel
	}
}
