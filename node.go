package hyparview

type Node struct {
	ID   string
	Addr string
}

func (n *Node) Equal(m *Node) bool {
	if n == nil || m == nil {
		return n == m
	}
	return n.Addr == m.Addr // FIXME both?
}
