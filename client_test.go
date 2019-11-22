package hyparview

type Client struct {
	v   *Hyparview
	in  []Message
	out []Message
}

func (c *Client) create(id string, addr string, n *Node) *Hyparview {
	v := CreateView(&Node{ID: id, Addr: addr}, 0)
	ms := n.v.RecvJoin(SendJoin(n, v.Self))
	c.out = append(c.out, ms...)
	return v
}

func (c *Client) failActive(failed *Node) {
	var node *Node
	for _, n := range c.v.Passive.Shuffled() {
		if v.Active.IsEmpty() {
			// simulate sync network call
			// TODO simulate failure
			m := SendNeighbor(n, c.v.Self, HighPriority)
			ms := n.v.RecvNeighbor(m)
			// forward maybe disconnect messages
			c.out = append(c.out, ms)
			break
		} else {
			m := SendNeighbor(n, c.v.Self, LowPriority)
			ms := n.v.RecvNeighbor(m)
			// any low priority response is failure
			if len(ms) == 0 {
				c.v.DelPassive(n)
				c.v.AddActive(n)
				break
			}
		}
	}
}
