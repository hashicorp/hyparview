package simulation

import h "github.com/hashicorp/hyparview"

type gossip struct {
	to   h.Node
	from h.Node
	app  int
	hops int
}

// Implement the message interface for gossip
func (r *gossip) To() h.Node { return r.to }
func (r *gossip) AssocTo(n h.Node) h.Message {
	o := *r
	o.to = n
	return &o
}
func (r *gossip) From() h.Node { return r.from }
func (r *gossip) Type() string { return "gossip" }

func (c *Client) gossip(i int) {
	c.recvGossip(&gossip{to: c.Self, app: i, hops: 0})
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) recvGossip(m *gossip) {
	if c.app >= m.app {
		c.appWaste += 1
		return
	}
	c.app = m.app
	c.appHops = m.hops
	c.appSeen += 1

	c.Gossip(&gossip{from: c.Self, app: m.app, hops: m.hops + 1})
}
