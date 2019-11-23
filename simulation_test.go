package hyparview

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	fail       WorldFailureRate
}

type Client struct {
	Hyparview
	in  []Message
	out []Message
}

type World struct {
	nodes map[string]*Client
	// morgue map[string]*Client
	queue []Message
}

func simulation(c WorldConfig) *World {
	w := &World{
		nodes: make(map[string]*Client, c.peers),
		// morgue: make(map[string]*Client),
		queue: make([]Message, 0),
	}

	// make all the nodes
	for i := 1; i <= c.peers; i++ {
		id := fmt.Sprintf("n%d", i)
		w.nodes[id] = create(id)
	}

	for i := 1; i <= c.rounds; i++ {
		ns := w.nodeKeys()
		shuffle(ns)
		root := w.get(ns[0])
		ns = ns[1:]

		for _, id := range ns {
			me := w.get(id)
			w.send(root.Recv(SendJoin(root.Self, me.Self)))
			w.drain(c.drainDepth)
		}

		w.drainAll()
	}

	return w
}

func TestSimulation(t *testing.T) {
	w := simulation(WorldConfig{
		rounds:    50,
		peers:     100,
		mortality: 30,
		fail: WorldFailureRate{
			active:  30,
			shuffle: 30,
			reply:   30,
		},
	})
	require.True(t, w.isConnected())
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

func (w *World) sendOne(m Message) {
	w.queue = append(w.queue, m)
}

func (w *World) send(ms []Message) {
	w.queue = append(w.queue, ms...)
}

// drain the queue, appending resulting messages back onto the queue
func (w *World) drain(depth int) {
	if len(w.queue) == 0 || depth == 0 {
		return
	}

	m := w.queue[0]
	w.queue = w.queue[1:]

	v := w.get(m.To().ID)
	if v == nil {
		return
	}

	ms := v.Recv(m)
	w.send(ms)
	w.drain(depth - 1)
}

func (w *World) drainAll() {
	w.drain(-1)
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
