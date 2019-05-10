package dummy

import (
	"testing"

	"github.com/SamuelMarks/dag1/src/common"
	"github.com/SamuelMarks/dag1/src/proxy"
)

func TestProxyHandlerImplementation(t *testing.T) {
	logger := common.NewTestLogger(t)

	state := interface{}(
		NewState(logger))

	_, ok := state.(proxy.ProxyHandler)
	if !ok {
		t.Fatal("State does not implement ProxyHandler interface!")
	}
}
