package instance

import (
	"context"
	"errors"
	"fmt"
	"github.com/yoseplee/vrf"
	"math"
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
		median := GetMedianTimestampFromMsg(prepare.TimeStamp)
		verified, _ := vrf.Verify(in.KeyStore.PublicKey[prepare.From], prepare.KProof, Int64ToBytes(median))
		if !verified {
			return &promise, nil
		}
		// converts proof(pi) into vrf output(hash) without verifying it
		vrfOutput := vrf.Hash(prepare.KProof)
		K := HashRatio(vrfOutput) * 3
		backOffTime := int64(K*math.Pow(2, float64(in.retries))) * in.roundTripTime.Microseconds()
		waiting := time.Now().Unix() - backOffTime
		time.Sleep(time.Duration(waiting))
	}
	in.retryTable[prepare.From]++

	// promise or not
	if prepare.Ballot > in.promisedBallot {
		//promise to this ballot number
		promise.Contention = false
		in.promisedBallot = prepare.Ballot
		promise.Ballot = prepare.Ballot
		// attach previously accepted color if there has been consensus in the past
		if in.acceptedValue != "" {
			promise.Accepted = true
			promise.AcceptedBallot = &in.acceptedBallot
			promise.AcceptedValue = &in.acceptedValue
			promise.PreAcceptQC = in.getQC(*in.preAcceptQC)
			fmt.Println(in.name, " promise ", prepare.From, " with ballot number ", prepare.Ballot, " with previous value")
		} else {
			promise.Accepted = false
			fmt.Println(in.name, " promise ", prepare.From, " with ballot number ", prepare.Ballot, " without previous value")
		}
	} else {
		fmt.Println(in.name, " refuse to promise ", prepare.From, " with ballot number ", prepare.Ballot)
		promise.Contention = true
		promise.Ballot = prepare.Ballot
	}

	// attach the time stamp
	promise.TimeStamp = in.generateTimeStamp(prepare.Ballot)
	// provide the next ballot number partial signature
	//promise.NextBallot = prepare.Ballot + uint64(len(in.peers)) + 1
	cert := in.generateBallotPartCert(prepare.Ballot + 1)
	promise.BallotPartCert = cert
	return &promise, nil
}

// prepare initiate the BFTBaxos instance and asks the quorum to promise on a ballot number
func (in *Instance) prepare(retry bool, K float64, KProof []byte) bool {
	in.mu.Lock()
	// cleanup promise set
	in.promiseSet = nil
	// promise to myself
	//in.prepareBallot += in.increment
	in.prepareBallot += 1
	fmt.Println(in.name, " initiate prepare phase with ballot number", in.prepareBallot)

	//my promise
	in.promisedBallot = in.prepareBallot
	myPromise := &verifiedPromise{
		contention: false,
		ballot:     in.prepareBallot,
		timeStamp:  RestoreTimeStamp(in.generateTimeStamp(in.promisedBallot)),
		ballotPC:   RestorePC(in.generateBallotPartCert(in.prepareBallot + 1)),
	}
	if in.acceptedValue != "" {
		myPromise.acceptedBallot = in.acceptedBallot
		myPromise.acceptedValue = in.acceptedValue
		myPromise.preAcceptQC = in.preAcceptQC
	}
	in.promiseSet = append(in.promiseSet, myPromise)

	in.mu.Unlock()

	promises := make(chan *verifiedPromise)
	prepareMsg := in.getPrepareMsg(retry, K, KProof)

	// cleanup backoff mechanism parameter
	in.timeStampSet = nil

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

			startTime := time.Now()

			resp, err := p.client.Promise(ctx, &prepareMsg)
			if err != nil {
				// We must collect quorum size valid response
				fmt.Println(err)
				return
			}
			fmt.Println("get promise resp from node ")
			in.roundTripTime = time.Now().Sub(startTime) // no mutex needed for roundTripTime variable because it can be stale without affecting correctness

			// We must collect quorum size responses which has valid QC
			vp, valid := in.getVerifiedPromise(resp)
			if valid {
				//only collect "valid" promise message
				in.promiseSet = append(in.promiseSet, vp)
				promises <- vp
			}
		}(p)
	}
	// close promises channel once all responses have been received, failed, or canceled
	go func() {
		wg.Wait()

		//fmt.Println("---------------")
		//fmt.Println("promise 0")
		//fmt.Println(in.promiseSet[0].contention)
		//fmt.Println(in.promiseSet[0].ballot)
		//fmt.Println(in.promiseSet[0].timeStamp)
		//fmt.Println("promise 1")
		//fmt.Println(in.promiseSet[1].contention)
		//fmt.Println(in.promiseSet[1].ballot)
		//fmt.Println(in.promiseSet[1].timeStamp)
		//fmt.Println("---------------")

		cancel()
		close(promises)
	}()

	// count the vote
	yea, nay := 1, 0
	// add our own nextBallotPC
	myNextBallot := in.generateBallotPartCert(in.promisedBallot + 1)
	//fmt.Println("My next ballot: ", myNextBallot)
	nextBallotPCs := []*PartCertificate{RestorePC(myNextBallot)}
	// add our own timeStamp
	myTimeStamp := in.generateTimeStamp(in.promisedBallot)
	//fmt.Println("My timeStamp: ", myTimeStamp)
	in.timeStampSet = append(in.timeStampSet, RestoreTimeStamp(myTimeStamp))
	//fmt.Println("my timestamp set", in.timeStampSet)
	var highestSeenAcceptedBallot = uint64(0)
	var highestSeenAcceptedColor = ""
	canceled := false
	//println("len of promises: ", len(in.promiseSet))
	for r := range promises {
		// count the promises
		if !r.contention {
			yea++
		} else {
			nay++
		}

		// find the highest accepted ballot and corresponding color
		if r.acceptedBallot > highestSeenAcceptedBallot {
			highestSeenAcceptedColor = r.acceptedValue
		}
		in.mu.Lock()
		if highestSeenAcceptedColor != "" {
			// if accepted value we pre-propose other value
			in.proposedValue = highestSeenAcceptedColor
		} else {
			// if no accepted value we pre-propose our value
			in.proposedValue = in.pendingValue
			in.pendingValue = ""
		}

		// collect partial signature for next ballot number
		//if the PC is what we expected
		if in.checkBallotPartCert(in.promisedBallot+1, r.ballotPC) {
			//add into list
			nextBallotPCs = append(nextBallotPCs, r.ballotPC)
			// combine when quorum
			if in.isMajority(len(nextBallotPCs)) {
				fmt.Println("combine next ballot QC")
				in.nextBallotQC = in.CombineQC(nextBallotPCs)
			}
		}

		// collect time stamp
		in.timeStampSet = append(in.timeStampSet, r.timeStamp)
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
	fmt.Println("yea: ", yea)
	fmt.Println("prepare: ", in.isMajority(yea))
	return in.isMajority(yea)
}

func (in *Instance) PreAccept(ctx context.Context, prePropose *pb.PreProposeMsg) (*pb.AcceptMsg, error) {
	var preAccept pb.AcceptMsg

	in.mu.Lock()
	defer in.mu.Unlock()

	if !in.checkPromiseSet(
		prePropose.Ballot,
		prePropose.GetPreProposeValue(),
		prePropose.GetPromiseSet()) {
		fmt.Println("invalid promise set")
		return nil, errors.New("invalid promise set")
	}
	fmt.Println(in.name, " valid promise set")
	// ballot of prePropose should equal to promised one
	if prePropose.Ballot >= in.promisedBallot {
		//pre accept this ballot number
		in.preAcceptedBallot = prePropose.Ballot
		//pre accept this value
		in.preAcceptedValue = prePropose.PreProposeValue
		// contention
		preAccept.Contention = false
		// tuple < Phase, Ballot, AcceptValue >
		preAccept.Phase = 1
		preAccept.Ballot = prePropose.Ballot
		preAccept.AcceptValue = in.preAcceptedValue
		// convert tuple < Phase, Ballot, AcceptValue > to bytes
		bytes := ByteTuple(preAccept.Phase, preAccept.Ballot, preAccept.AcceptValue)
		//fmt.Println("bytes: ", bytes)
		cert := in.generatePartCert(bytes)
		cert.Type = preAccept.Phase
		preAccept.AcceptPartCert = cert
		fmt.Println(in.name, " pre-accept from ", prePropose.From, " with ballot number ", prePropose.Ballot)
	} else {
		preAccept.Contention = true
		preAccept.Phase = 1
		preAccept.Ballot = prePropose.Ballot
		fmt.Println(in.name, " refuse to pre-accept from ", prePropose.From, " with ballot number ", prePropose.Ballot)
	}
	// attach the time stamp
	preAccept.TimeStamp = in.generateTimeStamp(prePropose.Ballot)

	return &preAccept, nil
}

func (in *Instance) Accept(ctx context.Context, propose *pb.ProposeMsg) (*pb.AcceptMsg, error) {
	fmt.Println(in.name, " enter accept phase")
	var accept pb.AcceptMsg

	in.mu.Lock()
	defer in.mu.Unlock()

	// ballot of propose should equal to promised one
	if propose.Ballot >= in.promisedBallot {
		// expected <phase, ballot value> should match with attached QC
		bytes := ByteTuple(uint64(1), propose.Ballot, propose.ProposeValue)
		//fmt.Println("expected ", bytes)
		if !in.verifyQC(RestoreQC(propose.GetPreAcceptQC()), bytes) {
			//fmt.Fprintf(os.Stderr, "preAcceptQC wrong")
			return nil, errors.New("preAcceptQC wrong")
		}
		in.acceptedBallot = propose.Ballot
		accept.Contention = false
		// tuple < Phase, Ballot, AcceptValue >
		accept.Ballot = propose.Ballot
		in.acceptedValue = propose.ProposeValue
		accept.AcceptValue = in.acceptedValue
		bytes = ByteTuple(uint64(2), accept.Ballot, accept.AcceptValue)
		// Partial signature over the tuple < Phase, Ballot, AcceptValue >
		cert := in.generatePartCert(bytes)
		cert.Type = accept.Phase
		accept.AcceptPartCert = cert
		fmt.Println(in.name, " accept from ", propose.From, " with ballot number ", propose.Ballot)
	} else {
		accept.Contention = true
		accept.Phase = 2
		accept.Ballot = propose.Ballot
		fmt.Println(in.name, " refuse to accept from ", propose.From, " with ballot number ", propose.Ballot)
	}
	// attach the time stamp
	accept.TimeStamp = in.generateTimeStamp(propose.Ballot)
	// provide the next ballot number partial signature
	// promise.NextBallot = prepare.Ballot + uint64(len(in.peers)) + 1
	// cert, err := in.generateBallotPartCert(propose.Ballot + 1)
	// if err != nil {
	//	 fmt.Fprintf(os.Stderr, "generate ballot part cert error: %v", err)
	//	 return nil, err
	// }
	// accept.BallotPartCert = cert
	return &accept, nil
}

func (in *Instance) propose(retry bool, phase uint64) bool {
	fmt.Println(in.name, " enter propose phase")
	in.mu.Lock()
	// cleanup timeStampSet
	in.timeStampSet = nil
	// pre-accept or accept myself
	if phase == 1 {
		in.preAcceptedValue = in.proposedValue
		in.preAcceptQC = nil
	} else if phase == 2 {
		in.acceptedValue = in.preAcceptedValue
		in.acceptQC = nil
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

			startTime := time.Now()

			if phase == 1 {
				resp, err := p.client.PreAccept(ctx, in.getPreProposeMsg())
				if err != nil {
					// We must collect quorum size valid responses to generate QC
					return
				}
				in.roundTripTime = time.Now().Sub(startTime) // no mutex needed for roundTripTime variable because it can be stale without affecting correctness
				accept := in.getAccept(resp)
				accepts <- &accept
			} else if phase == 2 {
				resp, err := p.client.Accept(ctx, in.getProposeMsg())
				if err != nil {
					// We must collect quorum size valid responses to generate QC
					return
				}
				in.roundTripTime = time.Now().Sub(startTime) // no mutex needed for roundTripTime variable because it can be stale without affecting correctness
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
	// add our own preAcceptPC
	var acceptPCs []*PartCertificate
	var bytes []byte
	if phase == 1 {
		//fmt.Println("phase 1, ballot: ", in.promisedBallot, " value: ", in.preAcceptedColor)
		bytes = ByteTuple(uint64(1), in.promisedBallot, in.preAcceptedValue)
		//fmt.Println("bytes: ", bytes)
	} else if phase == 2 {
		//fmt.Println("phase 2, ballot: ", in.promisedBallot, " value: ", in.acceptedColor)
		bytes = ByteTuple(uint64(2), in.promisedBallot, in.acceptedValue)
		//fmt.Println(bytes)
	}
	myAcceptQC := RestorePC(in.generatePartCert(bytes))
	acceptPCs = append(acceptPCs, myAcceptQC)
	// add our own timeStamp
	myTimeStamp := in.generateTimeStamp(in.promisedBallot)
	in.timeStampSet = append(in.timeStampSet, RestoreTimeStamp(myTimeStamp))

	quorumCertificate := false
	canceled := false

	for r := range accepts {
		// count the preAccept
		if !r.contention {
			yea++
		} else {
			nay++
		}

		in.mu.Lock()
		// collect time stamp
		in.timeStampSet = append(in.timeStampSet, r.timeStamp)

		// collect partial signature for preAcceptQC or acceptQC
		// if the PC is what we expected
		// < Phase, Ballot, AcceptValue >
		var message []byte
		if phase == 1 {
			message = ByteTuple(uint64(1), in.promisedBallot, in.proposedValue)
			//fmt.Println(uint64(1))
			//fmt.Println(in.promisedBallot)
			//fmt.Println(in.proposedColor)
			//fmt.Println("bytes: ", message)
		} else if phase == 2 {
			message = ByteTuple(uint64(2), in.promisedBallot, in.preAcceptedValue)
			//fmt.Println(uint64(2), "  ", in.promisedBallot, "  ", in.proposedColor)
		}
		if in.checkPartCert(message, r.acceptPC) {
			//this is what we look forward, let's add into list
			acceptPCs = append(acceptPCs, r.acceptPC)
		}
		//fmt.Println("phase ", phase, " the len of qc ", len(acceptPCs))
		// combine when quorum safely!
		if in.isMajority(len(acceptPCs)) && phase == 1 && in.preAcceptQC == nil {
			fmt.Println("combine preAccept QC")
			in.preAcceptQC = in.CombineQC(acceptPCs)
		} else if in.isMajority(len(acceptPCs)) && phase == 2 && in.acceptQC == nil {
			//fmt.Println("phase ", phase, " try to combine qc ")
			fmt.Println("combine accept QC")
			in.acceptQC = in.CombineQC(acceptPCs)
			//fmt.Println(in.acceptQC.Hash)
		}
		in.mu.Unlock()
		// stop counting as soon as contention has a majority
		if !canceled {
			// cancel all in-flight propose if we have reached a majority
			if in.isMajority(yea) && phase == 1 && in.preAcceptQC != nil {
				cancel()
				canceled = true
				quorumCertificate = true
			} else if in.isMajority(yea) && phase == 2 && in.acceptQC != nil {
				cancel()
				canceled = true
				quorumCertificate = true
			} else if in.isMajority(nay) {
				cancel()
				canceled = true
				quorumCertificate = true
			}
		}
	}
	fmt.Println("yea: ", yea)
	fmt.Println("quorum", quorumCertificate)
	return in.isMajority(yea) && quorumCertificate
}

func (in *Instance) Commit(ctx context.Context, commitMsg *pb.CommitMsg) (*pb.Empty, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if commitMsg.Ballot == in.promisedBallot {
		bytes := ByteTuple(uint64(2), commitMsg.Ballot, commitMsg.CommitValue)

		if !in.verifyQC(RestoreQC(commitMsg.GetAcceptQC()), bytes) {
			//fmt.Fprintf(os.Stderr, "AcceptQC wrong")
			return nil, errors.New("AcceptQC wrong")
		}
		in.committed = append(in.committed, in.acceptedValue)
		fmt.Println(in.name, " commit ", in.acceptedValue)
	}

	//clean up
	in.cleanup()
	// reset the retry table
	in.retryTable[commitMsg.From] = int64(0)

	//response
	empty := pb.Empty{}
	return &empty, nil
}

func (in *Instance) commit() {
	// commit locally
	in.mu.Lock()
	fmt.Println(in.name, " commit ", in.acceptedValue)
	in.committed = append(in.committed, in.acceptedValue)
	in.mu.Unlock()
	//snapshot the current status
	commitMsg := in.getCommitMsg()
	ctx := context.Background()
	for _, p := range in.peers {
		// send commit request
		go func(p peer) {
			_, err := p.client.Commit(ctx, commitMsg)
			if err != nil {
				return
			}
		}(p)
	}

	in.mu.Lock()
	//clean up
	in.cleanup()
	// reset the retry table
	in.retryTable[in.name] = int64(0)
	in.retries = 0
	in.mu.Unlock()
}

// cleanup should be used within a lock block
func (in *Instance) cleanup() {
	//reset the liveness related variable
	in.retries = 0
	// reset the consensus related variable
	in.preAcceptedBallot = uint64(0)
	in.preAcceptedValue = ""
	in.acceptedBallot = uint64(0)
	in.acceptedValue = ""
	in.preAcceptQC = nil
}
