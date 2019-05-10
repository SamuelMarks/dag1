package proxy

import (
	"github.com/SamuelMarks/dag1/src/poset"
	"github.com/SamuelMarks/dag1/src/proxy/proto"
)

// AppProxy provides an interface for dag1 to communicate
// with the application.
type AppProxy interface {
	SubmitCh() chan []byte
	SubmitInternalCh() chan poset.InternalTransaction
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int64) ([]byte, error)
	Restore(snapshot []byte) error
}

// DAG1Proxy provides an interface for the application to
// submit transactions to the dag1 node.
type DAG1Proxy interface {
	CommitCh() chan proto.Commit
	SnapshotRequestCh() chan proto.SnapshotRequest
	RestoreCh() chan proto.RestoreRequest
	SubmitTx(tx []byte) error
}
