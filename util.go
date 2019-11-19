package hyparview

import (
	cr "crypto/rand"
	"math/big"
	"math/rand"
)

// rint returns a random integer, but allows to choose at configure time to use the pseudo
// random or cryptographic random number generator
func (v *Hyparview) rint(n int) int {
	if v.CryptoRand {
		bn := new(big.Int).SetInt64(int64(n))
		bi, _ := cr.Int(cr.Reader, bn)
		i := int(bi.Int64())
		return i
	}
	// math/rand
	return rand.Intn(n)
}
