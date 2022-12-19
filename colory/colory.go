package colory

import (
	"errors"
	"fmt"
	consensus "student_22_BFT_Baxos/proto/BFTBaxos"
	application "student_22_BFT_Baxos/proto/application"
	"sync"
	"time"
)

// Instance represents a colory distributed color vote service instance
type Instance struct {
	mu sync.RWMutex
	// begin protected fields
	name    string
	timeout time.Duration
	peers   []peer
	f       int
	// end protected fields

	// BFTBaxos_consensus fields
	pendingColor  string // our value
	proposedColor string // real value should be pre-proposed/proposed
	// basic
	ballot uint64
	//TODO nextBallotQC      consensus.BallotQuorumCert
	increment         uint64
	promisedBallot    uint64            // promised ballotNumber
	promiseSet        []verifiedPromise // promise set
	preAcceptedBallot uint64            // preAccepted ballotNumber
	preAcceptedColor  string            // preAccepted Color(value)
	preAcceptQC       acceptQC          // QC for prePropose color
	acceptedBallot    uint64            // accepted ballotNumber
	acceptedColor     string            // accepted Color(value)
	AcceptQC          acceptQC          // QC for propose color
	// All Consensus colors will be mixed into the same color on the canvas!
	// Said differently, we are single instance version now
	canvas       []string
	paintedCount uint64 // the count of painting
	// backoff mechanism
	retries       int64
	roundTripTime time.Duration
	retryTable    map[string]int64
	//TODO Dummy now
	K            float64
	KProof       bool
	nextBallot   uint64
	timeStampSet []timeStamp // timestamp set collected for verifying K
	// TODO ballotQuorumCet consensus.AcceptQuorumCert // QC for retry ballot
	//TODO signature part

	// mustEmbedUnimplementedConsensusServer()
	consensus.UnimplementedBFTBaxosServer
	application.UnimplementedApplicationServer
}

type peer struct {
	name   string
	client consensus.BFTBaxosClient
}

var (
	// ErrDuplicatePeer is returned when peer already exists in the peer list
	ErrDuplicatePeer = errors.New("duplicate peer")
)

// New initializes a new colory instance
func New(name string, increment uint64, timeout time.Duration) *Instance {
	//TODO define all the parameters we need at here
	in := Instance{
		name:       name,
		increment:  increment,
		timeout:    timeout,
		retryTable: make(map[string]int64),
	}
	//initialize retry table
	for _, p := range in.peers {
		in.retryTable[p.name] = 0
	}
	fmt.Println("initialized")
	return &in
}

// AddPeer adds a new peer to the peer list
func (in *Instance) AddPeer(name string, client consensus.BFTBaxosClient) error {
	in.mu.Lock()
	defer in.mu.Unlock()

	// check for duplicate peers
	for _, p := range in.peers {
		if p.name == name || p.client == client {
			return ErrDuplicatePeer
		}
	}

	// add peer to the peer list
	in.peers = append(in.peers, peer{
		name:   name,
		client: client,
	})
	fmt.Printf("added peer %v\n", name)

	return nil
}

func (in *Instance) isMajority(n int) bool {
	return n >= 2*(in.f)+1
}

// paint start the color consensus
func (in *Instance) paint() bool {
	return true
}
