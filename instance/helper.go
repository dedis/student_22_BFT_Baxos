package instance

import (
	"encoding/binary"
	"fmt"
	"sort"
	"student_22_BFT_Baxos/EDDSA"
	pb "student_22_BFT_Baxos/proto/BFTBaxos"
	"time"
)

// phase 1 data structure
func (in *Instance) getPrepareMsg(retry bool, K float64, KProof []byte) pb.PrepareMsg {

	if !retry {
		return pb.PrepareMsg{
			From:   in.name,
			Ballot: in.promisedBallot,
		}
	}

	var timeStamps []*pb.BFTTimeStamp
	for _, t := range in.timeStampSet {
		timeStamps = append(timeStamps, in.getTimeStamp(t))
	}
	//retry
	return pb.PrepareMsg{
		From:      in.name,
		Ballot:    in.promisedBallot,
		K:         K,
		KProof:    KProof,
		TimeStamp: timeStamps,
		BallotQC:  in.getQC(*in.nextBallotQC),
	}
}

// verified simple version promise message
type verifiedPromise struct {
	contention     bool
	ballot         uint64
	accepted       bool
	acceptedBallot uint64
	acceptedValue  string
	preAcceptQC    *QuorumCertificate
	timeStamp      *timeStamp
	ballotPC       *PartCertificate
}

func (in *Instance) getVerifiedPromise(resp *pb.PromiseMsg) (*verifiedPromise, bool) {
	// check if the AcceptQuorumVert is valid
	if resp.Accepted {
		// restore the multiSignature
		multiSig := EDDSA.RestoreMultiSignature(resp.GetPreAcceptQC().GetSigs())
		// verify if the qc is valid
		msg := ByteTuple(uint64(1), resp.GetAcceptedBallot(), resp.GetAcceptedValue())
		if !EDDSA.VerifyMessage(multiSig, msg, in.KeyStore.PublicKey) {
			return nil, false
		}
		vPromise := verifiedPromise{
			contention:     resp.GetContention(),
			ballot:         resp.GetBallot(),
			accepted:       resp.GetAccepted(),
			acceptedBallot: resp.GetAcceptedBallot(),
			acceptedValue:  resp.GetAcceptedValue(),
			//verified
			preAcceptQC: RestoreQC(resp.GetPreAcceptQC()),
			timeStamp:   RestoreTimeStamp(resp.GetTimeStamp()),
			ballotPC:    RestorePC(resp.BallotPartCert),
		}
		return &vPromise, true
	}
	vPromise := verifiedPromise{
		contention: resp.GetContention(),
		ballot:     resp.GetBallot(),
		accepted:   resp.GetAccepted(),
		//verified
		timeStamp: RestoreTimeStamp(resp.GetTimeStamp()),
		ballotPC:  RestorePC(resp.BallotPartCert),
	}
	return &vPromise, true
}

func (in *Instance) getPromiseSet() []*pb.PromiseMsg {
	//the pointer of array
	var promiseSet []*pb.PromiseMsg

	for _, p := range in.promiseSet {
		msg := pb.PromiseMsg{
			Contention:     p.contention,
			Ballot:         p.ballot,
			Accepted:       p.accepted,
			TimeStamp:      in.getTimeStamp(p.timeStamp),
			BallotPartCert: in.getPC(*p.ballotPC),
		}
		if msg.Accepted {
			msg.AcceptedValue = &p.acceptedValue
			msg.PreAcceptQC = in.getQC(*p.preAcceptQC)
		}
		promiseSet = append(promiseSet, &msg)
	}
	return promiseSet
}

func (in *Instance) checkPromiseSet(
	ballot uint64,
	prProposeValue string,
	resp []*pb.PromiseMsg) bool {

	// promiseSet is not quorum size
	if !in.isMajority(len(resp)) {
		return false
	}

	var verifiedPromise []*verifiedPromise
	highestBallot := uint64(0)
	highestValue := ""

	//fmt.Println("promise set size: ", len(resp))

	for _, r := range resp {
		v, ok := in.getVerifiedPromise(r)
		//check promise ballot
		if ok && ballot == v.ballot {
			verifiedPromise = append(verifiedPromise, v)
			// record the highest
			if highestBallot < v.acceptedBallot {
				highestValue = v.acceptedValue
			}
		}
	}
	// only count promise with valid QC
	if !in.isMajority(len(verifiedPromise)) {
		//fmt.Println("valid promise size: ", len(verifiedPromise))
		//fmt.Println("promise set be false at here 1")
		return false
	}
	if highestValue == "" {
		return true
	}
	// check if the pre_propose/propose value is the highest value
	if highestValue != prProposeValue {
		//fmt.Println("promise set be false at here 2")
		return false
	}
	return false
}

type timeStamp struct {
	singer    string
	ballot    uint64
	timestamp int64
}

func RestoreTimeStamp(ts *pb.BFTTimeStamp) *timeStamp {
	return &timeStamp{
		singer:    ts.GetSinger(),
		ballot:    ts.GetBallot(),
		timestamp: ts.GetTimestamp(),
	}
}

func (in *Instance) generateTimeStamp(ballot uint64) *pb.BFTTimeStamp {
	return &pb.BFTTimeStamp{
		Singer:    in.name,
		Ballot:    ballot,
		Timestamp: time.Now().Unix(),
	}
}

func (in *Instance) getTimeStamp(ts *timeStamp) *pb.BFTTimeStamp {
	return &pb.BFTTimeStamp{
		Singer:    ts.singer,
		Ballot:    ts.ballot,
		Timestamp: ts.timestamp,
	}
}

func (in *Instance) getMedianTimestamp() int64 {
	var buf []int64
	for _, ts := range in.timeStampSet {
		buf = append(buf, ts.timestamp)
	}
	sort.Sort(int64Slice(buf))

	return buf[len(buf)/2]
}

func GetMedianTimestampFromMsg(ts []*pb.BFTTimeStamp) int64 {
	var buf []int64
	for _, ts := range ts {
		buf = append(buf, ts.Timestamp)
	}
	sort.Sort(int64Slice(buf))

	return buf[len(buf)/2]
}

type PartCertificate struct {
	Sigs *EDDSA.Signature
	Type uint64 // 1: phase2 2: phase3 3: nextBallot
}

func RestorePC(cert *pb.PartialCert) *PartCertificate {
	var pc PartCertificate
	pc.Type = cert.GetType()
	pc.Sigs = EDDSA.RestoreSignature(cert.GetSig())
	return &pc
}

func (in *Instance) getPC(pc PartCertificate) *pb.PartialCert {
	return &pb.PartialCert{
		Sig: &pb.Signature{
			Signer:    pc.Sigs.Signer(),
			Signature: pc.Sigs.Sig(),
		},
		Type: pc.Type,
	}
}

func (in *Instance) generatePartCert(bytes []byte) *pb.PartialCert {
	fmt.Println(in.name, " generates the PC")
	// sign this tuple
	sig := EDDSA.Sign(bytes, in.KeyStore.PrivateKey, in.name)
	//get the partial certificate
	cert := &pb.PartialCert{
		Sig: &pb.Signature{
			Signer:    sig.Signer(),
			Signature: sig.Sig(),
		},
	}
	//fmt.Println("hash in part cert", Hash(bytes))
	return cert
}

func (in *Instance) checkPartCert(message []byte, pc *PartCertificate) bool {
	//fmt.Println("enter check part cert")
	// verify signature
	return EDDSA.VerifySingleMsg(pc.Sigs, message, in.KeyStore.PublicKey[pc.Sigs.Signer()])
}

func (in *Instance) generateBallotPartCert(next uint64) *pb.PartialCert {

	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], next)
	// sign this ballot
	sig := EDDSA.Sign(buf[:], in.KeyStore.PrivateKey, in.name)
	//get the partial certificate
	cert := &pb.PartialCert{
		Sig: &pb.Signature{
			Signer:    sig.Signer(),
			Signature: sig.Sig(),
		},
	}
	return cert
}

func (in *Instance) checkBallotPartCert(next uint64, pc *PartCertificate) bool {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], next)

	// check signature
	return EDDSA.VerifySingleMsg(pc.Sigs, buf[:], in.KeyStore.PublicKey[pc.Sigs.Signer()])
}

// phase 2 and 3 data structure
type acceptMsg struct {
	contention  bool
	phase       uint64
	ballot      uint64
	acceptColor string
	acceptPC    *PartCertificate
	timeStamp   *timeStamp
	//nextBallot  uint64
	//ballotSig   *Signature
}

func (in *Instance) getAccept(resp *pb.AcceptMsg) acceptMsg {
	accept := acceptMsg{
		contention:  resp.GetContention(),
		phase:       resp.GetPhase(),
		ballot:      resp.GetBallot(),
		acceptColor: resp.GetAcceptValue(),
		acceptPC:    RestorePC(resp.GetAcceptPartCert()),
		timeStamp:   RestoreTimeStamp(resp.GetTimeStamp()),
	}

	return accept
}

func (in *Instance) getPreProposeMsg() *pb.PreProposeMsg {
	in.mu.Lock()
	defer in.mu.Unlock()

	preProposeMsg := pb.PreProposeMsg{
		From:            in.name,
		Ballot:          in.promisedBallot,
		PreProposeValue: in.proposedValue,
		PromiseSet:      in.getPromiseSet(),
	}
	return &preProposeMsg
}

func (in *Instance) getProposeMsg() *pb.ProposeMsg {
	in.mu.Lock()
	defer in.mu.Unlock()

	proposeMsg := pb.ProposeMsg{
		From:         in.name,
		Ballot:       in.promisedBallot,
		ProposeValue: in.preAcceptedValue,
		PreAcceptQC:  in.getQC(*in.preAcceptQC),
	}
	return &proposeMsg
}

type QuorumCertificate struct {
	Sigs []*EDDSA.Signature
	Type uint64 // 1: phase2 2: phase3 3: nextBallot
}

func RestoreQC(cert *pb.QuorumCert) *QuorumCertificate {
	var qc QuorumCertificate
	qc.Type = cert.GetType()
	for _, s := range cert.GetSigs() {
		qc.Sigs = append(qc.Sigs, EDDSA.RestoreSignature(s))
	}
	return &qc
}

// CombineQC (only the valid pcs can be placed in the pcs)
func (in *Instance) CombineQC(pcs []*PartCertificate) *QuorumCertificate {
	fmt.Println(in.name, " enter combineQC")
	qc := QuorumCertificate{}
	var signatures []*EDDSA.Signature
	//verify single signature
	for _, pc := range pcs {
		signatures = append(signatures, pc.Sigs)
	}
	qc.Sigs = signatures
	qc.Type = pcs[0].Type
	return &qc

}

func (in *Instance) getQC(qc QuorumCertificate) *pb.QuorumCert {
	var sigs []*pb.Signature
	for _, s := range qc.Sigs {
		sigs = append(sigs, &pb.Signature{
			Signer:    s.Signer(),
			Signature: s.Sig(),
		})
	}
	return &pb.QuorumCert{
		Sigs: sigs,
		Type: qc.Type,
	}
}

func (in *Instance) verifyQC(qc *QuorumCertificate, messages []byte) bool {
	fmt.Println(in.name, " enter verifyQC")
	multiSig, err := EDDSA.Combine(qc.Sigs)
	if err != nil {
		fmt.Println("verifyQC err ")
		return false
	}

	// verify multi signature
	if !EDDSA.VerifyMessage(*multiSig, messages, in.KeyStore.PublicKey) {
		//fmt.Println("verifyQC multi signature")
		return false
	}
	return true
}

func (in *Instance) getCommitMsg() *pb.CommitMsg {
	in.mu.Lock()
	defer in.mu.Unlock()

	commitMsg := pb.CommitMsg{
		From:        in.name,
		Ballot:      in.promisedBallot,
		CommitValue: in.acceptedValue,
		AcceptQC:    in.getQC(*in.acceptQC),
	}
	return &commitMsg
}
