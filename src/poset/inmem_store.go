package poset

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/golang-lru"

	"github.com/SamuelMarks/dag1/src/common"
	"github.com/SamuelMarks/dag1/src/kvdb"
	"github.com/SamuelMarks/dag1/src/peers"
	"github.com/SamuelMarks/dag1/src/pos"
	"github.com/SamuelMarks/dag1/src/state"
)

// InmemStore struct
type InmemStore struct {
	cacheSize              int
	participants           *peers.Peers
	eventCache             *lru.Cache           // hash => Event
	roundCreatedCache      *lru.Cache           // round number => RoundCreated
	roundReceivedCache     *lru.Cache           // round received number => RoundReceived
	blockCache             *lru.Cache           // index => Block
	frameCache             *lru.Cache           // round received => Frame
	clothoCheckCache       *lru.Cache           // frame + hash => hash
	clothoCheckCreatorCache *lru.Cache          // frame + creator => hash
	timeTableCache         *lru.Cache          // 
	consensusCache         *common.RollingIndex // consensus index => hash
	totConsensusEvents     int64
	participantEventsCache *ParticipantEventsCache // pubkey => Events
	rootsByParticipant     map[string]Root         // [participant] => Root
	rootsBySelfParent      map[EventHash]Root      // [Root.SelfParent.Hash] => Root
	lastRound              int64
	lastConsensusEvents    map[string]EventHash // [participant] => hex() of last consensus event
	lastBlock              int64

	lastRoundLocker          sync.RWMutex
	lastBlockLocker          sync.RWMutex
	totConsensusEventsLocker sync.RWMutex
	clothoCheckLocker        sync.RWMutex
	timeTableLocker          sync.RWMutex

	states    state.Database
	stateRoot common.Hash
}

// NewInmemStore constructor
func NewInmemStore(participants *peers.Peers, cacheSize int, posConf *pos.Config) *InmemStore {
	rootsByParticipant := make(map[string]Root)

	participants.RLock()
	for pk, pid := range participants.ByPubKey {
		root := NewBaseRoot(pid.ID)
		rootsByParticipant[pk] = root
	}
	participants.RUnlock()

	eventCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.eventCache:", err)
		os.Exit(31)
	}
	roundCreatedCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.roundCreatedCache:", err)
		os.Exit(32)
	}
	roundReceivedCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.roundReceivedCache:", err)
		os.Exit(35)
	}
	blockCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.blockCache:", err)
		os.Exit(33)
	}
	frameCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.frameCache:", err)
		os.Exit(34)
	}
	clothoCheckCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.checkClothoCache:", err)
		os.Exit(35)
	}
	clothoCheckCreatorCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.checkClothoCreatorCache:", err)
		os.Exit(36)
	}
	timeTableCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.timeTableCache:", err)
		os.Exit(36)
	}

	store := &InmemStore{
		cacheSize:              cacheSize,
		participants:           participants,
		eventCache:             eventCache,
		roundCreatedCache:      roundCreatedCache,
		roundReceivedCache:     roundReceivedCache,
		blockCache:             blockCache,
		frameCache:             frameCache,
		clothoCheckCache:       clothoCheckCache,
		clothoCheckCreatorCache:clothoCheckCreatorCache,
		timeTableCache:         timeTableCache,
		consensusCache:         common.NewRollingIndex("ConsensusCache", cacheSize),
		participantEventsCache: NewParticipantEventsCache(cacheSize, participants),
		rootsByParticipant:     rootsByParticipant,
		lastRound:              -1,
		lastBlock:              -1,
		lastConsensusEvents:    map[string]EventHash{},
		states: state.NewDatabase(
			kvdb.NewTable(
				kvdb.NewMemDatabase(), statePrefix)),
	}

	participants.OnNewPeer(func(peer *peers.Peer) {
		root := NewBaseRoot(peer.ID)
		store.rootsByParticipant[peer.Message.PubKeyHex] = root
		store.rootsBySelfParent = nil
		_ = store.RootsBySelfParent()
		old := store.participantEventsCache
		store.participantEventsCache = NewParticipantEventsCache(cacheSize, participants)
		store.participantEventsCache.Import(old)
	})

	// TODO: replace with real genesis
	store.stateRoot, err = pos.FakeGenesis(participants, posConf, store.states)
	if err != nil {
		fmt.Println("Unable to init genesis state:", err)
		os.Exit(36)
	}

	return store
}

/*
 * Store interface implementation:
 */

// TopologicalEvents returns event in topological order.
func (s *InmemStore) TopologicalEvents() ([]Event, error) {
	// NOTE: it's used for bootstrap only, so is not implemented
	return nil, nil
}

// CacheSize size of cache
func (s *InmemStore) CacheSize() int {
	return s.cacheSize
}

// Participants returns participants
func (s *InmemStore) Participants() (*peers.Peers, error) {
	return s.participants, nil
}

// RootsBySelfParent retrieve EventHash map of roots
func (s *InmemStore) RootsBySelfParent() map[EventHash]Root {
	if s.rootsBySelfParent == nil {
		s.rootsBySelfParent = make(map[EventHash]Root)
		for _, root := range s.rootsByParticipant {
			var hash EventHash
			hash.Set(root.SelfParent.Hash)
			s.rootsBySelfParent[hash] = root
		}
	}
	return s.rootsBySelfParent
}

// RootsByParticipant retrieve PubKeyHex map of roots
func (s *InmemStore) RootsByParticipant() map[string]Root {
	return s.rootsByParticipant
}

// GetEventBlock gets specific event block by hash
func (s *InmemStore) GetEventBlock(hash EventHash) (Event, error) {
	res, ok := s.eventCache.Get(hash)
	if !ok {
		return Event{}, common.NewStoreErr("EventCache", common.KeyNotFound, hash.String())
	}

	return res.(Event), nil
}

// SetEvent set event for event block
func (s *InmemStore) SetEvent(event Event) error {
	eventHash := event.Hash()
	_, err := s.GetEventBlock(eventHash)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	if common.Is(err, common.KeyNotFound) {
		if err := s.addParticipantEvent(event.GetCreator(), eventHash, event.Index()); err != nil {
			return err
		}
	}

	// fmt.Println("Adding event to cache", event.Hex())
	s.eventCache.Add(eventHash, event)

	return nil
}

func (s *InmemStore) addParticipantEvent(participant string, hash EventHash, index int64) error {
	return s.participantEventsCache.Set(participant, hash, index)
}

// ParticipantEvents events for the participant
func (s *InmemStore) ParticipantEvents(participant string, skip int64) (EventHashes, error) {
	return s.participantEventsCache.Get(participant, skip)
}

// ParticipantEvent specific event
func (s *InmemStore) ParticipantEvent(participant string, index int64) (hash EventHash, err error) {
	hash, err = s.participantEventsCache.GetItem(participant, index)
	if err == nil {
		return
	}

	root, ok := s.rootsByParticipant[participant]
	if !ok {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
		return
	}

	if root.SelfParent.Index == index {
		hash.Set(root.SelfParent.Hash)
		err = nil
	}
	return
}

// LastEventFrom participant
func (s *InmemStore) LastEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	// try to get the last event from this participant
	last, err = s.participantEventsCache.GetLast(participant)
	if err == nil || !common.Is(err, common.Empty) {
		return
	}
	// if there is none, grab the root
	if root, ok := s.rootsByParticipant[participant]; ok {
		last.Set(root.SelfParent.Hash)
		isRoot = true
		err = nil
	} else {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
	}
	return
}

// LastConsensusEventFrom participant
func (s *InmemStore) LastConsensusEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	// try to get the last consensus event from this participant
	last, ok := s.lastConsensusEvents[participant]
	if ok {
		return
	}
	// if there is none, grab the root
	root, ok := s.rootsByParticipant[participant]
	if ok {
		last.Set(root.SelfParent.Hash)
		isRoot = true
	} else {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
	}

	return
}

// ConsensusEvents returns all consensus events
func (s *InmemStore) ConsensusEvents() EventHashes {
	lastWindow, _ := s.consensusCache.GetLastWindow()
	res := make(EventHashes, len(lastWindow))
	for i, item := range lastWindow {
		res[i] = item.(EventHash)
	}
	return res
}

// ConsensusEventsCount returns count of all consnesus events
func (s *InmemStore) ConsensusEventsCount() int64 {
	s.totConsensusEventsLocker.RLock()
	defer s.totConsensusEventsLocker.RUnlock()
	return s.totConsensusEvents
}

// AddConsensusEvent to store
func (s *InmemStore) AddConsensusEvent(event Event) error {
	s.totConsensusEventsLocker.Lock()
	defer s.totConsensusEventsLocker.Unlock()
	err := s.consensusCache.Set(event.Hash(), s.totConsensusEvents)
	if err != nil {
		return err
	}
	s.totConsensusEvents++
	s.lastConsensusEvents[event.GetCreator()] = event.Hash()
	return nil
}

// GetRoundCreated retrieves created round by ID
func (s *InmemStore) GetRoundCreated(r int64) (RoundCreated, error) {
	res, ok := s.roundCreatedCache.Get(r)
	if !ok {
		return *NewRoundCreated(), common.NewStoreErr("RoundCreatedCache", common.KeyNotFound, strconv.FormatInt(r, 10))
	}
	return res.(RoundCreated), nil
}

// SetRoundCreated stores created round by ID
func (s *InmemStore) SetRoundCreated(r int64, round RoundCreated) error {
	s.lastRoundLocker.Lock()
	defer s.lastRoundLocker.Unlock()
	s.roundCreatedCache.Add(r, round)
	if r > s.lastRound {
		s.lastRound = r
	}
	return nil
}

// GetRoundReceived gets received round by ID
func (s *InmemStore) GetRoundReceived(r int64) (RoundReceived, error) {
	res, ok := s.roundReceivedCache.Get(r)
	if !ok {
		return *NewRoundReceived(), common.NewStoreErr("RoundReceivedCache", common.KeyNotFound, strconv.FormatInt(r, 10))
	}
	return res.(RoundReceived), nil
}

// SetRoundReceived stores received round by ID
func (s *InmemStore) SetRoundReceived(r int64, round RoundReceived) error {
	s.lastRoundLocker.Lock()
	defer s.lastRoundLocker.Unlock()
	s.roundReceivedCache.Add(r, round)
	if r > s.lastRound {
		s.lastRound = r
	}
	return nil
}

// LastRound getter
func (s *InmemStore) LastRound() int64 {
	s.lastRoundLocker.RLock()
	defer s.lastRoundLocker.RUnlock()
	return s.lastRound
}

// RoundClothos all clothos for the specified round
func (s *InmemStore) RoundClothos(r int64) EventHashes {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return EventHashes{}
	}
	return round.Clotho()
}

// RoundEvents returns events for the round
func (s *InmemStore) RoundEvents(r int64) int {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return 0
	}
	return len(round.Message.Events)
}

// GetRoot for participant
func (s *InmemStore) GetRoot(participant string) (Root, error) {
	res, ok := s.rootsByParticipant[participant]
	if !ok {
		return Root{}, common.NewStoreErr("RootCache", common.KeyNotFound, participant)
	}
	return res, nil
}

// GetBlock for index
func (s *InmemStore) GetBlock(index int64) (Block, error) {
	res, ok := s.blockCache.Get(index)
	if !ok {
		return Block{}, common.NewStoreErr("BlockCache", common.KeyNotFound, strconv.FormatInt(index, 10))
	}
	return res.(Block), nil
}

// SetBlock TODO
func (s *InmemStore) SetBlock(block Block) error {
	s.lastBlockLocker.Lock()
	defer s.lastBlockLocker.Unlock()
	index := block.Index()
	_, err := s.GetBlock(index)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	s.blockCache.Add(index, block)
	if index > s.lastBlock {
		s.lastBlock = index
	}
	return nil
}

// LastBlockIndex getter
func (s *InmemStore) LastBlockIndex() int64 {
	s.lastBlockLocker.RLock()
	defer s.lastBlockLocker.RUnlock()
	return s.lastBlock
}

// GetFrame by index
func (s *InmemStore) GetFrame(index int64) (Frame, error) {
	res, ok := s.frameCache.Get(index)
	if !ok {
		return Frame{}, common.NewStoreErr("FrameCache", common.KeyNotFound, strconv.FormatInt(index, 10))
	}
	return res.(Frame), nil
}

// SetFrame in the store
func (s *InmemStore) SetFrame(frame Frame) error {
	index := frame.Round
	_, err := s.GetFrame(index)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	s.frameCache.Add(index, frame)
	return nil
}

// Reset resets the store
func (s *InmemStore) Reset(roots map[string]Root) error {
	eventCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.eventCache:", errr)
		os.Exit(41)
	}
	roundCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.roundCreatedCache:", errr)
		os.Exit(42)
	}
	roundReceivedCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.roundReceivedCache:", errr)
		os.Exit(45)
	}
	clothoCheckCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.clothoCheckCache:", errr)
		os.Exit(46)
	}
	clothoCheckCreatorCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.clothoCheckCreatorCache:", errr)
		os.Exit(47)
	}
	timeTableCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.timeTableCache:", errr)
		os.Exit(48)
	}
	// FIXIT: Should we recreate blockCache, frameCache and participantEventsCache here as well
	//        and reset lastConsensusEvents ?
	s.rootsByParticipant = roots
	s.rootsBySelfParent = nil
	_ = s.RootsBySelfParent()
	s.eventCache = eventCache
	s.roundCreatedCache = roundCache
	s.roundReceivedCache = roundReceivedCache
	s.clothoCheckCache = clothoCheckCache
	s.clothoCheckCreatorCache = clothoCheckCreatorCache
	s.timeTableCache = timeTableCache
	s.consensusCache = common.NewRollingIndex("ConsensusCache", s.cacheSize)
	err := s.participantEventsCache.Reset()
	s.lastRoundLocker.Lock()
	s.lastRound = -1
	s.lastRoundLocker.Unlock()
	s.lastBlockLocker.Lock()
	s.lastBlock = -1
	s.lastBlockLocker.Unlock()

	return err
}

// Close the store
func (s *InmemStore) Close() error {
	return nil
}

// NeedBootstrap for the store
func (s *InmemStore) NeedBootstrap() bool {
	return false
}

// StorePath getter
func (s *InmemStore) StorePath() string {
	return ""
}

// StateDB returns state database
func (s *InmemStore) StateDB() state.Database {
	return s.states
}

// StateRoot returns genesis state hash.
func (s *InmemStore) StateRoot() common.Hash {
	return s.stateRoot
}

func checkClothoKeyStr(frame int64, hash EventHash) string {
	return fmt.Sprintf("%09d_%s", frame, hash.String())
}

func checkClothoCreatorKeyStr(frame int64, CreatorID uint64) string {
	return fmt.Sprintf("%09d_%d", frame, CreatorID)
}

// AddClothoCheck to store
func (s *InmemStore) AddClothoCheck(frame int64, creatorID uint64, hash EventHash) error {
	key := checkClothoKeyStr(frame, hash)
	s.clothoCheckLocker.Lock()
	defer s.clothoCheckLocker.Unlock()
	s.clothoCheckCache.Add(key, hash.Bytes())
	key = checkClothoCreatorKeyStr(frame, creatorID)
	s.clothoCheckCreatorCache.Add(key, hash.Bytes())
	return nil
}

// GetClothoCheck retrieves EventHash by frame + hash
func (s *InmemStore) GetClothoCheck(frame int64, phash EventHash) (hash EventHash, err error) {
	key := checkClothoKeyStr(frame, phash)
	s.clothoCheckLocker.Lock()
	defer s.clothoCheckLocker.Unlock()
	res, ok := s.clothoCheckCache.Get(key)
	if !ok {
		return EventHash{}, common.NewStoreErr("ClothoCheckCache", common.KeyNotFound, string(key))
	}
	hash.Set(res.([]byte))
	return hash, nil
}

// GetClothoCreatorCheck retrieves EventHash by frame + creatorID
func (s *InmemStore) GetClothoCreatorCheck(frame int64, creatorID uint64) (hash EventHash, err error) {
	key := checkClothoCreatorKeyStr(frame, creatorID)
	s.clothoCheckLocker.Lock()
	defer s.clothoCheckLocker.Unlock()
	res, ok := s.clothoCheckCreatorCache.Get(key)
	if !ok {
		return EventHash{}, common.NewStoreErr("ClothoCheckCreatorCache", common.KeyNotFound, string(key))
	}
	hash.Set(res.([]byte))
	return hash, nil
}

func timeTableKeyStr(hash EventHash) string {
	return fmt.Sprintf("timeTable_%s", hash.String())
}

// AddTimeTable adds lamport timestamp for pair of events for voting in atropos time selection
func (s *InmemStore) AddTimeTable(hashTo EventHash, hashFrom EventHash, lamportTime int64) error {
	ft := NewFlagTable()
	key := timeTableKeyStr(hashTo)
	s.timeTableLocker.Lock()
	defer s.timeTableLocker.Unlock()
	res, ok := s.timeTableCache.Get(key)
	if ok {
		ft.Unmarshal(res.([]byte))
	}
	ft[hashFrom] = lamportTime
	res = ft.Marshal()
	s.timeTableCache.Add(key, res)
	return nil
}

// GetTimeTable retrieve FlagTable with lamport time votes in atropos time selection for specified EventHash
func (s *InmemStore) GetTimeTable(hash EventHash) (FlagTable, error) {
	ft := NewFlagTable()
	key := timeTableKeyStr(hash)
	s.timeTableLocker.RLock()
	defer s.timeTableLocker.RUnlock()
	res, ok := s.timeTableCache.Get(key)
	if !ok {
		return nil, common.NewStoreErr("GetTimeTableCache", common.KeyNotFound, string(key))
	}
	ft.Unmarshal(res.([]byte))
	return ft, nil
}

// This is just a stub, yet to bee implemented if needed
func (s *InmemStore) CheckFrameFinality(frame int64) bool {
	return true
}

// This is just a stub, yet to bee implemented if needed
func (s *InmemStore) ProcessOutFrame(frame int64, address string) ([][]byte, error) {
	return nil, nil
}
