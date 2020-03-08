package hyparview

const (
	HighPriority = true
	LowPriority  = false
)

type JoinRequest struct {
	to   *Node
	From *Node
}

func SendJoin(to *Node, from *Node) *JoinRequest {
	return &JoinRequest{
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

func SendForwardJoin(to *Node, from *Node, join *Node, ttl int) *ForwardJoinRequest {
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
	Join     bool
}

func SendNeighbor(to *Node, from *Node, priority bool) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		From:     from,
		Priority: priority,
	}
}

func SendNeighborJoin(to *Node, from *Node) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		From:     from,
		Priority: HighPriority,
		Join:     true,
	}
}

type NeighborRefuse struct {
	to   *Node
	From *Node
}

func SendNeighborRefuse(to *Node, from *Node) *NeighborRefuse {
	return &NeighborRefuse{
		to:   to,
		From: from,
	}
}

type ShuffleRequest struct {
	to      *Node
	From    *Node
	Origin  *Node
	Active  []*Node
	Passive []*Node
	TTL     int
}

func SendShuffle(to, from, origin *Node, active, passive []*Node, ttl int) *ShuffleRequest {
	return &ShuffleRequest{
		to:      to,
		From:    from,
		Origin:  from,
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

func SendShuffleReply(to *Node, from *Node, passive []*Node) *ShuffleReply {
	return &ShuffleReply{
		to:      to,
		From:    from,
		Passive: passive,
	}
}

type Gossip struct {
	to      *Node
	From    *Node
	Payload int
	Hops    int
}