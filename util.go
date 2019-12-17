package hyparview

import (
	cr "crypto/rand"
	"math/big"
	"math/rand"
)

// rand [0, n] inclusive
func rintCrypto(n int) int {
	bn := new(big.Int).SetInt64(int64(n + 1))
	bi, _ := cr.Int(cr.Reader, bn)
	i := int(bi.Int64())
	return i
}

// rint is a placeholder so we can swap out for rintCrypto in testing
// rand [0, n] inclusive
func Rint(n int) int {
	return rand.Intn(n + 1)
}
