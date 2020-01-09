package hyparview

import (
	cr "crypto/rand"
	"math/big"
	"math/rand"
)

// rand [0, n] inclusive
func Rint64Crypto(n int64) int64 {
	bn := new(big.Int).SetInt64(int64(n + 1))
	bi, _ := cr.Int(cr.Reader, bn)
	return bi.Int64()
}

// rand [0, n] inclusive
func RintCrypto(n int) int {
	return int(Rint64Crypto(int64(n)))
}

// rint is a placeholder so we can swap out for rintCrypto in testing
// rand [0, n] inclusive
func Rint(n int) int {
	return rand.Intn(n + 1)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
