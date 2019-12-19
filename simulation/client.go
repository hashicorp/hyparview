package simulation

import h "github.com/hashicorp/hyparview"

type Client struct {
	h.Hyparview
	world    *World
	app      int // final value we got
	appHot   int // gossip hotness
	appSeen  int // if app == appSeen, we got every message
	appWaste int // count of app messages that didn't change the value
	in       []h.Message
	out      []h.Message
}

func makeClient(w *World, id string) *Client {
	v := h.CreateView(&h.Node{ID: id, Addr: ""}, 0)
	c := &Client{
		Hyparview: *v,
		world:     w,
		in:        make([]h.Message, 0),
		out:       make([]h.Message, 0),
	}
	return c
}

func (c *Client) failActive(peer *Client) (ns []h.Message) {
	for _, n := range c.Passive.Shuffled() {
		if c.Active.IsEmpty() {
			// High priority can't be rejected, so send async
			m := h.SendNeighbor(n, c.Self, h.HighPriority)
			ns = append(ns, m)
			break
		} else {
			m := h.SendNeighbor(n, c.Self, h.LowPriority)
			// simulate sync network call
			peer := c.world.get(n.ID)
			ms := peer.RecvNeighbor(m)
			// any low priority response is failure
			if len(ms) == 0 {
				c.DelPassive(n)
				ns = append(ns, c.AddActive(n)...)
				break
			}
			c.DelPassive(n)
		}
	}
	return ns
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) syncGossip(payload int) (hot bool, ms []h.Message) {
	if c.app >= payload {
		c.appWaste += 1
		return false, ms
	}
	c.app = payload
	c.appSeen += 1
	c.appHot = c.world.config.gossipHeat
	for c.appHot > 0 {
		if h.Rint(c.world.config.gossipHeat) > c.appHot {
			continue
		}
		if shouldFail(c.world.config.fail.gossip) {
			continue
		}

		peer := c.world.get(c.Peer().ID)
		if shouldFail(c.world.config.fail.active) {
			c.failActive(peer)
			continue
		}

		hot, ps := peer.syncGossip(payload)
		ms = append(ms, ps...)
		if !hot || shouldFail(c.world.config.fail.gossipReply) {
			c.appHot -= 1
		}
	}
	return true, ms
}
