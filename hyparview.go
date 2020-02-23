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

func (v *Hyparview) SendJoin(peer *Node) (ms []Message) {
	// Usually on run at bootstrap, where this will never produce disconnect messages
	ms = append(ms, v.AddActive(peer)...)
	ms = append(ms, SendJoin(peer, v.Self))
	return ms
}

// RecvJoin processes a Join following the paper
func (v *Hyparview) RecvJoin(r *JoinRequest) (ms []Message) {
	ms = append(ms, v.AddActive(r.From)...)

	// Forward to all active peers
	for _, n := range v.Active.Nodes {
		if n.Equal(r.From) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, v.Self, r.From, v.RWL.Active))
	}

	return ms
}

// RecvForwardJoin processes a ForwardJoin following the paper
func (v *Hyparview) RecvForwardJoin(r *ForwardJoinRequest) (ms []Message) {
	if r.TTL == 0 || !v.Active.IsFull() {
		ms = append(ms, v.AddActive(r.Join)...)
		ms = append(ms, SendNeighbor(r.Join, v.Self, HighPriority))
		return ms
	}

	if r.TTL == v.RWL.Passive {
		v.AddPassive(r.Join)
	}

	// Forward to one not-sender active peer
	for _, n := range v.Active.Shuffled() {
		if n.Equal(r.From) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, v.Self, r.Join, r.TTL-1))
		break
	}

	return ms
}

// DropRandActive removes a random active peer and returns the disconnect message following
// the paper
func (v *Hyparview) DropRandActive() (ms []Message) {
	idx := v.Active.RandIndex()
	node := v.Active.GetIndex(idx)
	v.Active.DelIndex(idx)
	v.AddPassive(node)
	ms = append(ms, SendDisconnect(node, v.Self))
	return ms
}

// AddActive adds a node to the active view, possibly dropping an active peer to make room.
// Paper
func (v *Hyparview) AddActive(node *Node) (ms []Message) {
	if node.Equal(v.Self) {
		return
	}

	if v.Active.IsFull() {
		ms = append(ms, v.DropRandActive()...)
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
		i := v.Passive.RandIndex()
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
	v.LastShuffle = ps
	return SendShuffle(node, v.Self, as, ps, v.RWL.Shuffle)
}

// RecvShuffle processes a shuffle request. Paper
func (v *Hyparview) RecvShuffle(r *ShuffleRequest) (ms []Message) {
	// If the active view size is one, it means that our only active peer is sender of
	// this shuffle message
	if r.TTL >= 0 && !v.Active.IsEmptyBut(r.From) {
		// Forward to one active non-sender
		for _, n := range v.Active.Shuffled() {
			if n.Equal(r.From) {
				continue
			}
			ms = append(ms, SendShuffle(n, r.From, r.Active, r.Passive, r.TTL-1))
			break
		}
		return ms
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
	ms = append(ms, SendShuffleReply(r.From, v.Self, ps))

	// Keep the sent passive peers
	// addShuffle is going to destructively use this

	v.addShuffle(r.From)
	for _, n := range r.Active {
		v.addShuffle(n)
	}
	for _, n := range r.Passive {
		v.addShuffle(n)
	}

	return ms
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

	v.AddPassive(n)
}

func (v *Hyparview) RecvShuffleReply(r *ShuffleReply) {
	for _, n := range r.Passive {
		v.addShuffle(n)
	}
}

// Recv is a helper method that dispatches to the correct recv
func (v *Hyparview) Recv(m Message) (ms []Message) {
	switch m1 := m.(type) {
	case *JoinRequest:
		ms = append(ms, v.RecvJoin(m1)...)
		// if len(ms) > v.Active.Max {
		// 	fmt.Printf("JOIN %d\n", len(ms))
		// }
	case *ForwardJoinRequest:
		ms = append(ms, v.RecvForwardJoin(m1)...)
		// if len(ms) > 1 {
		// 	fmt.Printf("FORWARD %d\n", len(ms))
		// }
	case *DisconnectRequest:
		v.RecvDisconnect(m1)
	case *NeighborRequest:
		ms = append(ms, v.RecvNeighbor(m1)...)
	case *ShuffleRequest:
		ms = append(ms, v.RecvShuffle(m1)...)
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

// Copy returns a copy thats safe to modify. Shuffle is copied as a pointer because each
// ShuffleRequest is immutable once created
func (v *Hyparview) Copy() *Hyparview {
	out := *v
	out.Active = v.Active.Copy()
	out.Passive = v.Passive.Copy()
	return &out
}
