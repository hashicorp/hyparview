package hyparview

const (
	HighPriority = true
	LowPriority  = false
)

type JoinRequest struct {
	to   *Node
	from *Node
}

func SendJoin(to *Node, from *Node) *JoinRequest {
	return &JoinRequest{
		to:   to,
		from: from,
	}
}

type ForwardJoinRequest struct {
	to   *Node
	from *Node
	Join *Node
	TTL  int
}

func SendForwardJoin(to *Node, from *Node, join *Node, ttl int) *ForwardJoinRequest {
	return &ForwardJoinRequest{
		to:   to,
		from: from,
		Join: join,
		TTL:  ttl,
	}
}

type DisconnectRequest struct {
	to   *Node
	from *Node
}

func SendDisconnect(to *Node, from *Node) *DisconnectRequest {
	return &DisconnectRequest{
		to:   to,
		from: from,
	}
}

type NeighborRequest struct {
	to       *Node
	from     *Node
	Priority bool
	Join     bool
}

func SendNeighbor(to *Node, from *Node, priority bool) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		from:     from,
		Priority: priority,
	}
}

func SendNeighborJoin(to *Node, from *Node) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		from:     from,
		Priority: HighPriority,
		Join:     true,
	}
}

type NeighborRefuse struct {
	to   *Node
	from *Node
}

func SendNeighborRefuse(to *Node, from *Node) *NeighborRefuse {
	return &NeighborRefuse{
		to:   to,
		from: from,
	}
}

type ShuffleRequest struct {
	to      *Node
	from    *Node
	Origin  *Node
	Active  []*Node
	Passive []*Node
	TTL     int
}

func SendShuffle(to, from, origin *Node, active, passive []*Node, ttl int) *ShuffleRequest {
	return &ShuffleRequest{
		to:      to,
		from:    from,
		Origin:  from,
		Active:  active,
		Passive: passive,
		TTL:     ttl,
	}
}

type ShuffleReply struct {
	to      *Node
	from    *Node
	Passive []*Node
}

func SendShuffleReply(to *Node, from *Node, passive []*Node) *ShuffleReply {
	return &ShuffleReply{
		to:      to,
		from:    from,
		Passive: passive,
	}
}

type Gossip struct {
	to      *Node
	from    *Node
	Payload int
	Hops    int
}
