package simulation

import "github.com/hashicorp/hyparview"

type ClientConfig struct {
	MaxHot int
	fail   WorldFailureRate
}

type Client struct {
	Hyparview
	clientConfig
	app      int // final value we got
	appHot   int // gossip hotness
	appSeen  int // if app == appSeen, we got every message
	appWaste int // count of app messages that didn't change the value
	in       []Message
	out      []Message
}

func create(id string, cfg clientConfig) *Client {
	v := CreateView(&Node{ID: id, Addr: ""}, 0)
	c := &Client{
		Hyparview:    *v,
		clientConfig: cfg,
		in:           make([]Message, 0),
		out:          make([]Message, 0),
	}
	return c
}

func (c *Client) failActive(peer *Client) (ns []*Message) {
	for _, n := range c.Passive.Shuffled() {
		if c.Active.IsEmpty() {
			// High priority can't be rejected, so send async
			m := hyparview.SendNeighbor(n, c.Self, HighPriority)
			ns = append(ns, m)
			break
		} else {
			m := hyparview.SendNeighbor(n, c.Self, LowPriority)
			// simulate sync network call
			ms := n.RecvNeighbor(m)
			// any low priority response is failure
			if len(ms) == 0 {
				c.DelPassive(n)
				ns = append(ns, c.AddActive(n))
				break
			}
			c.DelPassive(n)
		}
	}
	return ns
}

// Example gossip implementation. For deterministic testing, each payload runs until the
// message is completely distributed.
func (c *Client) syncGossip(payload int) (ms []*Message) {
	if c.app >= payload {
		return
	}
	c.app = payload
	c.appHot = c.MaxHot
	for c.appHot > 0 {
		if hyparview.rint(c.MaxHot) > c.appHot {
			continue
		}
		if shouldFail(c.fail.gossip) {
			continue
		}

		peer := c.Peer()
		if shouldFail(c.fail.active) {

		}

		if !peer.recv(payload) || shouldFail(c.fail.gossipReply) {
			c.appHot -= 1
		}
	}
}

func (c *Client) recv(payload int) bool {
	if payload < c.app {
		c.appWaste += 1
		return false
	}
	c.app = payload
	c.appSeen += 1
	return true
}
