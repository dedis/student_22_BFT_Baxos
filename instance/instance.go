package instance

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	consensus "student_22_BFT_Baxos/proto/BFTBaxos"
	"student_22_BFT_Baxos/proto/application"
	"sync"
	"time"
)

// Instance represents a instance distributed color vote service instance
type Instance struct {
	mu sync.RWMutex
	// begin protected fields
	name     string
	timeout  time.Duration
	peers    []peer
	KeyStore KeyStore
	f        int
	// end protected fields

	// BFTBaxos_consensus fields
	pendingValue  string // our value
	proposedValue string // real value should be pre-proposed/proposed
	// basic
	nextBallotQC      *QuorumCertificate
	increment         uint64
	prepareBallot     uint64
	promisedBallot    uint64             // promised ballotNumber
	promiseSet        []*verifiedPromise // promise set
	preAcceptedBallot uint64             // preAccepted ballotNumber
	preAcceptedValue  string             // preAccepted Color(value)
	preAcceptQC       *QuorumCertificate // QC for prePropose color
	acceptedBallot    uint64             // accepted ballotNumber
	acceptedValue     string             // accepted Color(value)
	acceptQC          *QuorumCertificate // QC for propose color
	committed         []string           // contain the value which is achieved by in consensus
	consensusCount    uint64             // the count of painting
	// backoff mechanism
	retries       int64
	roundTripTime time.Duration
	retryTable    map[string]int64
	k             float64

	timeStampSet []*timeStamp // timestamp set collected for verifying K

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

// New initializes a new instance instance
func New(name string, increment uint64, timeout time.Duration, f int) *Instance {
	in := Instance{
		name:       name,
		increment:  increment,
		timeout:    timeout,
		retryTable: make(map[string]int64),
		// initialize the keyStore
		KeyStore: KeyStore{
			PublicKey: make(map[string]ed25519.PublicKey),
		},
		f:             f,
		roundTripTime: time.Second,
	}
	// initialize retry table
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
	//fmt.Printf("added peer %v\n", name)

	return nil
}

func (in *Instance) GetPeer() []peer {
	return in.peers
}

// paint start the color consensus
func (in *Instance) paint() bool {
	return true
}

func (in *Instance) GetName() string {
	return in.name
}

// KeyStore store the private key and public keys
type KeyStore struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  map[string]ed25519.PublicKey
}

func (in *Instance) isMajority(n int) bool {
	if n >= 2*(in.f)+1 {
		return true
	}
	return false
}
