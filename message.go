package hyparview

type Message struct {
	To      Node
	From    Node
	Data    Node
	Command Command
	TTL     int
}

const (
	Join        = "j"
	ForwardJoin = "f"
)

func SendForwardJoin(to Node, payload Node, ttl int, from Node) Message {
	return Message{
		To:   to,
		From: from,
		Data: payload,
		TTL:  ttl,
	}
}
