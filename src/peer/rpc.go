package peer

import (
	"time"
)

// RPC Methods.
const (
	MethodSync        = "DAG1.Sync"
	MethodForceSync   = "DAG1.ForceSync"
	MethodFastForward = "DAG1.FastForward"
)

// DAG1 implements DAG1 synchronization methods.
type DAG1 struct {
	done           chan struct{}
	receiver       chan *RPC
	processTimeout time.Duration
	receiveTimeout time.Duration
}

// NewDAG1 creates new DAG1 RPC handler.
func NewDAG1(done chan struct{}, receiver chan *RPC,
	receiveTimeout, processTimeout time.Duration) *DAG1 {
	return &DAG1{
		done:           done,
		receiver:       receiver,
		processTimeout: processTimeout,
		receiveTimeout: receiveTimeout,
	}
}

// Sync handles sync requests.
func (r *DAG1) Sync(
	req *SyncRequest, resp *SyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*SyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// ForceSync handles force sync requests.
func (r *DAG1) ForceSync(
	req *ForceSyncRequest, resp *ForceSyncResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*ForceSyncResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

// FastForward handles fast forward requests.
func (r *DAG1) FastForward(
	req *FastForwardRequest, resp *FastForwardResponse) error {
	result, err := r.process(req)
	if err != nil {
		return err
	}

	item, ok := result.(*FastForwardResponse)
	if !ok {
		return ErrBadResult
	}
	*resp = *item
	return nil
}

func (r *DAG1) send(req interface{}) *RPCResponse {
	reply := make(chan *RPCResponse, 1) // Buffered.
	ticket := &RPC{
		Command:  req,
		RespChan: reply,
	}

	timer := time.NewTimer(r.receiveTimeout)

	select {
	case r.receiver <- ticket:
	case <-timer.C:
		return &RPCResponse{Error: ErrReceiverIsBusy}
	case <-r.done:
		return &RPCResponse{Error: ErrTransportStopped}
	}

	var result *RPCResponse

	timer.Reset(r.processTimeout)

	select {
	case result = <-reply:
	case <-timer.C:
		result = &RPCResponse{Error: ErrProcessingTimeout}
	case <-r.done:
		return &RPCResponse{Error: ErrTransportStopped}
	}

	return result
}

func (r *DAG1) process(req interface{}) (resp interface{}, err error) {
	result := r.send(req)
	if result.Error != nil {
		return nil, result.Error
	}
	resp = result.Response
	return
}
