package student_22_BFT_Baxos

import "strings"

// IDSet implements a set of replica IDs. It is used to show which replicas participated in some event.
type IDSet interface {
	// Add adds an ID to the set.
	Add(id string)
	// Contains returns true if the set contains the ID.
	Contains(id string) bool
	// ForEach calls f for each ID in the set.
	ForEach(f func(string))
	// RangeWhile calls f for each ID in the set until f returns false.
	RangeWhile(f func(string) bool)
	// Len returns the number of entries in the set.
	Len() int
}

// idSetMap implements IDSet using a map.
type idSetMap map[string]struct{}

// NewIDSet returns a new IDSet using the default implementation.
func NewIDSet() IDSet {
	return make(idSetMap)
}

// Add adds an ID to the set.
func (s idSetMap) Add(id string) {
	s[id] = struct{}{}
}

// Contains returns true if the set contains the given ID.
func (s idSetMap) Contains(id string) bool {
	_, ok := s[id]
	return ok
}

// ForEach calls f for each ID in the set.
func (s idSetMap) ForEach(f func(string)) {
	for id := range s {
		f(id)
	}
}

// RangeWhile calls f for each ID in the set until f returns false.
func (s idSetMap) RangeWhile(f func(string) bool) {
	for id := range s {
		if !f(id) {
			break
		}
	}
}

// Len returns the number of entries in the set.
func (s idSetMap) Len() int {
	return len(s)
}

func (s idSetMap) String() string {
	return IDSetToString(s)
}

// IDSetToString formats an IDSet as a string.
func IDSetToString(set IDSet) string {
	var sb strings.Builder
	sb.WriteString("[ ")
	set.ForEach(func(i string) {
		sb.WriteString(string(i))
		sb.WriteString(" ")
	})
	sb.WriteString("]")
	return sb.String()
}

// QuorumSignature is a signature that is only valid when it contains the signatures of a quorum of replicas.
type QuorumSignature interface {
	// ToBytes returns the object as bytes.
	ToBytes() []byte
	// Participants returns the IDs of replicas who participated in the threshold signature.
	Participants() IDSet
}
