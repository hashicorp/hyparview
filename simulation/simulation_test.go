package simulation

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSimulation is the only test entry point. Configure and assert everything here
func TestSimulation(t *testing.T) {
	w := simulation(WorldConfig{
		rounds:     5,
		peers:      1000,
		mortality:  30,
		drainDepth: 30,
		fail: WorldFailureRate{
			active:  30,
			shuffle: 30,
			reply:   30,
		},
	})
	assert.True(t, w.isConnected())
	assert.Equal(t, 0, len(w.queue))

	fwd, ttl, disc, misc, shuf, shufr := 0, 0, 0, 0, 0, 0
	for _, m := range w.queue {
		switch r := m.(type) {
		case *ForwardJoinRequest:
			fwd++
		case *DisconnectRequest:
			disc++
		case *ShuffleRequest:
			shuf++
			ttl += r.TTL
		case *ShuffleReply:
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

	w.PlotInDegree()
}

func simulation(c WorldConfig) *World {
	w := &World{
		nodes: make(map[string]*Client, c.peers),
		// morgue: make(map[string]*Client),
		queue: make([]Message, 0),
	}

	// Make all the nodes
	for i := 0; i < c.peers; i++ {
		id := fmt.Sprintf("n%d", i)
		w.nodes[id] = create(id)
	}

	// Connect all the nodes
	for i := 0; i < c.peers; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		root := w.get(ns[0])
		ns = ns[1:]

		for _, id := range ns {
			me := w.get(id)
			w.send(root.Recv(SendJoin(root.Self, me.Self)))
			w.drain(c.drainDepth)
		}
	}

	// Shuffle a few times
	for i := 0; i < c.rounds; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		for i := 1; i < len(ns); i++ {
			me := w.get(ns[i-1])
			thee := w.get(ns[i])
			w.send(me.SendShuffle(thee.Self))
			w.drain(9)
		}
	}

	// Send some messages
	ns := w.randNodes()
	for i := 0; i < c.payloads; i++ {
		node := ns[i] // client connects to a random node
		peer := node.Peer()
		if peer == nil {
			w.failActive(node)
			w.drain(100)
		}
		node.gossip(peer, i)
	}

	return w
}

func doFail(percentage int) bool {
	return rint(100) < percentage
}

func (w *World) failActive(n *Node) {
	v := w.get(n.ID)
	for _, n := range v.Passive.Shuffled() {
		if v.Active.IsEmpty() {
			// simulate sync network call
			// TODO simulate failure
			w.sendOne(SendNeighbor(n, v.Self, HighPriority))
			break
		} else {
			m := SendNeighbor(n, v.Self, LowPriority)
			ms := w.get(n.ID).RecvNeighbor(m)
			// any low priority response is failure
			if len(ms) == 0 {
				v.DelPassive(n)
				w.send(v.AddActive(n))
				break
			}
			v.DelPassive(n)
		}
	}
}
