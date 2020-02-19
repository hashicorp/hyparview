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

func testView(count int) (*Hyparview, *SliceSender, []*Node) {
	ns := makeNodes(count)
	hv := CreateView(ns[0], 0)
	s := NewSliceSender()
	hv.S = s
	return hv, s, ns
}

func TestShuffleSend(t *testing.T) {
	hv, s, ns := testView(2)

	hv.SendShuffle(ns[1])
	raw := s.Reset()[0]
	m := raw.(*ShuffleRequest)

	require.NotNil(t, m.Active)
	require.NotNil(t, m.Passive)
	require.Equal(t, 0, len(m.Active))
	require.Equal(t, 0, len(m.Passive))
}

func TestShuffleRecv(t *testing.T) {
	hv, _, ns := testView(10)

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
