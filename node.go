package hyparview

type Node interface {
	Addr() string
}

type node struct {
	addr string
}

func (n *node) Addr() string {
	return n.addr
}

func NewNode(addr string) Node {
	return &node{addr: addr}
}

func EqualNode(n, m Node) bool {
	if n == nil || m == nil {
		return n == m
	}
	return n.Addr() == m.Addr()
}
