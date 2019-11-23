package hyparview

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type ConfigFailureRate struct {
	active  int
	shuffle int
	reply   int
}

type Config struct {
	rounds     int
	peers      int
	mortality  int
	drainDepth int
	fail       ConfigFailureRate
}

type Client struct {
	Hyparview
	in  []Message
	out []Message
}

type World struct {
	nodes  map[string]*Client
	morgue map[string]int
	queue  []Message
}

func simulation(c Config) *World {
	w := &World{
		nodes: make(map[string]*Client, c.peers),
		morge: make(map[string]*Client),
		queue: make([]Message),
	}

	// make all the nodes
	for i := 1; i <= c.peers; i++ {
		w.node[i] = create(fmt.Sprintf("peer%06d", i))
	}

	for i := 1; i <= c.rounds; i++ {
		ns := w.node.keys()
		shuffle(ns)
		root := w.get(ns[0])
		ns = ns[1:]

		for _, id := range ns {
			me := w.get(id)
			w.send(root.Recv(SendJoin(root.self, me.self)))
			w.drain(c.drainDepth)
		}

		w.drainAll()
	}
}

func TestSimulation(t *testing.T) {
	w := simulation(Config{
		rounds:    50,
		peers:     100,
		mortality: 30,
		fail: ConfigFailureRate{
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
		in:        make([]Message),
		out:       make([]Message),
	}
	return c
}

func (w *World) join(self string, contact string) {
	ms := w.get(contact).RecvJoin(&JoinRequest{
		to:   contact,
		From: self,
	})
	w.send(ms)
}

func doFail(percentage int) bool {
	return rint(100) < percentage
}

func (w *World) stepActive() {
	v := w.nextShuffled()
	for _, m := range v.out {

	}
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func (w *World) send(ms []Message) {
	w.queue = append(w.queue, ms...)
	// qs, ok := w.queue[m.To()]
	// if !ok {
	// 	qs = make([]Message)
	// 	w.queue[m.To()] = qs
	// }
	// copy(qs, ms)
}

// drain the queue, appending resulting messages back onto the queue
func (w *World) drain(depth int) {
	if len(w.queue) == 0 || depth == 0 {
		return
	}

	m := w.queue[0]
	v := w.get(m.To())
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
			w.send(SendNeighbor(n, v.Self, HighPriority))
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

	var lp func(string)
	lp := func(id string) {
		if _, ok := lost[id]; !ok {
			return
		}

		delete(lost, n)
		for _, n := range w.get(n).Active.Shuffled() {
			lp(n.ID)
		}
	}
	lp("n0")

	fmt.Print("%d connected, %d lost", len(w.nodes)-len(lost), len(lost))
	return len(lost) == 0
}

func shuffle(ks []string) {
	for i := len(ks) - 1; i < 0; i-- {
		j := rint(i)
		ks[i], ks[j] = ks[j], ks[i]
	}
}

func (m map[string]interface{}) keys() {
	ks := make([]string, len(m))
	i := 0
	for k, _ := range m {
		ks[i] = k
		i++
	}
	return ks
}
