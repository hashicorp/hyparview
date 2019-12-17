package simulation

import h "github.com/hashicorp/hyparview"

func shouldFail(percentage int) bool {
	return h.Rint(100) < percentage
}

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
