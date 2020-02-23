package simulation

import (
	h "github.com/hashicorp/hyparview"
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
	var check := false

	switch m1 := m.(type) {
	// case *h.JoinRequest:
	// 	k := m1.From.ID + "-" + m.To().ID
	// 	w.symmetry[k] = m1
	// case *h.ForwardJoinRequest:
	// 	k := m1.From.ID + "-" + m1.Join.ID
	// 	// We only see these when they changed the active view
	// 	w.symmetry[k] = m1
	case *h.DisconnectRequest:
		check = true
	case *h.NeighborRequest:
		check = true
	default:
		// log unimplemented?
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
			pre := v.Active.Copy()

			w.send(v.Recv(m)...)

			if !v.Active.Equal(pre) {
				w.symCheck(m)
			}
		}
	}
}
