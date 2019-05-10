package dummy

import (
	"github.com/sirupsen/logrus"

	"github.com/SamuelMarks/dag1/src/proxy"
)

// DummyClient is a implementation of the dummy app. DAG1 and the
// app run in separate processes and communicate through proxy
type DummyClient struct {
	logger        *logrus.Logger
	state         proxy.ProxyHandler
	dag1Proxy proxy.DAG1Proxy
}

// NewInmemDummyApp constructor
func NewInmemDummyApp(logger *logrus.Logger) proxy.AppProxy {
	state := NewState(logger)
	return proxy.NewInmemAppProxy(state, logger)
}

// NewDummySocketClient constructor
func NewDummySocketClient(addr string, logger *logrus.Logger) (*DummyClient, error) {
	dag1Proxy, err := proxy.NewGrpcDAG1Proxy(addr, logger)
	if err != nil {
		return nil, err
	}

	return NewDummyClient(dag1Proxy, nil, logger)
}

// NewDummyClient instantiates an implementation of the dummy app
func NewDummyClient(dag1Proxy proxy.DAG1Proxy, handler proxy.ProxyHandler, logger *logrus.Logger) (c *DummyClient, err error) {
	// state := NewState(logger)

	c = &DummyClient{
		logger:        logger,
		state:         handler,
		dag1Proxy: dag1Proxy,
	}

	if handler == nil {
		return
	}

	go func() {
		for {
			select {

			case b, ok := <-dag1Proxy.CommitCh():
				if !ok {
					return
				}
				logger.Debugf("block commit event: %v", b.Block)
				hash, err := handler.CommitHandler(b.Block)
				b.Respond(hash, err)

			case r, ok := <-dag1Proxy.RestoreCh():
				if !ok {
					return
				}
				logger.Debugf("snapshot restore command: %v", r.Snapshot)
				hash, err := handler.RestoreHandler(r.Snapshot)
				r.Respond(hash, err)

			case s, ok := <-dag1Proxy.SnapshotRequestCh():
				if !ok {
					return
				}
				logger.Debugf("get snapshot query: %v", s.BlockIndex)
				hash, err := handler.SnapshotHandler(s.BlockIndex)
				s.Respond(hash, err)
			}
		}
	}()

	return
}

// SubmitTx sends a transaction to node via proxy
func (c *DummyClient) SubmitTx(tx []byte) error {
	return c.dag1Proxy.SubmitTx(tx)
}
