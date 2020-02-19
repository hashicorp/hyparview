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

type Sender func(...Message)

type Hyparview struct {
	Config
	Active  *ViewPart
	Passive *ViewPart
	Self    *Node
	Send    Sender
	// The passive window peers sent in the last shuffle request
	LastShuffle []*Node
}

// CreateView creates the view. Configuration is recommendations based on the cluster size
// n. Does not start any process.
func CreateView(self *Node, n int) *Hyparview {
	active := 5
	passive := 30

	return &Hyparview{
		Config: Config{
			ShuffleActive:  3,
			ShufflePassive: 6,
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
func (v *Hyparview) RecvJoin(r *JoinRequest) {
	if v.Active.IsFull() {
		v.DropRandActive()
	}

	// Forward to all active peers
	for _, n := range v.Active.Nodes {
		if n.Equal(r.From) {
			continue
		}
		v.Send(SendForwardJoin(n, v.Self, r.From, v.RWL.Active))
	}

	v.Active.Add(r.From)
}

// RecvForwardJoin processes a ForwardJoin following the paper
func (v *Hyparview) RecvForwardJoin(r *ForwardJoinRequest) {
	if r.TTL == 0 || v.Active.IsEmpty() {
		v.AddActive(r.Join)
		return
	}

	if r.TTL == v.RWL.Passive {
		v.AddPassive(r.Join)
	}

	// Forward to one not-sender active peer
	for _, n := range v.Active.Shuffled() {
		if n.Equal(r.From) {
			continue
		}
		v.Send(SendForwardJoin(n, v.Self, r.Join, r.TTL-1))
		break
	}
}

// DropRandActive removes a random active peer and returns the disconnect message following
// the paper
func (v *Hyparview) DropRandActive() {
	idx := v.Active.RandIndex()
	node := v.Active.GetIndex(idx)
	v.Active.DelIndex(idx)
	v.Passive.Add(node)
	v.Send(SendDisconnect(node, v.Self))
	return
}

// AddActive adds a node to the active view, possibly dropping an active peer to make room.
// Paper
func (v *Hyparview) AddActive(node *Node) {
	if node.Equal(v.Self) {
		return
	}

	if v.Active.IsFull() {
		v.DropRandActive()
	}

	v.Active.Add(node)
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
// Return at most one NeighborRefuse, which must be returned to the client
func (v *Hyparview) RecvNeighbor(r *NeighborRequest) *NeighborRefuse {
	node := r.From
	priority := r.Priority
	if v.Active.IsFull() && priority == LowPriority {
		return SendNeighborRefuse(node, v.Self)
	}
	idx := v.Passive.ContainsIndex(node)
	if idx >= 0 {
		v.Passive.DelIndex(idx)
	}
	v.AddActive(node)

	return nil
}

// SendShuffle creates the periodic state to mark and message for maintaining the passive
// view. Paper
func (v *Hyparview) SendShuffle(node *Node) {
	as := v.Active.Shuffled()[:min(v.ShuffleActive, v.Active.Size())]
	ps := v.Passive.Shuffled()[:min(v.ShufflePassive, v.Passive.Size())]
	v.LastShuffle = ps
	v.Send(SendShuffle(node, v.Self, as, ps, v.RWL.Shuffle))
}

// RecvShuffle processes a shuffle request. Paper
func (v *Hyparview) RecvShuffle(r *ShuffleRequest) {
	if r.TTL >= 0 && !v.Active.IsEmpty() { // FIXME this may be 1
		// Forward to one active non-sender
		for _, n := range v.Active.Shuffled() {
			if n.Equal(r.From) {
				continue
			}
			v.Send(SendShuffle(n, r.From, r.Active, r.Passive, r.TTL-1))
			break
		}
		return
	}

	// min(Number of peers in the request, my passive view)
	l := len(r.Active) + len(r.Passive) + 1
	m := v.Passive.Size()
	if l > m {
		l = m
	}

	// Send back l shuffled results
	// FIXME this maybe should be the number of configured peers, not the number sent
	ps := v.Passive.Shuffled()[0:l]
	v.Send(SendShuffleReply(r.From, v.Self, ps))

	// Keep the sent passive peers
	// addShuffle is going to destructively use this

	v.addShuffle(r.From)
	for _, n := range r.Active {
		v.addShuffle(n)
	}
	for _, n := range r.Passive {
		v.addShuffle(n)
	}
}

// addShuffle processes one node to be added to the passive view. If the node is us or
// otherwise known, ignore it. If passive is full, eject first one of the nodes we sent then
// a node at random to make room.
func (v *Hyparview) addShuffle(n *Node) {
	if n.Equal(v.Self) || v.Active.Contains(n) || v.Passive.Contains(n) {
		return
	}

	if v.Passive.IsFull() {
		idx := -1

		for len(v.LastShuffle) > 0 && idx < 0 {
			idx = v.Passive.ContainsIndex(v.LastShuffle[0])
			v.LastShuffle = v.LastShuffle[1:]
		}

		if idx < 0 {
			idx = v.Passive.RandIndex()
		}

		v.Passive.DelIndex(idx)
	}

	v.Passive.Add(n)
}

func (v *Hyparview) RecvShuffleReply(r *ShuffleReply) {
	for _, n := range r.Passive {
		v.addShuffle(n)
	}
}

// Recv is a helper method that dispatches to the correct recv
func (v *Hyparview) Recv(m Message) *NeighborRefuse {
	switch m1 := m.(type) {
	case *JoinRequest:
		v.RecvJoin(m1)
		// if len(ms) > v.Active.Max {
		// 	fmt.Printf("JOIN %d\n", len(ms))
		// }
	case *ForwardJoinRequest:
		v.RecvForwardJoin(m1)
		// if len(ms) > 1 {
		// 	fmt.Printf("FORWARD %d\n", len(ms))
		// }
	case *DisconnectRequest:
		v.RecvDisconnect(m1)
	case *NeighborRequest:
		return v.RecvNeighbor(m1)
	case *ShuffleRequest:
		v.RecvShuffle(m1)
	case *ShuffleReply:
		v.RecvShuffleReply(m1)
	default:
		// log unimplemented?
	}
	return nil
}

// Peer returns a random active peer
func (v *Hyparview) Peer() *Node {
	if v.Active.IsEmpty() {
		return nil
	}
	return v.Active.RandNode()
}

// Copy returns a copy thats safe to modify. Shuffle is copied as a pointer because each
// ShuffleRequest is immutable once created
func (v *Hyparview) Copy() *Hyparview {
	out := *v
	out.Active = v.Active.Copy()
	out.Passive = v.Passive.Copy()
	return &out
}
