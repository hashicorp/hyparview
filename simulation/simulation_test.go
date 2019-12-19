package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	h "github.com/hashicorp/hyparview"
	"github.com/stretchr/testify/assert"
)

// TestSimulation is the test entry point
func TestSimulation(t *testing.T) {
	for i := 0; i < 3; i++ {
		testSimulation(t, i)
	}
}

// testSimulation is the entry point to test a single world
// World configuration and assertion goes here
func testSimulation(t *testing.T, i int) {
	seed := h.Rint64Crypto(math.MaxInt64 - 1)
	rand.Seed(seed)
	fmt.Printf("Seed %d\n", seed)

	w := simulation(WorldConfig{
		rounds:     5,
		peers:      1000,
		mortality:  30,
		drainDepth: 30,
		payloads:   30,
		gossipHeat: 4,
		iteration:  i,
		fail: WorldFailureRate{
			active:      30,
			shuffle:     30,
			reply:       30,
			gossip:      5,
			gossipReply: 5,
		},
	})

	assert.True(t, w.isConnected())
	assert.Equal(t, 0, len(w.queue))

	// w.debugQueue()
	w.plotSeed(seed)
	w.plotInDegree()
	w.plotGossip()
}

func simulation(c WorldConfig) *World {
	w := &World{
		config: &c,
		nodes:  make(map[string]*Client, c.peers),
		morgue: make(map[string]*Client),
		queue:  make([]h.Message, 0),
	}

	// Make all the nodes
	for i := 0; i < c.peers; i++ {
		id := fmt.Sprintf("n%d", i)
		w.nodes[id] = makeClient(w, id)
	}

	// Connect all the nodes
	for i := 0; i < c.peers; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		root := w.get(ns[0])
		ns = ns[1:]

		for _, id := range ns {
			me := w.get(id)
			w.send(root.Recv(h.SendJoin(root.Self, me.Self)))
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
	for i := 1; i < c.payloads+1; i++ {
		node := ns[i] // client connects to a random node
		_, ms := node.syncGossip(i)
		w.send(ms)
		w.drain(len(ms))
		w.traceGossipRound(i)
	}

	return w
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
