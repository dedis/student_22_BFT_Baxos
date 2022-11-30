package colory

import (
	"errors"
	"fmt"
	pb "student_22_BFT_Baxos/proto/consensus"
	"sync"
	"time"
)

// Instance represents a colory distributed color vote service instance
type Instance struct {
	mu sync.Mutex
	// begin protected fields
	name      string
	increment uint64
	timeout   time.Duration
	promised  uint64
	ballot    uint64
	color     string
	peers     []peer
	// end protected fields

	// mustEmbedUnimplementedConsensusServer()
	pb.UnimplementedConsensusServer
}

type peer struct {
	name   string
	client pb.ConsensusClient
}

var (
	// ErrDuplicatePeer is returned when peer already exists in the peer list
	ErrDuplicatePeer = errors.New("duplicate peer")
)

// New initializes a new colory instance
func New(name string, increment uint64, timeout time.Duration) *Instance {
	in := Instance{
		name:      name,
		increment: increment,
		timeout:   timeout,
	}
	fmt.Println("initialized")
	return &in
}

// AddPeer adds a new peer to the peer list
func (in *Instance) AddPeer(name string, client pb.ConsensusClient) error {
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
