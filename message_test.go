package hyparview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssocTo(t *testing.T) {
	a := NewNode("a")
	b := NewNode("b")
	c := NewNode("c")
	m := NewJoin(a, b)
	n := m.AssocTo(c)

	require.NotEqual(t, m, n)
	require.Equal(t, c, n.To())
}
