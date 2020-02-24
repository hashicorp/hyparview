package simulation

import (
	"log"

	h "github.com/hashicorp/hyparview"
	"github.com/kr/pretty"
)

type World struct {
	config        *WorldConfig
	nodes         map[string]*Client
	morgue        map[string]*Client
	totalMessages int
	totalPayloads int

	gossipTotal *gossipRound
	gossipRound []*gossipRound

	symmetry map[string]h.Message
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

func (w *World) symCheck(m h.Message) {
	first := true

	switch m1 := m.(type) {
	case *h.DisconnectRequest:
		if m1.From.ID == m1.To().ID {
			return
		}
		n := w.get(m1.From.ID)
		m := w.get(m.To().ID)
		if n.Active.Contains(m.Self) {
			log.Printf("diss %s %s", m1.From.ID, m1.To().ID)

			if first {
				first = false
				pretty.Log(n.Active.Nodes, m.Self)
			}
		}

		if m.Active.Contains(n.Self) {
			log.Printf("disr %s %s", m1.From.ID, m1.To().ID)
		}

	case *h.NeighborRequest:
		if !m1.Join || m1.From.ID == m1.To().ID {
			return
		}
		n := w.get(m1.From.ID)
		m := w.get(m.To().ID)
		if !(n.Active.Contains(m.Self) && m.Active.Contains(n.Self)) {
			log.Printf("nei %s %s", m1.From.ID, m1.To().ID)
		}
	default:
	}
}

// Send the messages and all messages caused by them
func (w *World) send(ms ...h.Message) {
	w.totalMessages += len(ms)
	for _, m := range ms {
		// pretty.Log("send", m)
		if w.shouldFail() {
			continue
		}

		v := w.get(m.To().ID)
		if v != nil {
			ms := v.Recv(m)
			// Check after delivery
			w.symCheck(m)
			w.send(ms...)
			// fmt.Printf("%T\n", m)
		}
	}
}
