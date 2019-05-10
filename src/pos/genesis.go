package pos

import (
	"github.com/SamuelMarks/dag1/src/common"
	"github.com/SamuelMarks/dag1/src/peers"
	"github.com/SamuelMarks/dag1/src/state"
)

// FakeGenesis is a stub
func FakeGenesis(participants *peers.Peers, conf *Config, db state.Database) (common.Hash, error) {
	if conf == nil {
		conf = DefaultConfig()
	}

	balance := conf.TotalSupply / uint64(participants.Len())

	statedb, _ := state.New(common.Hash{}, db)

	for _, p := range participants.ToPeerSlice() {
		statedb.AddBalance(p.Address(), balance)
	}
	return statedb.Commit(true)
}
