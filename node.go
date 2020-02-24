package hyparview

type Node struct {
	ID   string
	Addr string
}

func NewNode(addr string) *Node {
	return &Node{
		ID:   addr,
		Addr: addr,
	}
}

func (n *Node) Equal(m *Node) bool {
	if n == nil || m == nil {
		return n == m
	}
	return n.ID == m.ID // FIXME both?
}
