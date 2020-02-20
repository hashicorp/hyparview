package simulation

import (
	"fmt"

	h "github.com/hashicorp/hyparview"
)

func simulation(c WorldConfig) *World {
	w := &World{
		config: &c,
		nodes:  make(map[string]*Client, c.peers),
		morgue: make(map[string]*Client),
		queue:  make([]h.Message, 0),
	}

	// log.Printf("debug: make all the nodes")
	for i := 0; i < c.peers; i++ {
		id := fmt.Sprintf("n%d", i)
		w.nodes[id] = makeClient(w, id)
	}

	// log.Printf("debug: connect all the nodes")
	for i := 0; i < c.peers; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		root := w.get(ns[0])
		ns = ns[1:]

		for _, id := range ns {
			me := w.get(id)
			ms := root.Recv(h.SendJoin(root.Self, me.Self))
			w.drain(ms)
		}
	}

	// log.Printf("debug: shuffle a few times")
	for i := 0; i < c.rounds; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		for i := 1; i < len(ns); i++ {
			me := w.get(ns[i-1])
			thee := w.get(ns[i])
			ms := me.SendShuffle(thee.Self)
			w.drain(ms)
		}
	}

	// log.Printf("debug: send some gossip messages")
	ns := w.randNodes()
	for i := 1; i < c.payloads+1; i++ {
		node := ns[i] // client connects to a random node

		// gossip drains all the hyparview messages and sends all the gossip
		// messages before returning
		node.gossip(i)

		w.traceRound(i)
	}

	return w
}

func (w *World) repairEmptyActive() (ms []h.Message) {
	for _, n := range w.nodes {
		if n.Active.IsEmpty() {
			return n.failActive(nil)
		}
	}
	return ms
}

func (w *World) debugQueue() {
	fwd, ttl, disc, misc, shuf, shufr := 0, 0, 0, 0, 0, 0
	for _, m := range w.queue {
		switch r := m.(type) {
		case *h.ForwardJoinRequest:
			fwd++
		case *h.DisconnectRequest:
			disc++
		case *h.ShuffleRequest:
			shuf++
			ttl += r.TTL
		case *h.ShuffleReply:
			shufr++
		default:
			misc++
		}
	}

	avg := 0
	if shuf > 0 {
		avg = ttl / shuf
	}

	fmt.Printf("FWD %d DISC %d MISC %d SHUF %d (TTL %d) SHUFR %d\n",
		fwd, disc, misc, shuf, avg, shufr)
}
