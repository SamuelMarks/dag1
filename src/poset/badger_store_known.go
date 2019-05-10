// +build test

package poset

// KnownEvents returns all known events
func (s *BadgerStore) KnownEvents() map[uint64]int64 {
	known := make(map[uint64]int64)
	s.participants.RLock()
	defer s.participants.RUnlock()
	for p, pid := range s.participants.ByPubKey {
		index := int64(-1)
		last, isRoot, err := s.LastEventFrom(p)
		if err == nil {
			if isRoot {
				root, err := s.GetRoot(p)
				if err != nil {
					index = root.SelfParent.Index
				}
			} else {
				lastEvent, err := s.GetEventBlock(last)
				if err == nil {
					index = lastEvent.Index()
				}
			}

		}
		known[pid.ID] = index
	}
	return known
}
