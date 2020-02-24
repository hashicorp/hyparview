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

	v.DelIndex(1)
	require.False(t, v.Contains(n))

	v.Add(NewNode("c"))
	n = NewNode("d")
	v.Add(n)
	v.Add(NewNode("e"))

	v.DelNode(n)

	require.Equal(t, "a", v.Nodes[0].ID)
	require.Equal(t, "c", v.Nodes[1].ID)
	require.Equal(t, "e", v.Nodes[2].ID)
}
