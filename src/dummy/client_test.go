package dummy

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/SamuelMarks/dag1/src/common"
	"github.com/SamuelMarks/dag1/src/crypto"
	"github.com/SamuelMarks/dag1/src/poset"
	"github.com/SamuelMarks/dag1/src/proxy"
	"github.com/SamuelMarks/dag1/src/utils"
)

func TestSocketProxyServer(t *testing.T) {
	const (
		timeout    = 2 * time.Second
		errTimeout = "time is over"
	)
	addr := utils.GetUnusedNetAddr(1, t)
	assertO := assert.New(t)
	logger := common.NewTestLogger(t)

	txOrigin := []byte("the test transaction")

	// Server
	app, err := proxy.NewGrpcAppProxy(addr[0], timeout, logger)
	assertO.NoError(err)

	//  listens for a request
	go func() {
		select {
		case tx := <-app.SubmitCh():
			assertO.Equal(txOrigin, tx)
		case <-time.After(timeout):
			assertO.Fail(errTimeout)
		}
	}()

	// Client part connecting to RPC service and calling methods
	dag1Proxy, err := proxy.NewGrpcDAG1Proxy(addr[0], logger)
	assertO.NoError(err)

	node, err := NewDummyClient(dag1Proxy, nil, logger)
	assertO.NoError(err)

	err = node.SubmitTx(txOrigin)
	assertO.NoError(err)
}

func TestDummySocketClient(t *testing.T) {
	const (
		timeout = 2 * time.Second
	)
	addr := utils.GetUnusedNetAddr(1, t)
	assertO := assert.New(t)
	logger := common.NewTestLogger(t)

	// server
	appProxy, err := proxy.NewGrpcAppProxy(addr[0], timeout, logger)
	assertO.NoError(err)
	defer func() {
		if err := appProxy.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// client
	dag1Proxy, err := proxy.NewGrpcDAG1Proxy(addr[0], logger)
	assertO.NoError(err)
	defer func() {
		if err := dag1Proxy.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	state := NewState(logger)

	_, err = NewDummyClient(dag1Proxy, state, logger)
	assertO.NoError(err)

	initialStateHash := state.stateHash
	//create a few blocks
	blocks := [5]poset.Block{}
	for i := int64(0); i < 5; i++ {
		blocks[i] = poset.NewBlock(i, i+1, []byte{}, [][]byte{[]byte(fmt.Sprintf("block %d transaction", i))})
	}

	<-time.After(timeout / 4)

	//commit first block and check that the client's statehash is correct
	stateHash, err := appProxy.CommitBlock(blocks[0])
	assertO.NoError(err)

	expectedStateHash := crypto.Keccak256(append([][]byte{initialStateHash}, blocks[0].Transactions()...)...)

	assertO.Equal(expectedStateHash, stateHash)

	snapshot, err := appProxy.GetSnapshot(blocks[0].Index())
	assertO.NoError(err)

	assertO.Equal(expectedStateHash, snapshot)

	//commit a few more blocks, then attempt to restore back to block 0 state
	for i := 1; i < 5; i++ {
		_, err := appProxy.CommitBlock(blocks[i])
		assertO.NoError(err)
	}

	err = appProxy.Restore(snapshot)
	assertO.NoError(err)
}
