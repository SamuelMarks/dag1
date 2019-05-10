package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/SamuelMarks/dag1/src/dummy"
	"github.com/SamuelMarks/dag1/src/dag1"
	"github.com/SamuelMarks/dag1/src/log"
	aproxy "github.com/SamuelMarks/dag1/src/proxy"
	"github.com/SamuelMarks/dag1/tester"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewRunCmd returns the command that starts a DAG1 node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run node",
		RunE:  runDAG1,
	}
	AddRunFlags(cmd)
	return cmd
}

func runSingleDAG1(config *CLIConfig) error {
	config.DAG1.Logger.Level = dag1.LogLevel(config.DAG1.LogLevel)
	config.DAG1.NodeConfig.Logger = config.DAG1.Logger
	if config.Log2file {
		f, err := os.OpenFile(fmt.Sprintf("dag1_%v.log", config.DAG1.BindAddr),
			os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
		}
		mw := io.MultiWriter(os.Stdout, f)
		config.DAG1.NodeConfig.Logger.SetOutput(mw)
	}
	if config.Syslog {
		hook, err := dag1_log.NewSyslogHook("", "", "dag1")
		if err == nil {
			config.DAG1.NodeConfig.Logger.Hooks.Add(hook)
			config.DAG1.NodeConfig.Logger.SetFormatter(&logrus.TextFormatter{
				DisableColors: true,
			})
		}
	}

	dag1_log.NewLocal(config.DAG1.Logger, config.DAG1.LogLevel)

	config.DAG1.Logger.WithFields(logrus.Fields{
		"proxy-listen":   config.ProxyAddr,
		"client-connect": config.ClientAddr,
		"standalone":     config.Standalone,
		"service-only":   config.DAG1.ServiceOnly,

		"dag1.datadir":        config.DAG1.DataDir,
		"dag1.bindaddr":       config.DAG1.BindAddr,
		"dag1.service-listen": config.DAG1.ServiceAddr,
		"dag1.maxpool":        config.DAG1.MaxPool,
		"dag1.store":          config.DAG1.Store,
		"dag1.loadpeers":      config.DAG1.LoadPeers,
		"dag1.log":            config.DAG1.LogLevel,

		"dag1.node.heartbeat":  config.DAG1.NodeConfig.HeartbeatTimeout,
		"dag1.node.tcptimeout": config.DAG1.NodeConfig.TCPTimeout,
		"dag1.node.cachesize":  config.DAG1.NodeConfig.CacheSize,
		"dag1.node.synclimit":  config.DAG1.NodeConfig.SyncLimit,
	}).Debug("RUN")

	if !config.Standalone {
		p, err := aproxy.NewGrpcAppProxy(
			config.ProxyAddr,
			config.DAG1.NodeConfig.HeartbeatTimeout,
			config.DAG1.Logger,
		)

		if err != nil {
			config.DAG1.Logger.Error("Cannot initialize socket AppProxy:", err)
			return nil
		}
		config.DAG1.Proxy = p
	} else {
		p := dummy.NewInmemDummyApp(config.DAG1.Logger)
		config.DAG1.Proxy = p
	}

	engine := dag1.NewDAG1(&config.DAG1)

	if err := engine.Init(); err != nil {
		config.DAG1.Logger.Error("Cannot initialize engine:", err)
		return nil
	}

	if config.DAG1.Test {
		p := engine.Peers
		go func() {
			for {
				time.Sleep(10 * time.Second)
				ct := engine.Node.GetConsensusTransactionsCount()
				pdl := engine.Node.GetPendingLoadedEvents()
				// 3 - number of notes in test; 10 - number of transactions sent at once
				if ct >= 3*10*config.DAG1.TestN && pdl < 1 {
					engine.Node.PrintStat() // this is for debug tag only
					time.Sleep(10 * time.Second)
					engine.Node.Shutdown()
					break
				}
			}
		}()
		go tester.PingNodesN(p.Sorted, p.ByPubKey, config.DAG1.TestN,
			config.DAG1.TestDelay, config.DAG1.Logger,
			config.ProxyAddr)
	}

	engine.Node.Register()
	engine.Run()

	return nil
}

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {

	// local config here is used to set default values for the flags below
	config := NewDefaultCLIConfig()

	cmd.Flags().String("datadir", config.DAG1.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", config.DAG1.LogLevel, "debug, info, warn, error, fatal, panic")
	cmd.Flags().Bool("log2file", config.Log2file, "duplicate log output into file dag1_<BindAddr>.log")
	switch runtime.GOOS {
	default:
		cmd.Flags().Bool("syslog", config.Syslog, "duplicate log output into syslog")
		fallthrough
	case "plan9", "nacl":
		cmd.Flags().String("pidfile", config.Pidfile, "pidfile location; /tmp/dag1.pid by default")
	case "windows":
	}

	// Network
	cmd.Flags().StringP("listen", "l", config.DAG1.BindAddr, "Listen IP:Port for dag1 node")
	cmd.Flags().DurationP("timeout", "t", config.DAG1.NodeConfig.TCPTimeout, "TCP Timeout")
	cmd.Flags().Int("max-pool", config.DAG1.MaxPool, "Connection pool size max")

	// Proxy
	cmd.Flags().Bool("standalone", config.Standalone, "Do not create a proxy")
	cmd.Flags().Bool("service-only", config.DAG1.ServiceOnly, "Only host the http service")
	cmd.Flags().StringP("proxy-listen", "p", config.ProxyAddr, "Listen IP:Port for dag1 proxy")
	cmd.Flags().StringP("client-connect", "c", config.ClientAddr, "IP:Port to connect to client")

	// Service
	cmd.Flags().StringP("service-listen", "s", config.DAG1.ServiceAddr, "Listen IP:Port for HTTP service")

	// Store
	cmd.Flags().Bool("store", config.DAG1.Store, "Use badgerDB instead of in-mem DB")
	cmd.Flags().Int("cache-size", config.DAG1.NodeConfig.CacheSize, "Number of items in LRU caches")

	// Node configuration
	cmd.Flags().Duration("heartbeat", config.DAG1.NodeConfig.HeartbeatTimeout, "Time between gossips")
	cmd.Flags().Int64("sync-limit", config.DAG1.NodeConfig.SyncLimit, "Max number of events for sync")

	// Test
	cmd.Flags().Bool("test", config.DAG1.Test, "Enable testing (sends transactions to random nodes in the network)")
	cmd.Flags().Uint64("test_n", config.DAG1.TestN, "Number of transactions to send")
	cmd.Flags().Uint64("test_delay", config.DAG1.TestDelay, "Number of second to delay before sending transactions")
	cmd.Flags().String("peer_selector", config.DAG1.PeerSelector, "Peer selector to user for the next peer; available: random,smart,fair,unfair,franky")
}

//Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command, config *CLIConfig) error {
	// cmd.Flags() includes flags from this command and all persistent flags from the parent
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	viper.SetConfigName("dag1")              // name of config file (without extension)
	viper.AddConfigPath(config.DAG1.DataDir) // search root directory
	// viper.AddConfigPath(filepath.Join(config.DAG1.DataDir, "dag1")) // search root directory /config
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		config.DAG1.Logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		config.DAG1.Logger.Debugf("No config file found in: %s", config.DAG1.DataDir)
	} else {
		return err
	}
	return nil
}
