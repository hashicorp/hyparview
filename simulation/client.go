package simulation

import (
	"log"

	h "github.com/hashicorp/hyparview"
)

type Client struct {
	h.Hyparview
	world    *World
	app      int // final value we got
	appHops  int // final value's hops
	appHot   int // gossip hotness
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
				log.Printf("faila %s => %s", c.Self.Addr, peer.Self.Addr)
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
	app  int
	hops int
}

func (g *gossip) To() *h.Node {
	return g.to
}

func (c *Client) gossip(i int) {
	c.recvGossip(&gossip{to: c.Self, app: i, hops: 0})
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) recvGossip(m *gossip) bool {
	if c.app >= m.app {
		c.appWaste += 1
		return false
	}
	c.app = m.app
	c.appHops = m.hops
	c.appSeen += 1
	c.appHot = c.world.config.gossipHeat

	// Count hops between peers
	m.hops += 1

	for c.appHot > 0 {
		if h.Rint(c.world.config.gossipHeat) > c.appHot {
			continue
		}
		if c.world.shouldFail() {
			continue
		}

		node := c.Peer()
		if node == nil {
			// We're disconnected and can't make forward progress
			if c.Passive.IsEmpty() {
				return true
			}

			c.world.send(c.failActive(nil)...)
			continue
		}

		peer := c.world.get(node.ID)
		if c.world.shouldFail() {
			c.world.send(c.failActive(peer)...)
			continue
		}

		hot := peer.recvGossip(m)
		if !hot || c.world.shouldFail() {
			c.appHot -= 1
		}
	}
	return true
}
