package simulation

import (
	h "github.com/hashicorp/hyparview"
)

type Client struct {
	h.Hyparview
	w        *World
	app      int // final value we got
	appHops  int // final value's hops
	appSeen  int // if app == appSeen, we got every message
	appWaste int // count of app messages that didn't change the value
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

// Implement the sender interface
func (c *Client) Send(m h.Message) (h.Message, error) {
	c.w.totalMessages += 1
	peer := c.w.get(m.To().ID)
	resp := peer.Recv(m)
	return resp, nil
}

func (c *Client) Failed(peer *h.Node) {
}
