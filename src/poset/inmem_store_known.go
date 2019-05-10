// +build test

package poset

// KnownEvents returns all known events
func (s *InmemStore) KnownEvents() map[uint64]int64 {
	known := s.participantEventsCache.Known()
	s.participants.RLock()
	defer s.participants.RUnlock()
	for p, pid := range s.participants.ByPubKey {
		if known[pid.ID] == -1 {
			root, ok := s.rootsByParticipant[p]
			if ok {
				known[pid.ID] = root.SelfParent.Index
			}
		}
	}
	return known
}

