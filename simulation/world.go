package simulation

import (
	"fmt"
	"sort"

	h "github.com/hashicorp/hyparview"
)

type World struct {
	config        *WorldConfig
	nodes         map[string]*Client
	morgue        map[string]*Client
	bootstrap     h.Node
	totalMessages int
	totalPayloads int

	gossipTotal *gossipRound
	gossipRound []*gossipRound

	spinCount  int
	spinCountM map[string]int
}

type WorldConfig struct {
	gossips     int
	peers       int
	mortality   int
	payloads    int
	iteration   int // count rounds for plot filenames
	failureRate int
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func makeID(i int) string {
	return fmt.Sprintf("n%d", i)
}

func (w *World) nodeKeys() []string {
	// make the slice of keys
	ks := make([]string, len(w.nodes))
	i := 0
	for k := range w.nodes {
		ks[i] = k
		i++
	}

	// start sorted to avoid map order indeterminacy
	sort.Strings(ks)

	// shuffle to have random order, but only math.Rand so we can replay a session
	shuffle(ks)

	return ks
}

func (w *World) randNodes() (ns []*Client) {
	for _, k := range w.nodeKeys() {
		ns = append(ns, w.get(k))
	}
	return ns
}
