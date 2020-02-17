package hyparview

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
)

func nodes(n int) []*Node {
	ns := make([]*Node, n)
	for i := 0; i < n; i++ {
		ns[i] = &Node{
			Addr: fmt.Sprintf("127.0.0.1:1000%d", i),
		}
	}
	return ns
}

func TestShuffleSend(t *testing.T) {
	ns := nodes(2)
	hv := CreateView(ns[0], 0)
	m := hv.SendShuffle(ns[1])
	pretty.Log(m)
}

func TestShuffleRecv(t *testing.T) {
	ns := nodes(10)
	hv := CreateView(ns[0], 0)

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
