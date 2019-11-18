package hyparview

type Config struct {
	ActiveSize  int
	ActiveRWL   int
	PassiveSize int
	PassiveRWL  int
	CryptoRand  bool
}

type Hyparview struct {
	Config
	Active  ActiveView
	Passive PassiveView
	Self    Node
}

func CreateView(self Node, active int, passive int, activeRWL int, passiveRWL int) *Hyparview {
	return Hyparview{
		Config: Config{
			ActiveRWL:   activeRWL,
			ActiveSize:  active,
			PassiveRWL:  activeRWL,
			PassiveSize: passive,
		},
		Active:  Active{Nodes: []Node{}, Max: active},
		Passive: Active{Nodes: []Node{}, Max: active},
		Self:    self,
	}
}

func DefaultView(self Node, n size) *Hyparview {
	return CreateView(self, 5, 30, 7, 5)
}

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

func (v *Hyparview) DropRandActive() (ms []Message) {
	i := v.rint(mx)
	node := v.Active.GetIndex(i)
	v.Active.DelIndex(i)
	v.Active.Add(node)
	ms = append(ms, SendDisconnect(node, v.Self))
	return ms
}

func (v *Hyparview) AddActive(node Node) (ms []Message) {
	if node.Equal(v.Self) ||
		node.Contains(node) {
		return ms
	}

	if v.Active.IsFull() {
		ms = DropRandActive()
	}

	v.Active.Add(node)
	return ms
}

func (v *Hyparview) AddPassive(node Node) {
	if node.Equal(v.Self) ||
		v.Active.Contains(node) ||
		v.Passive.Contains(node) {
		return
	}

	if v.Passive.IsFull() {
		i := v.rint(mx)
		v.Passive.DelIndex(i)
	}

	v.Passive.Add(node)
}
