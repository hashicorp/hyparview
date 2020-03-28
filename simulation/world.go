package simulation

import h "github.com/hashicorp/hyparview"

type World struct {
	config        *WorldConfig
	nodes         map[string]*Client
	morgue        map[string]*Client
	bootstrap     *h.Node
	totalMessages int
	totalPayloads int

	gossipTotal *gossipRound
	gossipRound []*gossipRound

	spinCount  int
	spinCountM map[string]int
}

type WorldConfig struct {
	rounds      int
	peers       int
	mortality   int
	payloads    int
	iteration   int // count rounds for plot filenames
	shuffleFreq int
	failureRate int
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func (w *World) nodeKeys() []string {
	m := w.nodes
	ks := make([]string, len(m))
	i := 0
	for k, _ := range m {
		ks[i] = k
		i++
	}
	return ks
}

func (w *World) randNodes() (ns []*Client) {
	for _, k := range w.nodeKeys() {
		ns = append(ns, w.get(k))
	}
	return ns
}
