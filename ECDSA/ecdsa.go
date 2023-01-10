package ECDSA

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"
	baxos "student_22_BFT_Baxos"
	pb "student_22_BFT_Baxos/proto/BFTBaxos"
)

// crypto part for ECDSA

// Signature is an ECDSA signature
type Signature struct {
	r, s   *big.Int
	signer string
}

// Signer returns the ID of the replica that generated the signature.
func (sig Signature) Signer() string {
	return sig.signer
}

// R returns the r value of the signature
func (sig Signature) R() *big.Int {
	return sig.r
}

// S returns the s value of the signature
func (sig Signature) S() *big.Int {
	return sig.s
}

// ToBytes returns a raw byte string representation of the signature
func (sig Signature) ToBytes() []byte {
	var b []byte
	b = append(b, sig.r.Bytes()...)
	b = append(b, sig.s.Bytes()...)
	return b
}

func Sign(message []byte, priv *ecdsa.PrivateKey, signer string) (signature Signature) {
	hash := sha256.Sum256(message)
	r, s, _ := ecdsa.Sign(rand.Reader, priv, hash[:])
	return Signature{
		r:      r,
		s:      s,
		signer: signer,
	}
}

func RestoreSignature(signatures *pb.ECDSASignature) *Signature {
	sigR := new(big.Int)
	sigR.SetBytes(signatures.GetR())
	sigS := new(big.Int)
	sigS.SetBytes(signatures.GetS())
	return &Signature{
		r:      sigR,
		s:      sigS,
		signer: signatures.Signer,
	}
}

// RestoreMultiSignature should only be used to restore an existing threshold signature from a set of signatures.
func RestoreMultiSignature(signatures []*pb.ECDSASignature) MultiSignature {
	multiSig := make(MultiSignature, len(signatures))
	for _, s := range signatures {
		sigR := new(big.Int)
		sigR.SetBytes(s.GetR())
		sigS := new(big.Int)
		sigS.SetBytes(s.GetS())
		multiSig[s.Signer] = &Signature{
			r:      sigR,
			s:      sigS,
			signer: s.Signer,
		}
	}
	return multiSig
}

// VerifyMessage verifies the given quorum signature against the message.
func VerifyMessage(signature MultiSignature, message []byte, keyStore map[string]*ecdsa.PublicKey) bool {
	hash := sha256.Sum256(message)

	return VerifyHash(signature, hash[:], keyStore)
}

// VerifyHash verifies the given quorum signature against the message.
func VerifyHash(signature MultiSignature, hash []byte, keyStore map[string]*ecdsa.PublicKey) bool {
	n := signature.Participants().Len()
	if n == 0 {
		return false
	}

	results := make(chan bool, n)

	for _, sig := range signature {
		go func(sig *Signature, hash []byte) {
			results <- VerifySingleHash(sig, hash[:], keyStore[sig.Signer()])
		}(sig, hash[:])
	}

	valid := true
	for range signature {
		if !<-results {
			valid = false
		}
	}
	return valid
}

func VerifySingleMsg(sig *Signature, message []byte, pub *ecdsa.PublicKey) bool {
	hash := sha256.Sum256(message)

	return ecdsa.Verify(pub, hash[:], sig.R(), sig.S())
}

func VerifySingleHash(sig *Signature, hash []byte, pub *ecdsa.PublicKey) bool {
	return ecdsa.Verify(pub, hash[:], sig.R(), sig.S())
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

// MultiSignature is a set of (partial) signatures.
type MultiSignature map[string]*Signature

// ToBytes returns the object as bytes.
func (sig MultiSignature) ToBytes() []byte {
	var b []byte
	// sort by ID to make it deterministic
	order := make([]string, 0, len(sig))
	for _, signature := range sig {
		order = append(order, signature.signer)
	}
	//slices.Sort(order)
	//sort the string
	sort.Strings(order)

	for _, id := range order {
		b = append(b, sig[id].ToBytes()...)
	}
	return b
}

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
