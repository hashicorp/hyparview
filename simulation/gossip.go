package simulation

import h "github.com/hashicorp/hyparview"

type gossip struct {
	to   *h.Node
	from *h.Node
	app  int
	hops int
}

func newGossip(to, from *h.Node, payload, hops int) *gossip {
	return &gossip{
		to:   to,
		from: from,
		app:  payload,
		hops: hops,
	}
}

func (c *Client) gossip(i int) []*gossip {
	return c.recvGossip(&gossip{to: c.Self, app: i, hops: 0})
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) recvGossip(m *gossip) (ms []*gossip) {
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
		// FIXME retry with the fixed connection
		if c.world.shouldFail() {
			c.world.send(c.failActive(peer)...)
		}

		ms = append(ms, newGossip(n, c.Self, m.app, m.hops+1))
	}

	return ms
}

// Send the gossip messages and all messages caused by them. Just like world.send, except
// for accounting; we want to keep counts separately to properly manage the graphs. FIXME
// integrate this sensibly
func (w *World) sendGossip(ms ...*gossip) {
	for _, m := range ms {
		v := w.get(m.to.ID)
		if v != nil {
			w.sendGossip(v.recvGossip(m)...)
		}
	}
}
