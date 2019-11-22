package hyparview

type Message interface {
	To() *Node
}

const (
	HighPriority = true
	LowPriority  = false
)

type JoinRequest struct {
	to   *Node
	From *Node
}

func (r *JoinRequest) To() *Node { return r.to }

func SendJoin(to *Node, from *Node) JoinRequest {
	return JoinRequest{
		to:   to,
		From: from,
	}
}

type ForwardJoinRequest struct {
	to   *Node
	From *Node
	Join *Node
	TTL  int
}

func (r *ForwardJoinRequest) To() *Node { return r.to }

func SendForwardJoin(to *Node, join *Node, ttl int, from *Node) *ForwardJoinRequest {
	return &ForwardJoinRequest{
		to:   to,
		From: from,
		Join: join,
		TTL:  ttl,
	}
}

type DisconnectRequest struct {
	to   *Node
	From *Node
}

func (r *DisconnectRequest) To() *Node { return r.to }

func SendDisconnect(to *Node, from *Node) *DisconnectRequest {
	return &DisconnectRequest{
		to:   to,
		From: from,
	}
}

type NeighborRequest struct {
	to       *Node
	From     *Node
	Priority bool
}

func (r *NeighborRequest) To() *Node { return r.to }

func SendNeighbor(to *Node, priority bool, from *Node) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		From:     from,
		Priority: priority,
	}
}

type NeighborRefuse struct {
	to   *Node
	From *Node
}

func (r *NeighborRefuse) To() *Node { return r.to }

func SendNeighborRefuse(to *Node, from *Node) *NeighborRefuse {
	return &NeighborRefuse{
		to:   to,
		From: from,
	}
}

type ShuffleRequest struct {
	to      *Node
	From    *Node
	Active  []*Node
	Passive []*Node
	TTL     int
}

func (m *ShuffleRequest) To() *Node { return m.to }

func SendShuffle(to *Node, from *Node, active []*Node, passive []*Node, ttl int) *ShuffleRequest {
	return &ShuffleRequest{
		to:      to,
		From:    from,
		Active:  active,
		Passive: passive,
		TTL:     ttl,
	}
}

type ShuffleReply struct {
	to      *Node
	From    *Node
	Passive []*Node
}

func (m *ShuffleReply) To() *Node { return m.to }

func SendShuffleReply(to *Node, from *Node, passive []*Node) *ShuffleReply {
	return &ShuffleReply{
		to:      to,
		From:    from,
		Passive: passive,
	}
}
