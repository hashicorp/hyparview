package hyparview

type Node struct {
	ID   string
	Addr string
}

func (n *Node) Equal(m Node) bool {
	return n.ID == m.ID &&
		n.Addr == m.Addr // FIXME both?
}
