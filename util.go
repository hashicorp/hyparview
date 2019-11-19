package hyparview

import (
	cr "crypto/rand"
	"math/big"
	"math/rand"
)

func rintCrypto(n int) int {
	bn := new(big.Int).SetInt64(int64(n))
	bi, _ := cr.Int(cr.Reader, bn)
	i := int(bi.Int64())
	return i
}

// rint returns a random integer, but allows to choose at configure time to use the pseudo
// random or cryptographic random number generator
func rint(n int) int {
	return rand.Intn(n)
}
