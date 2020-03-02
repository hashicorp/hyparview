package simulation

import (
	h "github.com/hashicorp/hyparview"
)

type Client struct {
	h.Hyparview
	world    *World
	app      int // final value we got
	appHops  int // final value's hops
	appSeen  int // if app == appSeen, we got every message
	appWaste int // count of app messages that didn't change the value
	s        *h.SliceSender
}

func makeClient(w *World, id string) *Client {
	s := h.newSliceSender()
	n := &Node{ID: id, Addr: id}
	v := h.CreateView(s, n, 0)
	c := &Client{
		Hyparview: *v,
		world:     w,
		s:         s,
	}
	return c
}

func (c *Client) messages() []h.Message {
	out := c.out
	c.out = make([]h.Message, 0)
	return out
}

func (c *Client) getPeer() (*h.Node, []h.Message) {
	p := c.Peer()
	if p != nil {
		return p, nil
	}
	ms := c.failActive(nil)
	return c.Peer(), ms
}

// failActive cheats the implementation: it delivers messages synchronously directly, and
// therefore avoids the awkward mismatch on returning the messages
func (c *Client) failActive(peer *Client) (ms []h.Message) {
	if peer != nil {
		c.Active.DelNode(peer.Self)
	}

	for _, n := range c.Passive.Shuffled() {
		// Failure always removes the node from our passive view
		if c.world.shouldFail() {
			c.DelPassive(n)
			continue
		}

		pri := c.Active.IsEmpty()
		m := h.SendNeighbor(n, c.Self, pri)

		// simulate sync network call
		peer := c.world.get(n.ID)
		resp := peer.RecvNeighbor(m)

		if pri == h.HighPriority {
			c.AddActive(n)
			c.DelPassive(n)
			return resp
		}

		// a refuse means we move on, but keep the peer
		if len(resp) != 0 {
			continue
		}

		// we don't need to return AddActive because we certainly have an empty slot
		c.AddActive(n)
		c.DelPassive(n)
	}

	return ms
}
