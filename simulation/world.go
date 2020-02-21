package simulation

import (
	h "github.com/hashicorp/hyparview"
)

type World struct {
	config        *WorldConfig
	nodes         map[string]*Client
	morgue        map[string]*Client
	queue         []h.Message
	totalMessages int
	totalPayloads int

	gossipTotal *gossipRound
	gossipRound []*gossipRound
}

type WorldConfig struct {
	rounds      int
	peers       int
	mortality   int
	payloads    int
	gossipHeat  int
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

// TODO: maybe accept the message we're deciding for and do different things?
func (w *World) shouldFail() bool {
	return h.Rint(100) < w.config.failureRate
}

// Send the messages and all messages caused by them
func (w *World) send(ms ...h.Message) {
	for _, m := range ms {
		if w.shouldFail() {
			continue
		}

		v := w.get(m.To().ID)
		if v != nil {
			w.send(v.Recv(m)...)
		}
	}
}
