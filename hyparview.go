package hyparview

type Hyparview struct {
	Active     ActiveView
	ActiveRWL  int
	Passive    PassiveView
	PassiveRWL int
	Self       Node
}

type Node string

func (v *Hyparview) ReceiveJoin(node Node) (ms []Message) {
	if v.Active.IsFull() {
		v.Active.DropRando()
	}

	v.Active.Add(node)

	for _, n := range v.Active.Nodes {
		if n.Equal(node) {
			continue
		}
		ms = append(ms, SendForwardJoin(n, node, v.ActiveRWL, v.Self))
	}

	return ms
}
