package colory

import (
	"context"
	"fmt"
	pb "student_22_BFT_Baxos/proto/consensus"
	"sync"
)

// Promise promises the ballot number in the prepare request
// piggyback the previously accepted color
func (in *Instance) Promise(ctx context.Context, req *pb.PrepareRequest) (*pb.PromiseResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	var promise pb.PromiseResponse
	attachment := ""

	//attach previously accepted color
	if in.ballot > 0 {
		promise.Ballot = in.ballot
		promise.Color = in.color
		attachment = fmt.Sprintf(" (attached previously accepted ballot number %v and color `%v`)", in.ballot, in.color)
	}

	//ballot number of request greater than the ballot number promised
	if req.Ballot > in.promised {
		promise.Promised = true
		//only update the "promised", not the "ballot number"
		in.promised = req.Ballot
		fmt.Printf("promised ballot numebr %v%v\n", req.Ballot, attachment)
	} else {
		fmt.Printf("did not promise ballot number %v%v\n", req.Ballot, attachment)
	}

	return &promise, nil
}

// Accept accepts the color for a previously promised ballot number
func (in *Instance) Accept(ctx context.Context, req *pb.ProposeRequest) (*pb.AcceptResponse, error) {
	in.mu.Lock()
	defer in.mu.Unlock()

	if req.Ballot >= in.promised {
		//update the "ballot number" and "color"
		in.ballot = req.Ballot
		in.color = req.Color
		fmt.Printf("accepted ballot number %v and color `%v`\n", in.ballot, in.color)
	} else {
		fmt.Printf("did not accept ballot number %v and color `%v`\n", req.Ballot, req.Color)
	}

	return &pb.AcceptResponse{
		Accepted: req.Ballot == in.ballot,
	}, nil
}

// prepare asks the quorum to promise a ballot number. It learns previous consensus if there is any
func (in *Instance) prepare() bool {
	type response struct {
		from string
		// consensus.PromiseResponse
		promised bool
		ballot   uint64
		color    string
	}

	// update "promised" according to "increment"
	in.promised += in.increment

	responses := make(chan *response)
	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)

		// broadcast prepare
		go func(p peer) {
			defer wg.Done()

			// use our "promised" as the ballot number in the "PrepareRequest"
			resp, err := p.client.Promise(ctx, &pb.PrepareRequest{
				Ballot: in.promised,
			})
			fmt.Printf("prepare Ballot number %v to %v: sent\n", in.promised, p.name)
			if err != nil {
				if ctx.Err() == context.Canceled {
					fmt.Printf("prepare Ballot number %v to %v: canceled\n", in.promised, p.name)
					return
				}
				// We want errors which are not the result of a canceled
				// proposal to be counted as a negative answer (nay) later.
				// For that we emit an empty response into the channel in those
				// cases.
				responses <- &response{from: p.name}
				fmt.Printf("propose Ballot number %v to %v: %v\n", in.promised, p.name, err)
			}
			responses <- &response{
				from:     p.name,
				promised: resp.Promised,
				ballot:   resp.Ballot,
				color:    resp.Color,
			}
		}(p)
	}

	// close responses channel once all responses have been received, failed, or canceled
	go func() {
		wg.Wait()
		close(responses)
	}()

	// count the votes
	// I am always my own boss! So vote to myself!
	yea, nay := 1, 0
	canceled := false
	for r := range responses {
		// count the promises
		if r.promised {
			yea++
			fmt.Printf("propose Ballot number %v to %v: got yea\n", in.promised, r.from)
		} else {
			nay++
			fmt.Printf("propose Ballot number %v to %v: got nay\n", in.promised, r.from)
		}

		// select previously accepted Ballot number and color from other instances
		if r.ballot > in.ballot {
			// update my ballot as same as the highest ballot number within the quorum responses
			in.ballot = r.ballot
			// update my color as same as the color accepted with the highest ballot number
			in.color = r.color
			fmt.Printf("propose Ballot %v to %v: selected Ballot %v and color `%v`\n", in.promised, r.from, r.ballot, r.color)
		}

		// stop counting as soon as we have a majority
		if !canceled {
			// cancel all in-flight proposals if we have reached a majority
			if in.isMajority(yea) || in.isMajority(nay) {
				cancel()
				canceled = true
			}
		}
	}

	// if we learned a higher ID than our initial proposal suggested, then we also promise this higher ID
	if in.ballot > in.promised {
		in.promised = in.ballot
		fmt.Printf("jumped to promise Ballot number %v\n", in.promised)
	}

	return in.isMajority(yea)
}

// propose asks the quorum to accept a color
func (in *Instance) propose(ballot uint64, color string) bool {
	type response struct {
		from     string
		accepted bool
	}

	fmt.Printf("proposing ID %v and color `%v`\n", ballot, color)

	responses := make(chan *response)
	ctx, cancel := context.WithTimeout(context.Background(), in.timeout)
	defer cancel()

	wg := sync.WaitGroup{}
	for _, p := range in.peers {
		wg.Add(1)

		// broadcast propose requests
		go func(p peer) {
			defer wg.Done()

			resp, err := p.client.Accept(ctx, &pb.ProposeRequest{
				Ballot: ballot,
				Color:  color,
			})
			fmt.Printf("propose Ballot %v and color `%v` to %v: sent\n", ballot, color, p.name)

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					fmt.Printf("propose Ballot %v and color `%v` to %v: deadline exceeded\n", ballot, color, p.name)
					return
				}
				// We want errors which are not the result of a canceled propose to be counted as a negative answer (nay)
				// later. For that we emit an empty response into the channel in those cases.
				responses <- &response{from: p.name}
				fmt.Printf("propose Ballot %v and color `%v` to %v: %v\n", ballot, color, p.name, err)
				return
			}
			responses <- &response{
				from:     p.name,
				accepted: resp.Accepted,
			}
		}(p)
	}

	// close responses channel once all responses have been received, failed, or canceled
	go func() {
		wg.Wait()
		close(responses)
	}()

	// we have to accept our own color
	in.ballot = ballot
	in.color = color

	// count the vote
	yea := 1 // we just committed our own data. make it count.
	for r := range responses {
		if r.accepted {
			yea++
			fmt.Printf("propose Ballot %v and color `%v` to %v: got yea\n", ballot, color, r.from)
			continue
		}
		fmt.Printf("propose Ballot %v and color `%v` to %v: got nay\n", ballot, color, r.from)
	}

	return in.isMajority(yea)
}

// isMajority returns true if the n represents a majority in the configured
// quorum. Caller must hold a (read) lock on i (Instance).
func (in *Instance) isMajority(n int) bool {
	return n > ((len(in.peers) + 1) / 2)
}
