package hyparview

type Action string
type Priority int

type Message interface {
	From() *Node
}

type Message struct {
	To     *Node
	From   *Node
	Data   *Node
	Action Action
	TTL    int
}

const (
	Join           = "j"
	ForwardJoin    = "f"
	Disconnect     = "d"
	Neighbor       = "n"
	NeighborRefuse = "r"
	HighPriority   = 1
	LowPriority    = 0
)

func SendJoin(to *Node, from *Node) Message {
	return Message{
		Action: Join,
		To:     to,
		From:   from,
	}
}

func SendForwardJoin(to *Node, payload *Node, ttl int, from *Node) Message {
	return Message{
		Action: ForwardJoin,
		To:     to,
		From:   from,
		Data:   payload,
		TTL:    ttl,
	}
}

func SendDisconnect(to *Node, from *Node) Message {
	return Message{
		Action: Disconnect,
		To:     to,
		From:   from,
	}
}

func SendNeighbor(to *Node, priority Priority, from *Node) Message {
	return Message{
		Action: Neighbor,
		To:     to,
		From:   from,
		TTL:    int(priority),
	}
}

func SendNeighborRefuse(to *Node, from *Node) Message {
	return Message{
		Action: NeighborRefuse,
		To:     to,
		From:   from,
	}
}

type ShuffleRequest struct {
	To      *Node
	From    *Node
	Active  []*Node
	Passive []*Node
}

func (m *ShuffleRequest) From() *Node {
	return m.From
}
