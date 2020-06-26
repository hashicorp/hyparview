package simulation

import (
	"math/rand"

	h "github.com/hashicorp/hyparview"
)

func shuffle(ks []string) {
	var t string
	rand.Shuffle(len(ks), func(i, j int) {
		t = ks[i]
		ks[i] = ks[j]
		ks[j] = t
	})
}

func nodeAddr(nodes []h.Node) (addr []string) {
	for _, n := range nodes {
		addr = append(addr, n.Addr())
	}
	return addr
}
