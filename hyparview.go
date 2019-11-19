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
	Active  *ActiveView
	Passive *ActiveView
	Self    *Node
}

// CreateView creates the view, does not start any process
func CreateView(self *Node, active int, passive int, activeRWL int, passiveRWL int) *Hyparview {
	return &Hyparview{
		Config: Config{
			ActiveRWL:   activeRWL,
			ActiveSize:  active,
			PassiveRWL:  activeRWL,
			PassiveSize: passive,
		},
		Active:  CreateActiveView(active),
		Passive: CreateActiveView(passive),
		Self:    self,
	}
}

// DefaultView calls CreateView with the recommended values for a cluster of size n
func DefaultView(self *Node, n int) *Hyparview {
	return CreateView(self, 5, 30, 7, 5)
}

// Recv dispatches the message, returning the resulting outgoing messages
func (v *Hyparview) Recv(message *Message) (ms []Message) {
	switch message.Action {
	case Join:
		ms = v.RecvJoin(message.From)
	case ForwardJoin:
		ms = v.RecvForwardJoin(message.Data, message.TTL, message.From)
	case Disconnect:
		v.RecvDisconnect(message.From)
	}
	return ms
}

// RecvJoin processes a Join following the paper
func (v *Hyparview) RecvJoin(node *Node) (ms []Message) {
	if v.Active.IsFull() {
		v.DropRandActive()
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

// RecvForwardJoin processes a ForwardJoin following the paper
func (v *Hyparview) RecvForwardJoin(node *Node, ttl int, sender *Node) (ms []Message) {
	if ttl == 0 || v.Active.IsEmpty() {
		ms = append(ms, v.AddActive(node)...)
	} else if ttl == v.PassiveRWL {
		v.AddPassive(node)
	}

	for _, n := range v.Active.Nodes {
		if n.Equal(sender) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, node, ttl-1, v.Self))
	}
	return ms
}

// DropRandActive removes a random active peer and returns the disconnect message following
// the paper
func (v *Hyparview) DropRandActive() (ms []Message) {
	i := v.rint(v.Active.Size())
	node := v.Active.GetIndex(i)
	v.Active.DelIndex(i)
	v.Active.Add(node)
	ms = append(ms, SendDisconnect(node, v.Self))
	return ms
}

// AddActive adds a node to the active view, possibly dropping an active peer to make room.
// Paper
func (v *Hyparview) AddActive(node *Node) (ms []Message) {
	if node.Equal(v.Self) ||
		v.Active.Contains(node) {
		return ms
	}

	if v.Active.IsFull() {
		ms = v.DropRandActive()
	}

	v.Active.Add(node)
	return ms
}

// AddPassive adds a node to the passive view, possibly dropping a passive peer to make
// room. Paper
func (v *Hyparview) AddPassive(node *Node) {
	if node.Equal(v.Self) ||
		v.Active.Contains(node) ||
		v.Passive.Contains(node) {
		return
	}

	if v.Passive.IsFull() {
		i := v.rint(v.Passive.Size())
		v.Passive.DelIndex(i)
	}

	v.Passive.Add(node)
}

// RecvDisconnect processes a disconnect, demoting the sender to the passive view
func (v *Hyparview) RecvDisconnect(node *Node) {
	idx := v.Active.ContainsIndex(node)
	if idx > 0 {
		v.Active.DelIndex(idx)
		v.AddPassive(node)
	}
}
