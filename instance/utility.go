package instance

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math/big"
)

func ByteTuple(phase uint64, ballot uint64, acceptValue string) []byte {
	var buf []byte

	var phaseBuf [8]byte
	binary.LittleEndian.PutUint64(phaseBuf[:], phase)
	buf = append(buf, phaseBuf[:]...)

	var ballotBuf [8]byte
	binary.LittleEndian.PutUint64(ballotBuf[:], ballot)
	buf = append(buf, ballotBuf[:]...)

	buf = append(buf, []byte(acceptValue)...)

	return buf
}

func Hash(message []byte) []byte {
	hash := sha256.Sum256(message)
	return hash[:]
}

type int64Slice []int64

func (p int64Slice) Len() int {
	return len(p)
}
func (p int64Slice) Less(i, j int) bool {
	return p[i] < p[j]
}
func (p int64Slice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func HashRatio(vrfOutput []byte) float64 {

	t := &big.Int{}
	t.SetBytes(vrfOutput[:])

	precision := uint(8 * (len(vrfOutput) + 1))
	max, b, err := big.ParseFloat("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Fatal("failed to parse big float constant for sortition")
	}

	//hash value as int expression.
	//hval, _ := h.Float64() to get the value
	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	ratio := big.Float{}
	cratio, _ := ratio.Quo(&h, max).Float64()

	return cratio
}
