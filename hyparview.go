package hyparview

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
	S       Send
	Active  *ViewPart
	Passive *ViewPart
	Self    Node
	// The passive window peers sent in the last shuffle request
	LastShuffle []Node
}

// CreateView creates the view. Configuration is recommendations based on the cluster size
// n. Does not start any process.
func CreateView(s Send, self Node, n int) *Hyparview {
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
		S:       s,
		Active:  CreateViewPart(active),
		Passive: CreateViewPart(passive),
		Self:    self,
	}
}

func (v *Hyparview) SendJoin(peer Node) {
	// Usually on run at bootstrap, where this will never produce disconnect messages
	v.AddActive(peer)
	v.Send(NewJoin(peer, v.Self))
}

// RecvJoin processes a Join following the paper
func (v *Hyparview) RecvJoin(r *JoinRequest) {
	v.AddActive(r.From())

	// Forward to all active peers
	for _, n := range v.Active.Nodes {
		if EqualNode(n, r.From()) {
			continue
		}
		v.Send(NewForwardJoin(n, v.Self, r.From(), v.RWL.Active))
	}
}

// RecvForwardJoin processes a ForwardJoin following the paper
func (v *Hyparview) RecvForwardJoin(r *ForwardJoinRequest) {
	v.repairAsymmetry(r)
	// if r.TTL == 0 || !v.Active.IsFull() {
	if r.TTL == 0 || v.Active.IsEmptyBut(r.From()) {
		if EqualNode(r.Join, v.Self) || v.Active.Contains(r.Join) {
			return
		}

		v.AddActive(r.Join)
		v.Send(NewNeighborJoin(r.Join, v.Self))
		return
	}

	if r.TTL == v.RWL.Passive {
		v.AddPassive(r.Join)
	}

	// Forward to one not-sender active peer
	for _, n := range v.Active.Shuffled() {
		if EqualNode(n, r.From()) || EqualNode(n, r.Join) {
			continue
		}
		v.Send(NewForwardJoin(n, v.Self, r.Join, r.TTL-1))
		break
	}
}

// DropRandActive removes a random active peer and returns the disconnect message following
// the paper
func (v *Hyparview) DropRandActive() {
	idx := v.Active.RandIndex()
	node := v.Active.GetIndex(idx)
	v.Active.DelIndex(idx)
	v.AddPassive(node)
	v.Send(NewDisconnect(node, v.Self))
}

// AddActive adds a node to the active view, possibly dropping an active peer to make room.
// Paper
func (v *Hyparview) AddActive(node Node) {
	if EqualNode(node, v.Self) || v.Active.Contains(node) {
		return
	}

	if v.Active.IsFull() {
		v.DropRandActive()
	}

	v.Active.Add(node)
}

// AddPassive adds a node to the passive view, possibly dropping a passive peer to make
// room. Paper
func (v *Hyparview) AddPassive(node Node) {
	if EqualNode(node, v.Self) ||
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
func (v *Hyparview) DelPassive(node Node) {
	idx := v.Passive.ContainsIndex(node)
	if idx >= 0 {
		v.Passive.DelIndex(idx)
	}
}

// RecvDisconnect processes a disconnect, demoting the sender to the passive view
func (v *Hyparview) RecvDisconnect(r *DisconnectRequest) {
	idx := v.Active.ContainsIndex(r.From())
	if idx >= 0 {
		v.Active.DelIndex(idx)
		v.AddPassive(r.From())
	}
	v.PromotePassiveBut(r.From())
	if v.Active.IsEmpty() {
		v.Bootstrap()
	}
}

// RecvNeighbor processes a neighbor, sent during failure recovery
// Returns at most one NeighborRefuse, which must be replied to the client
func (v *Hyparview) RecvNeighbor(r *NeighborRequest) *NeighborRefuse {
	node := r.From()
	priority := r.Priority
	if v.Active.IsFull() && priority == LowPriority {
		return NewNeighborRefuse(node, v.Self)
	}
	idx := v.Passive.ContainsIndex(node)
	if idx >= 0 {
		v.Passive.DelIndex(idx)
	}

	v.AddActive(node)
	return nil
}

// SendShuffle creates and sends a shuffle request to maintain the passive view
func (v *Hyparview) SendShuffle() {
	node := v.Peer()
	if node == nil {
		// the active view is empty, just ignore the shuffle
		return
	}
	m := v.composeShuffle(node)
	v.LastShuffle = m.Passive
	v.Send(m)
}

// composeShuffle is testable, it updates the view but returns the message
func (v *Hyparview) composeShuffle(node Node) *ShuffleRequest {
	as := v.Active.Shuffled()[:min(v.ShuffleActive, v.Active.Size())]
	ps := v.Passive.Shuffled()[:min(v.ShufflePassive, v.Passive.Size())]
	return NewShuffle(node, v.Self, v.Self, as, ps, v.RWL.Shuffle)
}

// RecvShuffle processes a shuffle request. Paper
func (v *Hyparview) RecvShuffle(r *ShuffleRequest) {
	v.repairAsymmetry(r)
	// If the active view size is one, it means that our only active peer is sender of
	// this shuffle message
	if r.TTL >= 0 && !v.Active.IsEmptyBut(r.From()) {
		// Forward to one active non-sender
		for _, n := range v.Active.Shuffled() {
			if EqualNode(n, r.From()) || EqualNode(n, r.Origin) {
				continue
			}
			v.Send(NewShuffle(n, v.Self, r.Origin, r.Active, r.Passive, r.TTL-1))
			break
		}
		return
	}

	// min(configured length of the shuffle request, my passive view)
	// FIXME the paper says "the number of peers in the request", but it's clear that
	// passive peers are distributed in the network more quickly if we use the number
	// that should have been in the request
	l := v.Config.ShuffleActive + v.Config.ShufflePassive + 1
	p := v.Passive.Size()
	if l > p {
		l = p
	}

	// Send back l shuffled results
	ps := v.Passive.Shuffled()[0:l]
	v.Send(NewShuffleReply(r.Origin, v.Self, ps))

	v.addShuffle(r.Origin, ps)
	for _, n := range r.Active {
		v.addShuffle(n, ps)
	}
	for _, n := range r.Passive {
		v.addShuffle(n, ps)
	}

	// v.greedyShuffle()
}

// addShuffle processes one node to be added to the passive view. If the node is us or
// otherwise known, ignore it. If passive is full, eject first one of the nodes we sent then
// a node at random to make room. Changes the list of exchanged peers in place
func (v *Hyparview) addShuffle(n Node, exchanged []Node) {
	if EqualNode(n, v.Self) || v.Active.Contains(n) || v.Passive.Contains(n) {
		return
	}

	if v.Passive.IsFull() {
		idx := -1

		for len(exchanged) > 0 && idx < 0 {
			idx = v.Passive.ContainsIndex(exchanged[0])
			exchanged = exchanged[1:]
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
		v.addShuffle(n, v.LastShuffle)
	}
}

// Recv is a helper method that dispatches to the correct recv
func (v *Hyparview) Recv(m Message) *NeighborRefuse {
	switch m1 := m.(type) {
	case *JoinRequest:
		v.RecvJoin(m1)
		// if len(ms) > v.Active.Max {
		// 	log.Printf("DEBUG join %d\n", len(ms))
		// }
	case *ForwardJoinRequest:
		v.RecvForwardJoin(m1)
		// if len(ms) > 1 {
		// 	log.Printf("DEBUG forward %d\n", len(ms))
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
func (v *Hyparview) Peer() Node {
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
