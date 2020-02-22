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
			Addr: fmt.Sprintf("127.0.0.1:1000%d", i),
		}
	}
	return ns
}

func testView(count int) (*Hyparview, []*Node) {
	ns := makeNodes(count)
	hv := CreateView(ns[0], 0)
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
		From:    ns[1],
		Active:  ns[1:3],
		Passive: ns[3:7],
		TTL:     0,
	}

	hv.RecvShuffle(req)

	require.True(t, hv.Passive.Contains(ns[3]))
	require.Equal(t, 0, hv.Active.Size())
	require.Equal(t, 6, hv.Passive.Size())
}

func TestViewMaxAdd(t *testing.T) {
	v := CreateView(NewNode("self"), 0)
	require.Equal(t, 30, v.Passive.Max)
	v.Passive.Max = 3
	v.AddPassive(NewNode("a"))
	v.AddPassive(NewNode("b"))
	v.AddPassive(NewNode("c"))
	v.AddPassive(NewNode("d"))
	v.AddPassive(NewNode("e"))
	require.Equal(t, 3, v.Passive.Size())
}
