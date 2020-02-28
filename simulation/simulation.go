package simulation

import (
	"fmt"
	"log"
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
	ns := w.randNodes()
	boot := ns[0]
	for _, me := range ns[1:] {
		// boot := w.nodes[fmt.Sprintf("n%d", h.Rint(i))]
		w.send(me.SendJoin(boot.Self)...)
		w.maybeShuffle()
	}

	log.Printf("debug: send some gossip messages")
	// avoid panic when rounds > peers
	rounds := c.payloads
	if rounds > c.peers {
		rounds = c.peers
	}

	ns = w.randNodes()
	for i := 0; i < c.rounds; i++ {
		// gossip drains all the hyparview messages and sends all the gossip
		// messages before returning. Also maintains the active view
		node := ns[i] // client connects to a random node
		p := i + 1
		w.sendGossip(node.gossip(p)...)
		w.traceRound(p)
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
