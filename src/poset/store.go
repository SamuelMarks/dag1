// +build !debug

package poset

import (
	"github.com/SamuelMarks/dag1/src/common"
	"github.com/SamuelMarks/dag1/src/peers"
	"github.com/SamuelMarks/dag1/src/state"
)

// Store provides an interface for persistent and non-persistent stores
// to store key dag1 consensus information on a node.
type Store interface {
	TopologicalEvents() ([]Event, error) // returns event in topological order
	CacheSize() int
	Participants() (*peers.Peers, error)
	RootsBySelfParent() map[EventHash]Root
	RootsByParticipant() map[string]Root
	GetEventBlock(EventHash) (Event, error)
	SetEvent(Event) error
	ParticipantEvents(string, int64) (EventHashes, error)
	ParticipantEvent(string, int64) (EventHash, error)
	LastEventFrom(string) (EventHash, bool, error)
	LastConsensusEventFrom(string) (EventHash, bool, error)
	ConsensusEvents() EventHashes
	ConsensusEventsCount() int64
	AddConsensusEvent(Event) error
	GetRoundCreated(int64) (RoundCreated, error)
	SetRoundCreated(int64, RoundCreated) error
	GetRoundReceived(int64) (RoundReceived, error)
	SetRoundReceived(int64, RoundReceived) error
	LastRound() int64
	RoundClothos(int64) EventHashes
	RoundEvents(int64) int
	GetRoot(string) (Root, error)
	GetBlock(int64) (Block, error)
	SetBlock(Block) error
	LastBlockIndex() int64
	GetFrame(int64) (Frame, error)
	SetFrame(Frame) error
	Reset(map[string]Root) error
	Close() error
	NeedBootstrap() bool // Was the store loaded from existing db
	StorePath() string
	GetClothoCheck(int64, EventHash) (EventHash, error)
	GetClothoCreatorCheck(int64, uint64) (EventHash, error)
	AddClothoCheck(int64, uint64, EventHash) error
	AddTimeTable(EventHash, EventHash, int64) error
	GetTimeTable(EventHash) (FlagTable, error)
	// StateDB returns state database
	StateDB() state.Database
	StateRoot() common.Hash
	CheckFrameFinality(int64) bool
	ProcessOutFrame(int64, string) ([][]byte, error)
}
