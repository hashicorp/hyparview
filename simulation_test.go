package hyparview

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

type WorldFailureRate struct {
	active  int
	shuffle int
	reply   int
}

type WorldConfig struct {
	rounds     int
	peers      int
	mortality  int
	drainDepth int
	payloads   int
	fail       WorldFailureRate
}

type Client struct {
	Hyparview
	app      int
	appSeen  int
	appWaste int
	in       []Message
	out      []Message
}

type World struct {
	nodes map[string]*Client
	// morgue map[string]*Client
	queue         []Message
	totalMessages int
	totalPayloads int
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
	for i := 0; i < c.payloads; i++ {
		for _, node := range w.randNodes() {
			peer := node.Peer()
			if peer == nil {
				w.failActive(node)
				w.drain(100)
			}
			w.gossip(node, peer, i)
		}
	}

	return w
}

func create(id string) *Client {
	v := CreateView(&Node{ID: id, Addr: ""}, 0)
	c := &Client{
		Hyparview: *v,
		in:        make([]Message, 0),
		out:       make([]Message, 0),
	}
	return c
}

func doFail(percentage int) bool {
	return rint(100) < percentage
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func (w *World) send(ms []Message) {
	w.totalMessages += len(ms)
	w.queue = append(w.queue, ms...)
}

func (w *World) sendOne(m Message) {
	w.send([]Message{m})
}

// drain the queue, appending resulting messages back onto the queue
func (w *World) drain(depth int) {
	for depth != 0 {
		if len(w.queue) == 0 {
			return
		}

		m := w.queue[0]
		w.queue = w.queue[1:]

		v := w.get(m.To().ID)
		if v != nil {
			ms := v.Recv(m)
			w.send(ms)
		}

		depth--
	}
}

func (w *World) drainAll() {
	// This should be -1, but it keeps blowing the stack. Suggests I've got a bug where
	// I never stop sending messages

	// w.drain(-1)
	w.drain(100)
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

func (w *World) isConnected() bool {
	lost := make(map[string]*Client, len(w.nodes))
	for k, v := range w.nodes {
		lost[k] = v
	}

	var lp func(*Node)
	lp = func(n *Node) {
		if _, ok := lost[n.ID]; !ok {
			return
		}

		delete(lost, n.ID)
		for _, m := range w.get(n.ID).Active.Shuffled() {
			lp(m)
		}
	}

	// I hate that this is lp(first(nodes))
	var start *Node
	for _, v := range w.nodes {
		start = v.Self
		break
	}
	lp(start)

	fmt.Printf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
	return len(lost) == 0
}

func (w *World) PlotInDegree() {
	plot := func(ns func(*Hyparview) []*Node, path string) {
		act := map[string]int{}
		for _, v := range w.nodes {
			for _, n := range ns(&v.Hyparview) {
				act[n.ID] += 1
			}
		}

		max := 0
		for _, c := range act {
			if c > max {
				max = c
			}
		}

		deg := make([]int, max+1)
		for _, c := range act {
			deg[c] += 1
		}

		f, _ := os.Create(path)
		defer f.Close()
		for i, c := range deg {
			f.WriteString(fmt.Sprintf("%d %d\n", i, c))
		}
	}

	plot(func(v *Hyparview) []*Node { return v.Active.Nodes }, "active.data")
	plot(func(v *Hyparview) []*Node { return v.Passive.Nodes }, "passive.data")
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

func shuffle(ks []string) {
	for i := len(ks) - 1; i < 0; i-- {
		j := rint(i)
		ks[i], ks[j] = ks[j], ks[i]
	}
}

// For the love...
func keys(m map[string]interface{}) []string {
	ks := make([]string, len(m))
	i := 0
	for k, _ := range m {
		ks[i] = k
		i++
	}
	return ks
}
