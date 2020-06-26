package hyparview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	v := CreateViewPart(2)
	n := NewNode("a")
	v.Add(n)
	require.Equal(t, 0, v.ContainsIndex(n))

	n = NewNode("b")
	v.Add(n)
	require.Equal(t, 1, v.ContainsIndex(n))
	require.True(t, v.IsFull())

	v.DelIndex(1)
	require.False(t, v.Contains(n))
	require.False(t, v.IsFull())

	v.Add(NewNode("c"))
	n = NewNode("d")
	v.Add(n)
	v.Add(NewNode("e"))

	v.DelNode(n)

	require.Equal(t, "a", v.Nodes[0].Addr())
	require.Equal(t, "c", v.Nodes[1].Addr())
	require.Equal(t, "e", v.Nodes[2].Addr())

	v.DelNode(v.RandNode())
	v.DelNode(v.RandNode())
	v.DelNode(v.RandNode())

	require.True(t, v.IsEmpty())
}
