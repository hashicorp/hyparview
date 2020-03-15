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

func makeClient(w *World, id string) *Client {
	n := &h.Node{ID: id, Addr: id}
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
		c.w.totalMessages += 1
		c.history = append(c.history, m)
		return c.Recv(m)
	}
}

// Implement the sender interface
func (c *Client) Send(m h.Message) (*h.NeighborRefuse, error) {
	if h.Rint(100) < c.w.config.failureRate {
		return nil, fmt.Errorf("random error")
	}

	c.history = append(c.history, m)
	peer := c.w.get(m.To().ID)
	o := peer.recv(m)
	if o != nil {
		c.history = append(c.history, o)
	}

	return o, nil
}

func (c *Client) Failed(peer *h.Node) {
}

func (c *Client) Bootstrap() *h.Node {
	c.bootstrapCount += 1
	c.SendJoin(c.w.bootstrap)
	return c.w.bootstrap
}
