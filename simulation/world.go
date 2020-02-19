package simulation

import (
	h "github.com/hashicorp/hyparview"
	"github.com/kr/pretty"
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

type WorldFailureRate struct {
	active      int
	shuffle     int
	reply       int
	gossip      int
	gossipReply int
}

type WorldConfig struct {
	rounds     int
	peers      int
	mortality  int
	payloads   int
	gossipHeat int
	iteration  int // count rounds for plot filenames
	fail       WorldFailureRate
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

// drain the queue of outgoing messages, delivering them
func (w *World) drain(c *Client) {
	ms := c.messages()
	pretty.Log("drain", ms)

	for _, m := range ms {
		v := w.get(m.To().ID)
		if v != nil {
			v.Recv(m)
			w.drain(v)
		}
	}

	ms = ns
}
