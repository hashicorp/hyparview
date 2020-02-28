package simulation

type gossip struct {
	to   *h.Node
	from *h.Node
	app  int
	hops int
}

// gossip is an h.Message
func (g *gossip) To() *h.Node {
	return g.to
}

func (c *Client) gossip(i int) (ms []h.Message) {
	return c.recvGossip(&gossip{to: c.Self, app: i, hops: 0})
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) recvGossip(m *gossip) (ms []h.Message) {
	if c.app >= m.app {
		c.appWaste += 1
		return ms
	}
	c.app = m.app
	c.appHops = m.hops
	c.appSeen += 1

	for _, n := range c.Active.Nodes {
		if n.Equal(m.from) {
			continue
		}

		peer := c.world.get(n.ID)
		if c.world.shouldFail() {
			ms = append(ms, c.failActive(peer)...)
		}

		ms = append(ms, newGossip(n, c.Self, m.app, m.hops+1))
	}

	return ms
}

func newGossip(to, from *h.Node, payload, hops int) *gossip {
	return &gossip{
		to:   to,
		from: from,
		app:  payload,
		hops: hops,
	}
}

func (c *Client) recv(m h.Message) []h.Message {
	switch m1 := m.(type) {
	case *gossip:
		return c.recvGossip(m1)
	default:
	}
	return c.Recv(m)
}
