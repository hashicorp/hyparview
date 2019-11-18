package hyparview

type Hyparview struct {
	Active     ActiveView
	ActiveRWL  int
	Passive    PassiveView
	PassiveRWL int
	Self       Node
}

type Node string

func (v *Hyparview) Recv(message *Message) {
	switch Message.Command {
	case Join:
		ms := RecvJoin(Message.From)
	case ForwardJoin:
		ms := RecvForwardJoin(Message.Data, Message.TTL, Message.From)
	}
}

func (v *Hyparview) RecvJoin(node Node) (ms []Message) {
	if v.Active.IsFull() {
		v.Active.DropRando()
	}

	v.Active.Add(node)

	for _, n := range v.Active.Nodes {
		if n.Equal(node) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, node, v.ActiveRWL, v.Self))
	}
	return ms
}

func (v *Hyparview) RecvForwardJoin(node Node, ttl int, sender Node) (ms []Message) {
	if ttl == 0 || len(v.Active) == 0 {
		v.Active.Add(node)
	} else {
		if ttl == v.PassiveRWL {
			v.Passive.Add(node)
		}
	}

	for _, n := range v.Active.Nodes {
		if n.Equal(sender) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, node, ttl-1, v.Self))
	}
	return ms
}
