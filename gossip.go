package hyparview

func (v *Hyparview) Gossip(m Message) {
	v.repairAsymmetry(m)
	for _, n := range v.Active.Nodes {
		if EqualNode(n, m.From()) {
			continue
		}

		v.Send(m.AssocTo(n))
	}
}
