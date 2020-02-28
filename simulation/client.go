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
	out      []h.Message
}

func makeClient(w *World, id string) *Client {
	v := h.CreateView(&h.Node{ID: id, Addr: id}, 0)
	c := &Client{
		Hyparview: *v,
		world:     w,
		out:       make([]h.Message, 0),
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

func (c *Client) failActive(peer *Client) (ms []h.Message) {
	if peer != nil {
		c.Active.DelNode(peer.Self)
	}

	for _, n := range c.Passive.Shuffled() {
		if c.Active.IsEmpty() {
			// High priority can't be rejected, so send async
			// FIXME: this send may fail, we want to add to active only if it
			// succeeds
			ms = append(ms, h.SendNeighbor(n, c.Self, h.HighPriority))
			break
		} else {
			m := h.SendNeighbor(n, c.Self, h.LowPriority)
			// simulate sync network call
			peer := c.world.get(n.ID)
			refuse := peer.RecvNeighbor(m)
			// any low priority response is failure
			if len(refuse) == 0 {
				c.DelPassive(n)
				ms = append(ms, c.AddActive(n)...)
				break
			}
			c.DelPassive(n)
		}
	}

	return ms
}

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

func (c *Client) gossip(i int) {
	c.recvGossip(&gossip{to: c.Self, app: i, hops: 0})
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

	for _, peer := range c.Active.Nodes {
		if peer.Equal(m.from) {
			continue
		}

		if c.world.shouldFail() {
			ms = append(ms, c.failActive(peer)...)
		}

		ms = append(ms, &gossip{from: c.Self, app: m.app, hops: m.hops + 1})
	}

	return ms
}
