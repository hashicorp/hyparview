package hyparview

// Message allows clients to redefine hyparview messages to carry additional meta information
type Message interface {
	To() *Node
	AssocTo(*Node) Message
	From() *Node
}

// Methods that can be generated should be added to message.go.genny, and build by `make
// test`. The genny generator does some funny things around interfaces and type receivers,
// so the template file isn't included in the build for now.

const (
	HighPriority = true
	LowPriority  = false
)

type JoinRequest struct {
	to   *Node
	from *Node
}

func NewJoin(to *Node, from *Node) *JoinRequest {
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

func NewForwardJoin(to *Node, from *Node, join *Node, ttl int) *ForwardJoinRequest {
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

func NewDisconnect(to *Node, from *Node) *DisconnectRequest {
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

func NewNeighbor(to *Node, from *Node, priority bool) *NeighborRequest {
	return &NeighborRequest{
		to:       to,
		from:     from,
		Priority: priority,
	}
}

func NewNeighborJoin(to *Node, from *Node) *NeighborRequest {
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

func NewNeighborRefuse(to *Node, from *Node) *NeighborRefuse {
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

func NewShuffle(to, from, origin *Node, active, passive []*Node, ttl int) *ShuffleRequest {
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

func NewShuffleReply(to *Node, from *Node, passive []*Node) *ShuffleReply {
	return &ShuffleReply{
		to:      to,
		from:    from,
		Passive: passive,
	}
}
