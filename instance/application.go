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
		verified, err := vrf.Verify(in.KeyStore.PublicKey[in.name], proof, Int64ToBytes(median))
		if !verified {
			fmt.Println("vrf wrong: ", err)
		}
		//fmt.Println("kproof", proof)
		//fmt.Println("median", median)
		//fmt.Println("all ts: ", in.timeStampSet[0], in.timeStampSet[1], in.timeStampSet[2])
		KProof = proof
		K = HashRatio(vrfOutput)
		//backOffTime := int64(K*math.Pow(3, float64(in.retries))) * in.roundTripTime.Microseconds()
		backOffTime := K * math.Pow(2, float64(in.retries)) * float64(in.roundTripTime.Microseconds())
		backoff := time.Duration(backOffTime) * time.Microsecond
		fmt.Println("backOff time: ", backoff)
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
		prePrepared := in.propose(1)
		if !prePrepared { // if the phase 2 fail due to contention
			fmt.Println("node: " + in.name + " retry phase 2")
			retry = true
			goto retry
		} else {
			//the second accept phase : accept phase
			proposed := in.propose(2)
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
