package hyparview

type Config struct {
	ActiveSize     int
	ActiveRWL      int
	PassiveSize    int
	PassiveRWL     int
	ShuffleActive  int
	ShufflePassive int
	ShuffleRWL     int
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
			ActiveRWL:      6,
			ActiveSize:     5,
			PassiveRWL:     3,
			PassiveSize:    30,
			ShuffleActive:  3,
			ShufflePassive: 15,
			ShuffleRWL:     3,
		},
		Active:  CreateViewPart(5),
		Passive: CreateViewPart(30),
		Self:    self,
	}
}

// RecvJoin processes a Join following the paper
func (v *Hyparview) RecvJoin(node *Node) (ms []Message) {
	if v.Active.IsFull() {
		ms = append(ms, v.DropRandActive()...)
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
func (v *Hyparview) RecvNeighbor(priority bool, node *Node) (ms []Message) {
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
	as := v.Active.Shuffled()[0:v.ShuffleActive]
	ps := v.Passive.Shuffled()[0:v.ShufflePassive]
	m := SendShuffle(node, v.Self, as, ps, v.ShuffleRWL)
	v.Shuffle = m
	return append(ms, m)
}

// RecvShuffle processes a shuffle request. Paper
func (v *Hyparview) RecvShuffle(r *ShuffleRequest) (ms []Message) {
	if r.TTL > 0 ||
		!v.Active.IsEmpty() { // FIXME this may be 1
		m := SendShuffle(v.Active.RandNode(), v.Self, r.Active, r.Passive, r.TTL+1)
		ms = append(ms, m)
		return ms
	}

	// Number of peers in the request or all of my passive view
	l := len(r.Active) + len(r.Passive) + 1
	m := v.Passive.Size()
	if l > m {
		l = m
	}

	ps := v.Passive.Shuffled()[0:l]
	ms = append(ms, SendShuffleReply(r.To(), v.Self, ps))

	// Keep the new passive peers
	sent := make([]*Node, l)
	copy(sent, ps) // recvShuffle is going to destructively use this

	v.recvShuffle(r.From, sent)
	for _, n := range r.Active {
		v.recvShuffle(n, sent)
	}
	for _, n := range r.Passive {
		v.recvShuffle(n, sent)
	}

	return ms
}

// recvShuffle processes one node to be added to the passive view
func (v *Hyparview) recvShuffle(n *Node, sent []*Node) {
	if n.Equal(v.Self) || v.Active.Contains(n) || v.Passive.Contains(n) {
		return
	}

	if v.Passive.IsFull() {
		if len(sent) > 0 {
			i := v.Passive.ContainsIndex(sent[0])
			v.Passive.DelIndex(i)
			sent = sent[1:]
		}

		v.Passive.DelIndex(v.Passive.RandIndex())
	}

	v.Passive.Add(n)
}
