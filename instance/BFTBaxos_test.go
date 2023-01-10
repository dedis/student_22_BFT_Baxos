package instance

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"student_22_BFT_Baxos/EDDSA"
	consensus "student_22_BFT_Baxos/proto/BFTBaxos"
	"testing"
)

func newMockNode(name string) *Instance {
	in := Instance{
		name:       "Acceptor",
		retryTable: make(map[string]int64),
		KeyStore: KeyStore{
			PublicKey: make(map[string]ed25519.PublicKey),
		},
		f: 1,
	}

	privateKey := EDDSA.GenerateEDDSAPrivateKey()
	publicKey := privateKey.Public().(ed25519.PublicKey)

	in.KeyStore.PrivateKey = privateKey
	in.KeyStore.PublicKey[name] = publicKey

	return &in

}

func updateMockInstanceKey(in *Instance, name string, publicKey ed25519.PublicKey) *Instance {
	in.KeyStore.PublicKey[name] = publicKey
	return in
}

func TestInstance_PromiseRPC(t *testing.T) {
	t.Run("promise", func(t *testing.T) {

		in := newMockNode("acceptor")

		resp, err := in.Promise(context.Background(), &consensus.PrepareMsg{
			From:   "Proposer",
			Ballot: 1,
		})
		if err != nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}

		verifiedPromise, valid := in.getVerifiedPromise(resp)
		if valid != true {
			t.Errorf("expected `%v`, got `%v`", true, valid)
		}

		if err != nil {
			t.Fatalf("expected `%v`, got `%v`", nil, err)
		}
		if verifiedPromise.contention {
			t.Errorf("expected `%v`, got `%v`", true, verifiedPromise.contention)
		}
		if verifiedPromise.ballot != 1 {
			t.Errorf("expected `%v`, got `%v`", 1, verifiedPromise.ballot)
		}
		if verifiedPromise.ballotPC.Type != 0 {
			t.Errorf("expected `%v`, got `%v`", uint64(0), verifiedPromise.ballotPC.Type)
		}
		if verifiedPromise.timeStamp.timestamp <= int64(0) {
			t.Errorf("expected greater than 0, got `%v`", verifiedPromise.timeStamp.timestamp)
		}
		if verifiedPromise.preAcceptQC != nil {
			t.Errorf("expected nil, got `%v`", verifiedPromise.preAcceptQC)
		}
		if verifiedPromise.acceptedValue != "" {
			t.Errorf("expected `%v`, got `%v`", "", verifiedPromise.acceptedValue)
		}
	})
}

func TestInstance_PreAcceptAcceptRPC(t *testing.T) {
	t.Run("preAccept without promiseSet", func(t *testing.T) {

		//crypto component
		proposerPrivateKey := EDDSA.GenerateEDDSAPrivateKey()
		proposerPublicKey := proposerPrivateKey.Public().(ed25519.PublicKey)

		in1 := newMockNode("1")
		updateMockInstanceKey(in1, "proposer", proposerPublicKey)

		// phase 1
		_, err := in1.Promise(context.Background(), &consensus.PrepareMsg{
			From:   "Proposer",
			Ballot: 1,
		})
		if err != nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}

		// phase 2 without carrying PromiseSet
		resp, err := in1.PreAccept(context.Background(), &consensus.PreProposeMsg{
			From:            "Proposer",
			Ballot:          1,
			PreProposeValue: "value",
			PromiseSet:      nil,
		})
		if err == nil {
			t.Errorf("expected `%v`, got `%v`", errors.New("invalid promise set"), err)
		}
		if resp != nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}
	})

	t.Run("preAccept with no quorum promiseSet", func(t *testing.T) {

		//crypto component
		proposerPrivateKey := EDDSA.GenerateEDDSAPrivateKey()
		proposerPublicKey := proposerPrivateKey.Public().(ed25519.PublicKey)

		in1 := newMockNode("1")
		updateMockInstanceKey(in1, "proposer", proposerPublicKey)

		in2 := newMockNode("2")
		updateMockInstanceKey(in2, "proposer", proposerPublicKey)

		updateMockInstanceKey(in1, "2", in2.KeyStore.PublicKey["2"])
		updateMockInstanceKey(in2, "1", in2.KeyStore.PublicKey["1"])

		// phase 1
		p1, err := in1.Promise(context.Background(), &consensus.PrepareMsg{
			From:   "Proposer",
			Ballot: 1,
		})

		// phase 2 with carrying PromiseSet
		resp, err := in1.PreAccept(context.Background(), &consensus.PreProposeMsg{
			From:            "Proposer",
			Ballot:          1,
			PreProposeValue: "value",
			PromiseSet: []*consensus.PromiseMsg{
				// proposer promise message
				{
					Contention:     p1.Contention,
					Ballot:         p1.Ballot,
					Accepted:       false,
					BallotPartCert: p1.BallotPartCert,
				},
				// p1 promise message
				{
					Contention:     p1.Contention,
					Ballot:         p1.Ballot,
					Accepted:       false,
					BallotPartCert: p1.BallotPartCert,
				},
				// we miss a promise message at here
			},
		})
		if err == nil {
			t.Errorf("expected `%v`, got `%v`", errors.New("invalid promise set"), err)
		}
		if resp != nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}
	})

	t.Run("preAccept with valid promiseSet", func(t *testing.T) {

		//crypto component
		proposerPrivateKey := EDDSA.GenerateEDDSAPrivateKey()
		proposerPublicKey := proposerPrivateKey.Public().(ed25519.PublicKey)
		// next ballot qc
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], 1)
		// sign this ballot
		sig := EDDSA.Sign(buf[:], proposerPrivateKey, "proposer")
		//get the partial certificate
		proposerBalotPC := &consensus.PartialCert{
			Sig: &consensus.Signature{
				Signer:    sig.Signer(),
				Signature: sig.Sig(),
			},
		}

		in1 := newMockNode("1")
		updateMockInstanceKey(in1, "proposer", proposerPublicKey)

		in2 := newMockNode("2")
		updateMockInstanceKey(in2, "proposer", proposerPublicKey)

		updateMockInstanceKey(in1, "2", in2.KeyStore.PublicKey["2"])
		updateMockInstanceKey(in2, "1", in2.KeyStore.PublicKey["1"])

		// phase 1
		p1, err := in1.Promise(context.Background(), &consensus.PrepareMsg{
			From:   "Proposer",
			Ballot: 1,
		})

		p2, err := in2.Promise(context.Background(), &consensus.PrepareMsg{
			From:   "Proposer",
			Ballot: 1,
		})
		//
		// phase 2 with carrying PromiseSet
		resp, err := in1.PreAccept(context.Background(), &consensus.PreProposeMsg{
			From:            "Proposer",
			Ballot:          1,
			PreProposeValue: "value",
			PromiseSet: []*consensus.PromiseMsg{
				// proposer promise message
				{
					Contention:     false,
					Ballot:         1,
					Accepted:       false,
					BallotPartCert: proposerBalotPC,
				},
				// p1 promise message
				{
					Contention:     p1.Contention,
					Ballot:         p1.Ballot,
					Accepted:       false,
					BallotPartCert: p1.BallotPartCert,
				},
				// p2 promise message
				{
					Contention:     p2.Contention,
					Ballot:         p2.Ballot,
					Accepted:       false,
					BallotPartCert: p2.BallotPartCert,
				},
			},
		})
		if err != nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}
		if resp == nil {
			t.Errorf("expected `%v`, got `%v`", nil, err)
		}
	})
}
