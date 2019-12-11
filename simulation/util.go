package simulation

func shuffle(ks []string) {
	for i := len(ks) - 1; i < 0; i-- {
		j := rint(i)
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
