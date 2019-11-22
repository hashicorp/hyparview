package hyparview

type Config struct {
	ActiveSize     int
	ActiveRWL      int
	PassiveSize    int
	PassiveRWL     int
	ShuffleActive  int
	ShufflePassive int
}

type Hyparview struct {
	Config
	Active  *ViewPart
	Passive *ViewPart
	Self    *Node
	Shuffle *ShuffleRequest
}

// CreateView creates the view. Configuration is recommendations based on the cluster size
// n. Does not start any process.
func CreateView(self *Node, n int) *Hyparview {
	return &Hyparview{
		Config: Config{
			ActiveRWL:      7,
			ActiveSize:     5,
			PassiveRWL:     5,
			PassiveSize:    30,
			ShuffleActive:  3,
			ShufflePassive: 15,
		},
		Active:  CreateViewPart(5),
		Passive: CreateViewPart(30),
		Self:    self,
	}
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
	i := rint(v.Active.Size())
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
		i := rint(v.Passive.Size())
		v.Passive.DelIndex(i)
	}

	v.Passive.Add(node)
}

// RecvDisconnect processes a disconnect, demoting the sender to the passive view
func (v *Hyparview) RecvDisconnect(node *Node) {
	idx := v.Active.ContainsIndex(node)
	if idx >= 0 {
		v.Active.DelIndex(idx)
		v.AddPassive(node)
	}
}

// RecvNeighbor processes a neighbor, sent during failure recovery
func (v *Hyparview) RecvNeighbor(priority Priority, node *Node) (ms []Message) {
	if v.Active.IsFull() && priority == LowPriority {
		ms = append(ms, SendNeighborRefuse(node, v.Self))
		return ms
	}
	idx := v.Passive.ContainsIndex(node)
	if idx >= 0 {
		v.Passive.DelIndex(idx)
	}
	return v.AddActive(node)
}

// SendShuffle creates the periodic state to mark and message for maintaining the passive
// view. Paper
func (v *Hyparview) SendShuffle(node *Node) (ms []Message) {
	ns := []*Node{v.Self}
	as := v.Active.Shuffled()[0:v.ShuffleActive]
	ps := v.Passive.Shuffled()[0:v.ShufflePassive]
	// ns = append(ns, as...)
	// ns = append(ns, ps...)

	req := ShuffleRequest{
		To:      node,
		From:    v.Self,
		Active:  as,
		Passive: ps,
	}

	v.Shuffle = &req

	return append(ms, req)
}
