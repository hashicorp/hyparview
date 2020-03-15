package hyparview

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func makeNodes(n int) []*Node {
	ns := make([]*Node, n)
	for i := 0; i < n; i++ {
		ns[i] = &Node{
			ID:   fmt.Sprintf("127.0.0.1:1000%d", i),
			Addr: fmt.Sprintf("127.0.0.1:1000%d", i),
		}
	}
	return ns
}

type sliceSender struct {
	ms []Message
}

func (s *sliceSender) Send(m Message) (*NeighborRefuse, error) {
	s.ms = append(s.ms, m)
	return nil, nil
}

func (s *sliceSender) Failed(n *Node) {
}

func (s *sliceSender) reset() (ms []Message) {
	out := s.ms
	s.ms = []Message{}
	return out
}

func newSliceSender() *sliceSender {
	s := &sliceSender{}
	s.reset()
	return s
}

func testView(count int) (*Hyparview, []*Node) {
	ns := makeNodes(count)
	hv := CreateView(newSliceSender(), ns[0], 0)
	return hv, ns
}

func TestShuffleSend(t *testing.T) {
	hv, ns := testView(2)
	m := hv.SendShuffle(ns[1])
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
		Active:  ns[1:3],
		Passive: ns[3:7],
		TTL:     0,
	}

	hv.AddActive(ns[1])
	require.True(t, hv.Active.Contains(ns[1]))
	require.True(t, hv.Active.IsEmptyBut(ns[1]))

	hv.RecvShuffle(req)
	require.True(t, hv.Passive.Contains(ns[3]))
	require.Equal(t, 1, hv.Active.Size())
	require.Equal(t, 5, hv.Passive.Size())
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

	v.AddActive(a)
	m := NewForwardJoin(v.Self, b, a, 6)
	v.RecvForwardJoin(m)
	ms := s.reset()
	require.Equal(t, 0, len(ms))

	m = NewForwardJoin(v.Self, a, b, 6)
	v.RecvForwardJoin(m)
	ms = s.reset()
	require.Equal(t, 1, len(ms))
	_, ok := ms[0].(*NeighborRequest)
	require.True(t, ok)

	m = NewForwardJoin(v.Self, a, c, 6)
	v.RecvForwardJoin(m)
	ms = s.reset()
	require.Equal(t, 1, len(ms))
	fwd, ok := ms[0].(*ForwardJoinRequest)
	require.True(t, ok)
	require.Equal(t, 5, fwd.TTL)
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
	v["a"].PromotePassive()
	for _, m := range s["a"].reset() {
		v["b"].Recv(m)
	}

	require.Equal(t, 0, v["a"].Passive.Size())
	require.Equal(t, v["a"].Self, v["b"].Active.Nodes[0])
	require.Equal(t, v["b"].Self, v["a"].Active.Nodes[0])

	for _, n := range []string{"a", "b"} {
		v[n].Active.Add(v["c"].Self)
		v["c"].Active.Add(v[n].Self)
	}

	// Active view overflows, sends a disconnect
	v["a"].RecvJoin(NewJoin(v["a"].Self, NewNode("d")))
	for _, m := range s["a"].reset() {
		t := m.To().ID
		v[t].Recv(m)
	}

	require.True(t, v["a"].Active.Contains(v["b"].Self))
	require.True(t, v["a"].Active.Contains(v["d"].Self))

	require.True(t, v["b"].Active.Contains(v["a"].Self))
	require.True(t, v["b"].Active.Contains(v["c"].Self))

	require.True(t, v["c"].Active.Contains(v["b"].Self))
}
