package EDDSA

import (
	"crypto/ed25519"
	"fmt"
	baxos "student_22_BFT_Baxos"
	pb "student_22_BFT_Baxos/proto/BFTBaxos"
)

type Signature struct {
	sig    []byte
	signer string
}

type MultiSignature map[string]*Signature

// Participants returns the IDs of replicas who participated in the threshold signature.
func (sig MultiSignature) Participants() baxos.IDSet {
	return sig
}

// Add adds an ID to the set.
func (sig MultiSignature) Add(id string) {
	panic("not implemented")
}

// Contains returns true if the set contains the ID.
func (sig MultiSignature) Contains(id string) bool {
	_, ok := sig[id]
	return ok
}

// ForEach calls f for each ID in the set.
func (sig MultiSignature) ForEach(f func(string)) {
	for id := range sig {
		f(id)
	}
}

// RangeWhile calls f for each ID in the set until f returns false.
func (sig MultiSignature) RangeWhile(f func(string) bool) {
	for id := range sig {
		if !f(id) {
			break
		}
	}
}

// Len returns the number of entries in the set.
func (sig MultiSignature) Len() int {
	return len(sig)
}

func (sig MultiSignature) String() string {
	return baxos.IDSetToString(sig)
}

// Signer returns the ID of the replica that generated the signature.
func (sig Signature) Signer() string {
	return sig.signer
}

func (sig Signature) Sig() []byte {
	return sig.sig
}

func Sign(message []byte, priv ed25519.PrivateKey, signer string) Signature {
	sig := ed25519.Sign(priv, message[:])
	return Signature{
		sig:    sig,
		signer: signer,
	}
}

func RestoreSignature(signatures *pb.Signature) *Signature {
	return &Signature{
		sig:    signatures.Signature,
		signer: signatures.Signer,
	}
}

// RestoreMultiSignature should only be used to restore an existing threshold signature from a set of signatures.
func RestoreMultiSignature(signatures []*pb.Signature) MultiSignature {
	multiSig := make(MultiSignature, len(signatures))
	for _, s := range signatures {
		multiSig[s.Signer] = &Signature{
			sig:    s.Signature,
			signer: s.Signer,
		}
	}
	return multiSig
}

// VerifyMessage verifies the given quorum signature against the message.
func VerifyMessage(signature MultiSignature, message []byte, keyStore map[string]ed25519.PublicKey) bool {
	n := signature.Participants().Len()
	if n == 0 {
		return false
	}

	results := make(chan bool, n)

	for _, sig := range signature {
		go func(sig *Signature, hash []byte) {
			results <- VerifySingleMsg(sig, message[:], keyStore[sig.Signer()])
		}(sig, message[:])
	}

	valid := true
	for range signature {
		if !<-results {
			valid = false
		}
	}
	return valid
}

func VerifySingleMsg(sig *Signature, message []byte, pub ed25519.PublicKey) bool {
	return ed25519.Verify(pub, message[:], sig.Sig())
}

// Combine combines multiple signatures into a single signature.
func Combine(signatures []*Signature) (*MultiSignature, error) {
	if len(signatures) < 2 {
		return nil, fmt.Errorf("ErrCombineMultiple")
	}

	ts := make(MultiSignature)

	for _, sig := range signatures {
		ts[sig.signer] = sig
	}

	return &ts, nil
}
