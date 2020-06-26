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
	network       []message

	gossipTotal *gossipRound
	gossipRound []*gossipRound

	spinCount  int
	spinCountM map[string]int
}

func newWorld(c WorldConfig) *World {
	return &World{
		config:  &c,
		nodes:   make(map[string]*Client, c.peers),
		morgue:  make(map[string]*Client),
		network: []message{},
	}
}

type WorldConfig struct {
	gossips     int
	peers       int
	mortality   int
	payloads    int
	iteration   int // count rounds for plot filenames
	failureRate int
}

type message struct {
	m h.Message
	k h.SendCallback
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

func (w *World) sendMesg(m h.Message, k h.SendCallback) {
	w.network = append(w.network, message{m: m, k: k})
}

func (w *World) deliver() {
	ms := w.network
	w.network = []message{}
	for _, m := range ms {
		n := w.get(m.m.To().Addr())
		o := n.recv(m.m)

		from := w.get(m.m.From().Addr())
		from.callback(o, m.k)
	}
}

func (w *World) deliverAll() {
	for tries := 100; tries > 0; tries-- {
		w.deliver()
		if len(w.network) == 0 {
			break
		}
	}
}
