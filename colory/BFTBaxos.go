package colory

import (
	"context"
	pb "student_22_BFT_Baxos/proto/BFTBaxos"
	"sync"
	"time"
)

// Promise promises a ballot or refuse a ballot
func (in *Instance) Promise(ctx context.Context, prepare *pb.PrepareMsg) (*pb.PromiseMsg, error) {
	var promise pb.PromiseMsg
	in.mu.Lock()
	defer in.mu.Unlock()

	if in.retryTable[prepare.From] != 0 {
		//TODO update and check local table / IF BACKOFF!
		in.retryTable[prepare.From]++
	}
	in.retryTable[prepare.From]++

	// promise or not
	if prepare.Ballot > in.promisedBallot {
		in.promisedBallot = prepare.Ballot
		promise.Ballot = prepare.Ballot
		promise.Contention = false
		// attach previously accepted color if there has been consensus in the past
		if in.acceptedColor != "" {
			promise.AcceptedBallot = &in.acceptedBallot
			promise.AcceptedValue = &in.acceptedColor
			promise.PreAcceptQC = in.getPreAcceptQC()
		}
	} else {
		promise.Ballot = prepare.Ballot
		promise.Contention = true
	}

	// attach the time stamp
	promise.TimeStamp = in.generateTimeStamp(prepare.Ballot)
	// provide the next ballot number partial signature
	//promise.NextBallot = prepare.Ballot + uint64(len(in.peers)) + 1
	promise.NextBallot = in.promisedBallot + 1
	//TODO partially sign the NextBallot
	//promise.PartialSignature

	return &promise, nil
}

// prepare initiate the BFTBaxos instance and asks the quorum to promise on a ballot number
func (in *Instance) prepare(retry bool, K float64, KProof bool) bool {

	in.mu.Lock()
	// cleanup promise set
	in.promiseSet = nil
	// cleanup backoff mechanism parameter
	in.timeStampSet = nil
	// promise to myself
	in.promisedBallot += in.increment
	in.mu.Unlock()

	promises := make(chan *verifiedPromise)
	prepareMsg := in.getPrepareMsg(retry, K, KProof)

	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	// We cancel as soon as we have a majority to speed things up.
	// We always cancel before leaving the function to prevent a context leak.
	defer cancel()

	//broadcast prepare message
	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)

		// send prepare
		go func(p peer) {
			defer wg.Done()

			resp, err := p.client.Promise(ctx, &prepareMsg)
			if err != nil {
				// We must collect quorum size valid response
				return
			}
			// We must collect quorum size responses which has valid QC
			vp, valid := in.getVerifiedPromise(resp)
			if valid {
				//only collect "valid" promise message
				in.promiseSet = append(in.promiseSet, vp)
				promises <- &vp
			}
		}(p)
	}

	// close promises channel once all responses have been received, failed, or canceled
	go func() {
		wg.Wait()
		cancel()
		close(promises)
	}()

	// count the vote
	yea, nay := 1, 0
	canceled := false

	var highestSeenAcceptedBallot = uint64(0)
	var highestSeenAcceptedColor = ""
	for r := range promises {
		// count the promises
		if !r.promised {
			yea++
		} else {
			nay++
		}

		// find the highest accepted ballot and corresponding color
		if r.acceptedBallot > highestSeenAcceptedBallot {
			highestSeenAcceptedColor = r.acceptedColor
		}
		in.mu.Lock()
		if highestSeenAcceptedColor != "" {
			// if accepted value we pre-propose other value
			in.proposedColor = highestSeenAcceptedColor
		} else {
			// if no accepted value we pre-propose our value
			in.proposedColor = in.pendingColor
			in.pendingColor = ""
		}

		// collect time stamp
		in.timeStampSet = append(in.timeStampSet, r.timeStamp)
		// TODO collect partial signature for next ballot number
		in.mu.Unlock()

		// stop counting as soon as we have a majority
		if !canceled {
			// cancel all in-flight prepare if we have reached a majority
			if in.isMajority(yea) || in.isMajority(nay) {
				cancel()
				canceled = true
			}
		}
	}
	return in.isMajority(yea)
}

func (in *Instance) PreAccept(ctx context.Context, prePropose *pb.PreProposeMsg) (*pb.AcceptMsg, error) {

	//TODO check promise set
	var preAccept pb.AcceptMsg
	in.mu.Lock()
	defer in.mu.Unlock()

	// ballot of prePropose should equal to promised one
	if prePropose.Ballot == in.promisedBallot {
		in.preAcceptedBallot = prePropose.Ballot
		preAccept.Contention = false
		preAccept.Phase = 2
		preAccept.Ballot = prePropose.Ballot
		in.preAcceptedColor = prePropose.PreProposeValue
		preAccept.AcceptValue = in.preAcceptedColor
		//TODO // Partial signature over the tuple < Phase, Ballot, AcceptValue >
		//preAccept.SafetySig = sign(Phase, Ballot, AcceptValue)
	} else {
		preAccept.Contention = true
		preAccept.Phase = 2
		preAccept.Ballot = prePropose.Ballot
	}
	// attach the time stamp
	preAccept.TimeStamp = in.generateTimeStamp(prePropose.Ballot)
	// provide the next ballot number partial signature
	//promise.NextBallot = prepare.Ballot + uint64(len(in.peers)) + 1
	preAccept.NextBallot = in.promisedBallot + 1
	//TODO partially sign the NextBallot
	//promise.PartialSignature
	return &preAccept, nil
}

func (in *Instance) Accept(ctx context.Context, propose *pb.ProposeMsg) (*pb.AcceptMsg, error) {

	var accept pb.AcceptMsg
	in.mu.Lock()
	defer in.mu.Unlock()

	// ballot of propose should equal to promised one
	//TODO check preaccetQC
	if propose.Ballot == in.preAcceptedBallot {
		in.acceptedBallot = propose.Ballot
		accept.Contention = false
		accept.Phase = 3
		accept.Ballot = propose.Ballot
		in.acceptedColor = propose.ProposeValue
		accept.AcceptValue = in.acceptedColor
		//TODO // Partial signature over the tuple < Phase, Ballot, AcceptValue >
		//Accept.SafetySig = sign(Phase, Ballot, AcceptValue)
	} else {
		accept.Contention = true
		accept.Phase = 3
		accept.Ballot = propose.Ballot
	}
	// attach the time stamp
	accept.TimeStamp = in.generateTimeStamp(propose.Ballot)
	// provide the next ballot number partial signature
	//promise.NextBallot = prepare.Ballot + uint64(len(in.peers)) + 1
	accept.NextBallot = in.promisedBallot + 1
	//TODO partially sign the NextBallot
	//promise.PartialSignature
	return &accept, nil
}

func (in *Instance) propose(retry bool, phase uint64) bool {

	in.mu.Lock()
	// cleanup timeStampSet
	in.timeStampSet = nil
	// pre-accept or accept myself
	if phase == 2 {
		in.preAcceptedColor = in.proposedColor
	} else if phase == 3 {
		in.acceptedColor = in.proposedColor
	}
	in.mu.Unlock()

	accepts := make(chan *acceptMsg)
	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	// We cancel as soon as we have a majority to speed things up.
	// We always cancel before leaving the function to prevent a context leak.
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)
		// send prepropose
		go func(p peer) {
			defer wg.Done()
			if phase == 2 {
				resp, err := p.client.PreAccept(ctx, in.getPreProposeMsg())
				if err != nil {
					// We must collect quorum size valid responses to generate QC
					return
				}
				accept := in.getAccept(resp)
				accepts <- &accept
			} else if phase == 3 {
				resp, err := p.client.Accept(ctx, in.getProposeMsg())
				if err != nil {
					// We must collect quorum size valid responses to generate QC
					return
				}
				accept := in.getAccept(resp)
				accepts <- &accept
			}
		}(p)
	}

	// close preAccept channel once all responses have been received, failed, or canceled
	go func() {
		wg.Wait()
		cancel()
		close(accepts)
	}()

	// count the vote
	yea, nay := 1, 0
	var collection []acceptMsg
	canceled := false

	for r := range accepts {
		// count the preAccept
		if !r.contention {
			yea++
			//collect preAcceptMag
			collection = append(collection, *r)
		} else {
			nay++
		}

		in.mu.Lock()
		// collect time stamp
		in.timeStampSet = append(in.timeStampSet, r.timeStamp)
		// TODO collect partial signature for next ballot number
		in.mu.Unlock()

		// Check if QC can be combined
		if in.isMajority(yea) {
			//TODO check if combined successfully
			//TODO do not to forget update our local variable
			success := true
			if success {

				cancel()
				return true
			}
		}

		// stop counting as soon as contention has a majority
		if !canceled {
			// cancel all in-flight prepare if we have reached a majority
			if in.isMajority(nay) {
				cancel()
				canceled = true
			}
		}
	}
	return false
}

func (in *Instance) Commit(ctx context.Context, commitMsg *pb.CommitMsg) {
	in.mu.Lock()
	defer in.mu.Unlock()

	//TODO check acceptQC
	if commitMsg.Ballot == in.promisedBallot {
		in.canvas = append(in.canvas, in.acceptedColor)
	}
	//TODO cleanup local variable
}

func (in *Instance) commit() {
	// commit locally
	in.mu.Lock()
	in.canvas = append(in.canvas, in.acceptedColor)
	in.mu.Unlock()

	ctx := context.Background()
	for _, p := range in.peers {
		// send commit request
		go func(p peer) {
			_, err := p.client.Commit(ctx, in.getCommitMsg())
			if err != nil {
				return
			}
		}(p)
	}
	//TODO cleanup local variable
}

// phase 1 data structure
func (in *Instance) getPrepareMsg(retry bool, K float64, KProof bool) pb.PrepareMsg {
	//TODO complete packaging
	return pb.PrepareMsg{
		From:   in.name,
		Ballot: in.promisedBallot,
	}
}

// verified simple version promise message
type verifiedPromise struct {
	promised       bool
	ballot         uint64
	acceptedBallot uint64
	acceptedColor  string
	acceptQC       acceptQC
	timeStamp      timeStamp
	nextBallot     uint64
	//ballotSig partialSignature
}

func (in *Instance) getVerifiedPromise(resp *pb.PromiseMsg) (verifiedPromise, bool) {
	// TODO change the QC type later
	// TODO check if the AcceptQuorumVert is valid
	//if QC is invalid{
	//	return nil, false
	//}
	vPromise := verifiedPromise{
		promised:       resp.GetContention(),
		ballot:         resp.GetBallot(),
		acceptedBallot: resp.GetAcceptedBallot(),
		acceptedColor:  resp.GetAcceptedValue(),
		acceptQC: acceptQC{
			phase:       resp.GetPreAcceptQC().GetPhase(),
			ballot:      resp.GetPreAcceptQC().GetBallot(),
			acceptColor: resp.GetPreAcceptQC().GetAcceptValue(),
		},
		timeStamp: in.getTimeStamp(resp.GetTimeStamp()),
	}
	return vPromise, true
}

func (in *Instance) getPromiseSet() []*pb.PromiseMsg {
	//the pointer of array
	var promiseSet []*pb.PromiseMsg

	for _, p := range in.promiseSet {
		promiseSet = append(promiseSet,
			&pb.PromiseMsg{
				Contention:     p.promised,
				Ballot:         p.ballot,
				AcceptedBallot: &p.acceptedBallot,
				AcceptedValue:  &p.acceptedColor,
				//TODO QC
				//PreAcceptQC:    &p.acceptQC,
			})
	}

	return promiseSet
}

type timeStamp struct {
	singer    string
	ballot    uint64
	timestamp int64
}

func (in *Instance) getTimeStamp(ts *pb.BFTTimeStamp) timeStamp {
	return timeStamp{
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

type acceptQC struct {
	phase       uint64
	ballot      uint64
	acceptColor string
	//TODO complete encryption component
	//Sig         QuorumSignature
}

// TODO complete this method
func (in *Instance) getPreAcceptQC() *pb.AcceptQuorumCert {
	in.mu.RLock()
	defer in.mu.RUnlock()

	return &pb.AcceptQuorumCert{
		Phase:       2,
		Ballot:      in.preAcceptQC.ballot,
		AcceptValue: in.preAcceptQC.acceptColor,
		//Sig: in.preAcceptQC.Sig
	}
}

// phase 2 and 3 data structure
type acceptMsg struct {
	contention  bool
	phase       uint64
	ballot      uint64
	acceptColor string
	//safetySig partialSignature
	timeStamp  timeStamp
	nextBallot uint64
	//ballotSig partialSignature
}

func (in *Instance) getAccept(resp *pb.AcceptMsg) acceptMsg {
	accept := acceptMsg{
		contention:  resp.GetContention(),
		phase:       resp.GetPhase(),
		ballot:      resp.GetBallot(),
		acceptColor: resp.GetAcceptValue(),
		//TODO PartialSignature SafetySig
		timeStamp: in.getTimeStamp(resp.GetTimeStamp()),
		//TODO BallotSig
	}

	return accept
}

func (in *Instance) getPreProposeMsg() *pb.PreProposeMsg {
	in.mu.Lock()
	defer in.mu.Unlock()

	preProposeMsg := pb.PreProposeMsg{
		From:            in.name,
		Ballot:          in.promisedBallot,
		PreProposeValue: in.proposedColor,
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
		ProposeValue: in.preAcceptedColor,
		//TODO AcceptQuorumCert
	}
	return &proposeMsg
}

func (in *Instance) getCommitMsg() *pb.CommitMsg {
	in.mu.Lock()
	defer in.mu.Unlock()

	commitMsg := pb.CommitMsg{
		From:        in.name,
		Ballot:      in.promisedBallot,
		CommitValue: in.acceptedColor,
		//AcceptQC: in.AcceptQC,
	}
	return &commitMsg
}
