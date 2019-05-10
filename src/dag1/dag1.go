package dag1

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/SamuelMarks/dag1/src/crypto"
	"github.com/SamuelMarks/dag1/src/log"
	"github.com/SamuelMarks/dag1/src/node"
	"github.com/SamuelMarks/dag1/src/peer"
	"github.com/SamuelMarks/dag1/src/peers"
	"github.com/SamuelMarks/dag1/src/poset"
	"github.com/SamuelMarks/dag1/src/service"
)

// DAG1 struct
type DAG1 struct {
	Config    *DAG1Config
	Node      *node.Node
	Transport peer.SyncPeer
	Store     poset.Store
	Peers     *peers.Peers
	Service   *service.Service
}

// NewDAG1 constructor
func NewDAG1(config *DAG1Config) *DAG1 {
	engine := &DAG1{
		Config: config,
	}

	return engine
}

func (l *DAG1) initTransport() error {
	createCliFu := func(target string,
		timeout time.Duration) (peer.SyncClient, error) {

		rpcCli, err := peer.NewRPCClient(
			peer.TCP, target, time.Second, l.Config.ConnFunc)
		if err != nil {
			return nil, err
		}

		return peer.NewClient(rpcCli)
	}

	producer := peer.NewProducer(
		l.Config.MaxPool, l.Config.NodeConfig.TCPTimeout, createCliFu)
	backend := peer.NewBackend(
		peer.NewBackendConfig(), l.Config.Logger, net.Listen)
	if err := backend.ListenAndServe(peer.TCP, l.Config.BindAddr); err != nil {
		return err
	}
	l.Transport = peer.NewTransport(l.Config.Logger, producer, backend)
	return nil
}

func (l *DAG1) initPeers() error {
	if !l.Config.LoadPeers {
		if l.Peers == nil {
			return fmt.Errorf("did not load peers but none was present")
		}

		return nil
	}

	peerStore := peers.NewJSONPeers(l.Config.DataDir)

	// We read "old" format of peers.json here, so only peer messages are specified
	// TODO: upgrade batch-ethkey to generate peers.json iin new format
	participants, err := peerStore.GetPeersFromMessages()

	if err != nil {
		return err
	}

	if participants.Len() < 2 {
		return fmt.Errorf("peers.json should define at least two peers")
	}

	l.Peers = participants

	return nil
}

func (l *DAG1) initStore() (err error) {
	if !l.Config.Store {
		l.Store = poset.NewInmemStore(l.Peers, l.Config.NodeConfig.CacheSize, &l.Config.PoSConfig)
		l.Config.Logger.Debug("created new in-mem store")
	} else {
		dbDir := l.Config.BadgerDir()
		l.Config.Logger.WithField("path", dbDir).Debug("Attempting to load or create database")
		l.Store, err = poset.LoadOrCreateBadgerStore(l.Peers, l.Config.NodeConfig.CacheSize, dbDir, &l.Config.PoSConfig)
		if err != nil {
			return
		}
	}

	if l.Store.NeedBootstrap() {
		l.Config.Logger.Debug("loaded store from existing database")
	} else {
		l.Config.Logger.Debug("created new store from blank database")
	}

	return
}

func (l *DAG1) initKey() error {
	if l.Config.Key == nil {
		pemKey := crypto.NewPemKey(l.Config.DataDir)

		privKey, err := pemKey.ReadKey()

		if err != nil {
			l.Config.Logger.Warn("Cannot read private key from file", err)

			privKey, err = Keygen(l.Config.DataDir)

			if err != nil {
				l.Config.Logger.Error("Cannot generate a new private key", err)

				return err
			}

			pem, _ := crypto.ToPemKey(privKey)

			l.Config.Logger.Info("Created a new key:", pem.PublicKey)
		}

		l.Config.Key = privKey
	}

	return nil
}

func (l *DAG1) initNode() error {
	key := l.Config.Key

	nodePub := fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey))
	n, ok := l.Peers.ReadByPubKey(nodePub)

	if !ok {
		return fmt.Errorf("cannot find self pubkey in peers.json")
	}

	nodeID := n.ID

	l.Config.Logger.WithFields(logrus.Fields{
		"participants": l.Peers,
		"id":           nodeID,
	}).Debug("PARTICIPANTS")

	var selectorArgs node.SelectorCreationFnArgs
	var selectorFn node.SelectorCreationFn

	switch strings.ToLower(l.Config.PeerSelector) {
	case "random":
		selectorArgs = node.RandomPeerSelectorCreationFnArgs{
			LocalAddr:    l.Config.BindAddr,
		}
		selectorFn =  node.NewRandomPeerSelectorWrapper
	case "smart":
		selectorArgs = node.SmartPeerSelectorCreationFnArgs{
			LocalAddr:    l.Config.BindAddr,
		}
		selectorFn =  node.NewSmartPeerSelectorWrapper
	case "fair":
		selectorArgs = node.FairPeerSelectorCreationFnArgs{
			LocalAddr:    l.Config.BindAddr,
		}
		selectorFn =  node.NewFairPeerSelectorWrapper
	case "unfair":
		selectorArgs = node.UnfairPeerSelectorCreationFnArgs{
			LocalAddr:    l.Config.BindAddr,
		}
		selectorFn =  node.NewUnfairPeerSelectorWrapper
	case "franky":
		selectorArgs = node.FrankyPeerSelectorCreationFnArgs{
			LocalAddr:    l.Config.BindAddr,
		}
		selectorFn =  node.NewFrankyPeerSelectorWrapper
	default:
		panic(fmt.Errorf("Unknown peer selector %v", l.Config.PeerSelector))
	}

	l.Node = node.NewNode(
		&l.Config.NodeConfig,
		nodeID,
		key,
		l.Peers,
		l.Store,
		l.Transport,
		l.Config.Proxy,
		selectorFn,
		selectorArgs,
		l.Config.BindAddr,
	)

	if err := l.Node.Init(); err != nil {
		return fmt.Errorf("failed to initialize node: %s", err)
	}

	return nil
}

func (l *DAG1) initService() error {
	if l.Config.ServiceAddr != "" {
		l.Service = service.NewService(l.Config.ServiceAddr, l.Node, l.Config.Logger)
	}
	return nil
}

// Init initializes the dag1 node
func (l *DAG1) Init() error {
	if l.Config.Logger == nil {
		l.Config.Logger = logrus.New()
		dag1_log.NewLocal(l.Config.Logger, l.Config.LogLevel)
	}

	if err := l.initPeers(); err != nil {
		return err
	}

	if err := l.initStore(); err != nil {
		return err
	}

	if err := l.initTransport(); err != nil {
		return err
	}

	if err := l.initKey(); err != nil {
		return err
	}

	if err := l.initNode(); err != nil {
		return err
	}

	if err := l.initService(); err != nil {
		return err
	}

	return nil
}

// Run hosts the services for the dag1 node
func (l *DAG1) Run() {
	if l.Service != nil {
		go l.Service.Serve()
	}
	l.Node.Run(true)
}

// Keygen generates a new key pair
func Keygen(datadir string) (*ecdsa.PrivateKey, error) {
	pemKey := crypto.NewPemKey(datadir)

	_, err := pemKey.ReadKey()

	if err == nil {
		return nil, fmt.Errorf("another key already lives under %s", datadir)
	}

	privKey, err := crypto.GenerateECDSAKey()

	if err != nil {
		return nil, err
	}

	if err := pemKey.WriteKey(privKey); err != nil {
		return nil, err
	}

	return privKey, nil
}
