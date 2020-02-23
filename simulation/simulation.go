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
	}

	// log.Printf("debug: make all the nodes")
	for i := 0; i < c.peers; i++ {
		id := fmt.Sprintf("n%d", i)
		w.nodes[id] = makeClient(w, id)
	}

	// log.Printf("debug: connect all the nodes")
	for i := 0; i < c.peers; i++ {
		ns := w.randNodes()
		// boots := w.randNodes()
		boot := ns[0]

		for _, me := range ns[1:] {
			// boot := boots[i]
			me.AddActive(boot.Self)
			ms := boot.Recv(h.SendJoin(boot.Self, me.Self))
			w.send(ms...)
		}

		w.maybeShuffle()
	}

	// log.Printf("debug: send some gossip messages")
	ns := w.randNodes()
	for i := 1; i < c.payloads+1; i++ {
		node := ns[i] // client connects to a random node

		// gossip drains all the hyparview messages and sends all the gossip
		// messages before returning. Also maintains the active view
		node.gossip(i)

		w.traceRound(i)
		w.maybeShuffle()
	}

	return w
}

func (w *World) maybeShuffle() {
	if w.totalMessages%w.config.shuffleFreq != 0 {
		return
	}

	ns := w.randNodes()
	for _, me := range ns {
		p, ms := me.getPeer()
		w.send(ms...)
		if p != nil {
			w.send(me.SendShuffle(p))
		}
	}
}
