package hyparview

func (v *Hyparview) Gossip(m Message) {
	for _, n := range v.Active.Nodes {
		if n.Equal(m.From()) {
			continue
		}

		v.Send(m.AssocTo(n))
	}
}
