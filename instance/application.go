package instance

import (
	"context"
	"fmt"
	"github.com/yoseplee/vrf"
	"math"
	pb "student_22_BFT_Baxos/proto/application"
	"time"
)

func (in *Instance) Propose(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	fmt.Println("Enter propose")
	in.mu.Lock()
	in.pendingValue = req.Color
	in.mu.Unlock()
	retry := false
	K := float64(0)
	var KProof []byte
retry:
	// backoff mechanism at here
	if retry {
		in.retries++
		fmt.Println("node: "+in.name+" retry: ", in.retries)
		// below we compute the backoff time before retrying
		median := in.getMedianTimestamp()
		proof, vrfOutput, _ := vrf.Prove(in.KeyStore.PublicKey[in.name], in.KeyStore.PrivateKey, Int64ToBytes(median))
		KProof = proof
		K = HashRatio(vrfOutput)
		fmt.Println("K: ", K)
		fmt.Println("retries: ", in.retries)
		fmt.Println("RTT: ", in.roundTripTime.Microseconds())
		backOffTime := int64(K*math.Pow(3, float64(in.retries))) * in.roundTripTime.Microseconds()
		fmt.Println("backoff time: ", backOffTime)
		backoff := time.Duration(backOffTime) * time.Microsecond
		// backoff at here
		time.Sleep(backoff)
	}
	// start the color consensus, block until our consensus is completed
	// other consensus instance might be achieved within retry duration
	// we only care our color is painted on the canvas eventually since our single instance version
	prepared := in.prepare(retry, K, KProof)
	if !prepared { // if the phase 1 fail due to contention
		fmt.Println("node: " + in.name + " retry phase 1")
		retry = true
		goto retry
	} else {
		//the first accept phase : pre-accept phase
		prePrepared := in.propose(retry, 1)
		if !prePrepared { // if the phase 2 fail due to contention
			fmt.Println("node: " + in.name + " retry phase 2")
			retry = true
			goto retry
		} else {
			//the second accept phase : accept phase
			proposed := in.propose(retry, 2)
			if !proposed { // if the phase 3 fail due to contention
				fmt.Println("node: " + in.name + " retry phase 3")
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
