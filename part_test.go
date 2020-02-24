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
}
