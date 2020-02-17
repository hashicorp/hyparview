package hyparview

type Node struct {
	ID   string
	Addr string
}

func (n *Node) Equal(m *Node) bool {
	return n.Addr == m.Addr // FIXME both?
}
