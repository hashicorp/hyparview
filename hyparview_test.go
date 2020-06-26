package hyparview

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeNodes(n int) []Node {
	ns := make([]Node, n)
	for i := 0; i < n; i++ {
		ns[i] = &node{
			addr: fmt.Sprintf("127.0.0.1:1000%d", i),
		}
	}
	return ns
}

type wrapper struct {
	m Message
	k SendCallback
}

func wrap(m Message, k SendCallback) wrapper {
	return wrapper{m: m, k: k}
}

type sliceSender struct {
	ms []wrapper
}

func (s *sliceSender) Send(m Message, k SendCallback) {
	s.ms = append(s.ms, wrap(m, k))
}

func (s *sliceSender) Failed(n Node) {}
func (s *sliceSender) Bootstrap()    {}

func (s *sliceSender) reset() (ms []Message) {
	for _, w := range s.ms {
		ms = append(ms, w.m)
	}
	s.ms = []wrapper{}
	return ms
}

func (s *sliceSender) deliver(ns map[string]*Hyparview) {
	for _, w := range s.ms {
		m := w.m
		t := m.To()
		a := t.Addr()
		r := ns[a].Recv(w.m)
		w.k(r, nil)
	}
	s.reset()
}

func newSliceSender() *sliceSender {
	s := &sliceSender{}
	s.reset()
	return s
}

func testView(count int) (*Hyparview, []Node) {
	ns := makeNodes(count)
	hv := CreateView(newSliceSender(), ns[0], 0)
	return hv, ns
}

func TestShuffleSend(t *testing.T) {
	hv, ns := testView(2)
	m := hv.composeShuffle(ns[1])
	require.NotNil(t, m.Active)
	require.NotNil(t, m.Passive)
	require.Equal(t, 0, len(m.Active))
	require.Equal(t, 0, len(m.Passive))
}

func TestShuffleRecv(t *testing.T) {
	hv, ns := testView(10)

	req := &ShuffleRequest{
		to:      ns[0],
		from:    ns[1],
		Origin:  ns[1],
		Active:  ns[1:3],
		Passive: ns[3:7],
		TTL:     0,
	}

	hv.AddActive(ns[1])
	require.True(t, hv.Active.Contains(ns[1]))
	require.True(t, hv.Active.IsEmptyBut(ns[1]))

	hv.RecvShuffle(req)
	require.True(t, hv.Passive.Contains(ns[3]))
	// still only 1, which was added explicitly
	require.Equal(t, 1, hv.Active.Size())
	require.False(t, hv.Passive.Contains(ns[1])) // excluded, it's in active view
	require.True(t, hv.Passive.Contains(ns[2]))
	require.Equal(t, 1+4, hv.Passive.Size()) // 4 more from Passive
}

func TestViewMaxAdd(t *testing.T) {
	v := CreateView(newSliceSender(), NewNode("self"), 0)
	require.Equal(t, 30, v.Passive.Max)
	v.Passive.Max = 3
	v.AddPassive(NewNode("a"))
	v.AddPassive(NewNode("b"))
	v.AddPassive(NewNode("c"))
	v.AddPassive(NewNode("d"))
	v.AddPassive(NewNode("e"))
	require.Equal(t, 3, v.Passive.Size())
}

func TestDisconnect(t *testing.T) {
	v := CreateView(newSliceSender(), NewNode("self"), 0)
	n := NewNode("a")
	v.AddActive(n)
	v.RecvDisconnect(NewDisconnect(v.Self, n))
	require.False(t, v.Active.Contains(n))
	require.True(t, v.Passive.Contains(n))
}

func TestRecvForwardJoin(t *testing.T) {
	s := newSliceSender()
	v := CreateView(s, NewNode("self"), 0)
	a := NewNode("a")
	b := NewNode("b")
	c := NewNode("c")

	// with repairAsymmetry, RecvForwardJoin adds the sender b to the active view. We
	// don't forward the message because it's from `a` and Joining `b`, so there's no
	// peer to forward to
	v.AddActive(a)
	m := NewForwardJoin(v.Self, b, a, 6)
	v.RecvForwardJoin(m)
	ms := s.reset()
	require.Equal(t, 0, len(ms))

	// with repairAsymmetry, RecvForwardJoin adds the sender a to the active view and so fo
	m = NewForwardJoin(v.Self, a, c, 6)
	v.RecvForwardJoin(m)
	ms = s.reset()
	require.Equal(t, 1, len(ms))
	fwd, ok := ms[0].(*ForwardJoinRequest)
	require.True(t, ok)
	require.Equal(t, 5, fwd.TTL)

	// the ttl is exhausted, so add the peer and send them a neighbor request
	m = NewForwardJoin(v.Self, a, c, 0)
	v.RecvForwardJoin(m)
	ms = s.reset()
	require.Equal(t, 1, len(ms))
	_, ok = ms[0].(*NeighborRequest)
	require.True(t, ok)
}

func TestNeighborSymmetry(t *testing.T) {
	s := map[string]*sliceSender{}
	v := map[string]*Hyparview{}

	for _, n := range []string{"a", "b", "c", "d"} {
		s[n] = newSliceSender()
		v[n] = CreateView(s[n], NewNode(n), 0)
		v[n].Active.Max = 2
		v[n].Passive.Max = 2
	}

	// Promote passive
	v["a"].Passive.Add(v["b"].Self)
	v["a"].PromotePassive(nil)
	s["a"].deliver(v)

	require.Equal(t, 0, v["a"].Passive.Size())
	require.Equal(t, v["a"].Self, v["b"].Active.Nodes[0])
	require.Equal(t, v["b"].Self, v["a"].Active.Nodes[0])

	for _, n := range []string{"a", "b"} {
		v[n].Active.Add(v["c"].Self)
		v["c"].Active.Add(v[n].Self)
	}

	// Active view overflows, sends a disconnect. The disconnected peer is selected at
	// random, so we need to seed a deterministic outcome
	rand.Seed(0)
	v["a"].RecvJoin(NewJoin(v["a"].Self, NewNode("d")))
	for _, m := range s["a"].reset() {
		t := m.To().Addr()
		v[t].Recv(m)
	}

	require.True(t, v["a"].Active.Contains(v["c"].Self))
	require.True(t, v["a"].Active.Contains(v["d"].Self))

	require.True(t, v["b"].Active.Contains(v["c"].Self))

	require.True(t, v["c"].Active.Contains(v["a"].Self))
	require.True(t, v["c"].Active.Contains(v["b"].Self))
}
