package colory

import (
	"context"
	"math"
	"math/rand"
	pb "student_22_BFT_Baxos/proto/application"
	"time"
)

// TODO 3 phases consensus skeleton (completed!)
func (in *Instance) Paint(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	// TODO disable or support concurrent requests from different clients, assume only one call happen now
	in.mu.Lock()
	in.pendingColor = req.Color
	in.mu.Unlock()
	retry := false
	K := float64(0)
	KProof := true
retry:
	// backoff mechanism at here
	if retry {
		in.retries++
		backoff := time.Duration(0) // below we compute the backoff time before retrying
		// TODO generate the K randomly by VRF later
		K = rand.Float64() * 3
		KProof = true
		// Why we use int at here?
		backOffTime := int64(K*math.Pow(2, float64(in.retries))) * in.roundTripTime.Microseconds()
		backoff = time.Duration(backOffTime) * time.Microsecond

		// backoff at here
		time.Sleep(backoff)
	}
	// start the color consensus, block until our consensus is completed
	// other consensus instance might be achieved within retry duration
	// we only care our color is painted on the canvas eventually since our single instance version
	prepared := in.prepare(retry, K, KProof)
	if !prepared { // if the phase 1 fail due to contention
		retry = true
		goto retry
	} else {
		prePrepared := in.propose(retry, 2)
		if !prePrepared { // if the phase 2 fail due to contention
			retry = true
			goto retry
		} else {
			proposed := in.propose(retry, 3)
			if !proposed { // if the phase 3 fail due to contention
				retry = true
				goto retry
			} else {
				response := pb.Response{
					Result: true,
				}
				go in.commit()
				return &response, nil
			}
		}
	}
}
