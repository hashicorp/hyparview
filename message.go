package hyparview

type Message struct {
	To     *Node
	From   *Node
	Data   *Node
	Action string
	TTL    int
}

const (
	Join        = "j"
	ForwardJoin = "f"
	Disconnect  = "d"
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
