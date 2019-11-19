package hyparview

type Client struct {
	v: *Hyparview,
	in: []Message,
	out []Message,
}

func (c *Client) failActive(failed *Node) {
	var node *Node
	for _, n := range c.v.Passive.Shuffled() {
		err := c.Dial(p)
		if err == nil {
			// if too few send hi neighbor
			// else send lo neighbor, maybe error
			node = p
		}
	}
}
