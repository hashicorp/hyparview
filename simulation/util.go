package simulation

import h "github.com/hashicorp/hyparview"

func shuffle(ks []string) {
	for i := len(ks) - 1; i < 0; i-- {
		j := h.Rint(i)
		ks[i], ks[j] = ks[j], ks[i]
	}
}

// For the love...
func keys(m map[string]interface{}) []string {
	ks := make([]string, len(m))
	i := 0
	for k, _ := range m {
		ks[i] = k
		i++
	}
	return ks
}

func nodeAddr(nodes []h.Node) (addr []string) {
	for _, n := range nodes {
		addr = append(addr, n.Addr())
	}
	return addr
}
