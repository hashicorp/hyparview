package hyparview

// import "fmt"

type ConfigRandomWalkLength struct {
	Active  int
	Passive int
	Shuffle int
}

type Config struct {
	ShuffleActive  int
	ShufflePassive int
	RWL            ConfigRandomWalkLength
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
	active := 5
	passive := 30

	return &Hyparview{
		Config: Config{
			ShuffleActive:  3,
			ShufflePassive: 4,
			RWL: ConfigRandomWalkLength{
				Active:  6,
				Passive: 3,
				Shuffle: 6,
			},
		},
		Active:  CreateViewPart(active),
		Passive: CreateViewPart(passive),
		Self:    self,
	}
}

// RecvJoin processes a Join following the paper
func (v *Hyparview) RecvJoin(r *JoinRequest) (ms []Message) {
	if v.Active.IsFull() {
		ms = append(ms, v.DropRandActive()...)
	}

	// Forward to all active peers
	for _, n := range v.Active.Nodes {
		ms = append(ms, SendForwardJoin(n, v.Self, r.From, v.RWL.Active))
	}

	v.Active.Add(r.From)
	return ms
}

// RecvForwardJoin processes a ForwardJoin following the paper
func (v *Hyparview) RecvForwardJoin(r *ForwardJoinRequest) (ms []Message) {
	node := r.Join
	sender := r.From
	ttl := r.TTL

	if ttl == 0 || v.Active.IsEmpty() {
		v.AddActive(node) // Don't bother to catch disconnect, there won't be any
		// Stop on active empty because who else am I going to send it to
		return ms
	}

	if ttl == v.RWL.Passive {
		v.AddPassive(node)
	}

	// Forward to one not-sender active peer
	for _, n := range v.Active.Shuffled() {
		if n.Equal(sender) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, v.Self, node, ttl-1))
		break
	}

	return ms
}

// DropRandActive removes a random active peer and returns the disconnect message following
// the paper
func (v *Hyparview) DropRandActive() (ms []Message) {
	i := Rint(v.Active.Size() - 1)
	node := v.Active.GetIndex(i)
	v.Active.DelIndex(i)
	ms = append(ms, SendDisconnect(node, v.Self))
	return ms
}

// AddActive adds a node to the active view, possibly dropping an active peer to make room.
// Paper
func (v *Hyparview) AddActive(node *Node) (ms []Message) {
	if node.Equal(v.Self) {
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
		i := Rint(v.Passive.Size() - 1)
		v.Passive.DelIndex(i)
	}

	v.Passive.Add(node)
}

// DelPassive is a helper function to delete the node from the passive view
func (v *Hyparview) DelPassive(node *Node) {
	idx := v.Passive.ContainsIndex(node)
	if idx >= 0 {
		v.Passive.DelIndex(idx)
	}
}

// RecvDisconnect processes a disconnect, demoting the sender to the passive view
func (v *Hyparview) RecvDisconnect(r *DisconnectRequest) {
	node := r.From
	idx := v.Active.ContainsIndex(node)
	if idx >= 0 {
		v.Active.DelIndex(idx)
		v.AddPassive(node)
	}
}

// RecvNeighbor processes a neighbor, sent during failure recovery
func (v *Hyparview) RecvNeighbor(r *NeighborRequest) (ms []Message) {
	node := r.From
	priority := r.Priority
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
func (v *Hyparview) SendShuffle(node *Node) *ShuffleRequest {
	as := v.Active.Shuffled()[:min(v.ShuffleActive, v.Active.Size())]
	ps := v.Passive.Shuffled()[:min(v.ShufflePassive, v.Passive.Size())]
	m := SendShuffle(node, v.Self, as, ps, v.RWL.Shuffle)
	v.Shuffle = m
	return m
}

// RecvShuffle processes a shuffle request. Paper
func (v *Hyparview) RecvShuffle(r *ShuffleRequest) (ms []Message) {
	if r.TTL >= 0 && !v.Active.IsEmpty() { // FIXME this may be 1
		m := SendShuffle(v.Active.RandNode(), r.From, r.Active, r.Passive, r.TTL-1)
		ms = append(ms, m)
		return ms
	}

	// min(Number of peers in the request, my passive view)
	l := len(r.Active) + len(r.Passive) + 1
	m := v.Passive.Size()
	if l > m {
		l = m
	}

	ps := v.Passive.Shuffled()[0:l]
	ms = append(ms, SendShuffleReply(r.From, v.Self, ps))

	// Keep the sent passive peers
	// addShuffle is going to destructively use this
	sent := make([]*Node, l)
	copy(sent, ps)
	new := v.Passive.Copy()

	v.addShuffle(r.From, sent, new)
	for _, n := range r.Active {
		v.addShuffle(n, sent, new)
	}
	for _, n := range r.Passive {
		v.addShuffle(n, sent, new)
	}
	v.Passive = new

	return ms
}

// addShuffle processes one node to be added to the passive view. If the node is us or
// otherwise known, ignore it. If passive is full, eject first one of the nodes we sent then
// a node at random to make room.
func (v *Hyparview) addShuffle(n *Node, sent []*Node, passive *ViewPart) {
	if n.Equal(v.Self) || v.Active.Contains(n) || passive.Contains(n) {
		return
	}

	if passive.IsFull() {
		idx := -1
		if len(sent) > 0 {
			idx = passive.ContainsIndex(sent[0])
		}
		if idx >= 0 {
			sent = sent[1:]
		} else {
			idx = passive.RandIndex()
		}
		passive.DelIndex(idx)
	}

	passive.Add(n)
}

func (v *Hyparview) RecvShuffleReply(r *ShuffleReply) {
	var sent []*Node
	if v.Shuffle != nil {
		sent = v.Shuffle.Passive
	}
	v.Shuffle = nil

	new := v.Passive.Copy()
	for _, n := range r.Passive {
		v.addShuffle(n, sent, new)
	}
	v.Passive = new
}

// Recv is a helper method that dispatches to the correct recv
func (v *Hyparview) Recv(m Message) (ms []Message) {
	switch m1 := m.(type) {
	case *JoinRequest:
		ms = v.RecvJoin(m1)
		// if len(ms) > v.Active.Max {
		// 	fmt.Printf("JOIN %d\n", len(ms))
		// }
		return ms
	case *ForwardJoinRequest:
		ms = v.RecvForwardJoin(m1)
		// if len(ms) > 1 {
		// 	fmt.Printf("FORWARD %d\n", len(ms))
		// }
		return ms
	case *DisconnectRequest:
		v.RecvDisconnect(m1)
	case *NeighborRequest:
		return v.RecvNeighbor(m1)
	case *ShuffleRequest:
		return v.RecvShuffle(m1)
	case *ShuffleReply:
		v.RecvShuffleReply(m1)
	default:
		// log unimplemented?
	}
	return ms
}

// Peer returns a random active peer
func (v *Hyparview) Peer() *Node {
	if v.Active.IsEmpty() {
		return nil
	}
	return v.Active.RandNode()
}
