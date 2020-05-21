package simulation

import (
	"fmt"

	h "github.com/hashicorp/hyparview"
)

type Client struct {
	h.Hyparview
	w              *World
	history        []h.Message // debug history of messages
	bootstrapCount int
	app            int // final value we got
	appHops        int // final value's hops
	appSeen        int // if app == appSeen, we got every message
	appWaste       int // count of app messages that didn't change the value
}

func makeClient(w *World, addr string) *Client {
	n := h.NewNode(addr)
	v := h.CreateView(nil, n, 0)
	c := &Client{
		Hyparview: *v,
		w:         w,
	}
	c.S = c
	return c
}

func (c *Client) recv(m h.Message) *h.NeighborRefuse {
	switch m1 := m.(type) {
	case *gossip:
		c.recvGossip(m1)
		return nil
	default:
		c.history = append(c.history, m)
		return c.Recv(m)
	}
}

func (c *Client) shouldFail() bool {
	// return h.Rint(100) < c.w.config.failureRate

	// Retry
	for i := 2; i > 0; i-- {
		if h.Rint(100) > c.w.config.failureRate {
			return false
		}
	}
	return true
}

// Implement the sender interface
func (c *Client) Send(m h.Message) (*h.NeighborRefuse, error) {
	c.w.totalMessages += 1

	if c.shouldFail() {
		return nil, fmt.Errorf("request error")
	}

	c.history = append(c.history, m)
	peer := c.w.get(m.To().Addr())
	o := peer.recv(m)
	if o != nil {
		c.history = append(c.history, o)
		if c.shouldFail() {
			return nil, fmt.Errorf("response error")
		}
	}

	return o, nil
}

func (c *Client) Failed(peer h.Node) {
}

func (c *Client) Bootstrap() h.Node {
	c.bootstrapCount += 1
	c.SendJoin(c.w.bootstrap)
	return c.w.bootstrap
}
