package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/SamuelMarks/dag1/src/dag1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//NewRunCmd returns the command that starts a DAG1 node
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Short:   "Run node",
		PreRunE: loadConfig,
		RunE:    runDAG1,
	}

	AddRunFlags(cmd)

	return cmd
}

/*******************************************************************************
* RUN
*******************************************************************************/

func buildConfig() error {
	dag1Port := 1337

	peersJSON := `[`

	for i := 0; i < config.NbNodes; i++ {
		nb := strconv.Itoa(i)

		dag1PortStr := strconv.Itoa(dag1Port + (i * 10))

		dag1Node := exec.Command("dag1", "keygen", "--pem=/tmp/dag1_configs/.dag1"+nb+"/priv_key.pem", "--pub=/tmp/dag1_configs/.dag1"+nb+"/key.pub")

		res, err := dag1Node.CombinedOutput()
		if err != nil {
			log.Fatal(err, res)
		}

		pubKey, err := ioutil.ReadFile("/tmp/dag1_configs/.dag1" + nb + "/key.pub")
		if err != nil {
			log.Fatal(err, res)
		}

		peersJSON += `	{
		"NetAddr":"127.0.0.1:` + dag1PortStr + `",
		"PubKeyHex":"` + string(pubKey) + `"
	},
`
	}

	peersJSON = peersJSON[:len(peersJSON)-2]
	peersJSON += `
]
`

	for i := 0; i < config.NbNodes; i++ {
		nb := strconv.Itoa(i)

		err := ioutil.WriteFile("/tmp/dag1_configs/.dag1"+nb+"/peers.json", []byte(peersJSON), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func sendTxs(dag1Node *exec.Cmd, i int) {
	ticker := time.NewTicker(1 * time.Second)
	nb := strconv.Itoa(i)

	txNb := 0

	for range ticker.C {
		if txNb == config.SendTxs {
			ticker.Stop()

			break
		}

		network := exec.Command("network", "proxy", "--node="+nb, "--submit="+nb+"_"+strconv.Itoa(txNb))

		err := network.Run()
		if err != nil {
			continue
		}

		txNb++
	}
}

func runDAG1(cmd *cobra.Command, args []string) error {
	if err := os.RemoveAll("/tmp/dag1_configs"); err != nil {
		log.Fatal(err)
	}

	if err := buildConfig(); err != nil {
		log.Fatal(err)
	}

	dag1Port := 1337
	servicePort := 8080

	wg := sync.WaitGroup{}

	var processes = make([]*os.Process, config.NbNodes)

	for i := 0; i < config.NbNodes; i++ {
		wg.Add(1)

		go func(i int) {
			nb := strconv.Itoa(i)
			dag1PortStr := strconv.Itoa(dag1Port + (i * 10))
			proxyServPortStr := strconv.Itoa(dag1Port + (i * 10) + 1)
			proxyCliPortStr := strconv.Itoa(dag1Port + (i * 10) + 2)

			servicePort := strconv.Itoa(servicePort + i)

			defer wg.Done()

			dag1Node := exec.Command("dag1", "run", "-l=127.0.0.1:"+dag1PortStr, "--datadir=/tmp/dag1_configs/.dag1"+nb, "--proxy-listen=127.0.0.1:"+proxyServPortStr, "--client-connect=127.0.0.1:"+proxyCliPortStr, "-s=127.0.0.1:"+servicePort, "--heartbeat="+config.DAG1.NodeConfig.HeartbeatTimeout.String())
			err := dag1Node.Start()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Running", i)

			if config.SendTxs > 0 {
				go sendTxs(dag1Node, i)
			}

			processes[i] = dag1Node.Process

			if err := dag1Node.Wait(); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Terminated", i)

		}(i)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range c {
			for _, proc := range processes {
				if err := proc.Kill(); err != nil {
					panic(err)
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

/*******************************************************************************
* CONFIG
*******************************************************************************/

//AddRunFlags adds flags to the Run command
func AddRunFlags(cmd *cobra.Command) {
	cmd.Flags().Int("nodes", config.NbNodes, "Amount of nodes to spawn")
	cmd.Flags().String("datadir", config.DAG1.DataDir, "Top-level directory for configuration and data")
	cmd.Flags().String("log", config.DAG1.LogLevel, "debug, info, warn, error, fatal, panic")
	cmd.Flags().Duration("heartbeat", config.DAG1.NodeConfig.HeartbeatTimeout, "Time between gossips")

	cmd.Flags().Int64("sync-limit", config.DAG1.NodeConfig.SyncLimit, "Max number of events for sync")
	cmd.Flags().Int("send-txs", config.SendTxs, "Send some random transactions")
}

func loadConfig(cmd *cobra.Command, args []string) error {

	err := bindFlagsLoadViper(cmd)
	if err != nil {
		return err
	}

	config, err = parseConfig()
	if err != nil {
		return err
	}

	config.DAG1.Logger.Level = dag1.LogLevel(config.DAG1.LogLevel)
	config.DAG1.NodeConfig.Logger = config.DAG1.Logger

	return nil
}

//Bind all flags and read the config into viper
func bindFlagsLoadViper(cmd *cobra.Command) error {
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

//Retrieve the default environment configuration.
func parseConfig() (*CLIConfig, error) {
	conf := NewDefaultCLIConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}
