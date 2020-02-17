package hyparview

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShuffleRecv(t *testing.T) {
	ns := make([]*Node, 10)
	for i := 0; i < 10; i++ {
		ns[i] = &Node{
			Addr: fmt.Sprintf("127.0.0.1:1000%d", i),
		}
	}

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
