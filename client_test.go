package hyparview

type Client struct {
	v   *Hyparview
	in  []Message
	out []Message
}

func (c *Client) failActive(failed *Node) {
	var node *Node
	for _, n := range c.v.Passive.Shuffled() {
		if v.Active.IsEmpty() {
			// simulate sync network call
			// TODO simulate failure
			ms := n.v.RecvNeighbor(HighPriority, c.v.Self)
			// forward maybe disconnect messages
			c.out = append(c.out, ms)
			break
		} else {
			ms := n.v.RecvNeighbor(LowPriority, c.v.Self)
			// any low priority response is failure
			if len(ms) == 0 {
				c.v.DelPassive(n)
				c.v.AddActive(n)
				break
			}
		}
	}
}
