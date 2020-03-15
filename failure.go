package hyparview

type Send interface {
	// Send sends one message at a time to a peer. TODO simplify batching?
	// send should use a timeout to detect blocking as failure
	Send(Message) (*NeighborRefuse, error)
	// Failed is called after hyparview has handled the failure, to handle e.g.
	// connection cleanup
	Failed(*Node)
}

// Send wraps the S.Send sender in appropriate error handling
func (v *Hyparview) Send(ms ...Message) {
	subs := map[string]*Node{}

	for i := 0; i < len(ms); {
		m := ms[i]
		n := m.To()

		// Send messages to the replacement node, if we have one
		if sub, ok := subs[n.ID]; ok {
			m = m.AssocTo(sub)
		}

		_, err := v.S.Send(m)
		if err != nil {
			v.Active.DelNode(n)
			sub := v.PromotePassive()
			if sub == nil {
				// FIXME re-Join
				// log.Printf("WARN empty passive view, fail %d", len(ms)-i)
				return
			}
			subs[n.ID] = sub
			v.S.Failed(n)
		} else {
			// On failure, retry the failed message with the replacement server
			i++
		}
	}
}

func (v *Hyparview) PromotePassive() *Node {
	return v.PromotePassiveBut(nil)
}

func (v *Hyparview) PromotePassiveBut(peer *Node) *Node {
	pri := v.Active.IsEmpty()

	for _, n := range v.Passive.Shuffled() {
		if n.Equal(peer) {
			continue
		}

		m := NewNeighbor(n, v.Self, pri)

		resp, err := v.S.Send(m)
		if err != nil {
			v.Passive.DelNode(n)
			continue
		}

		if pri == HighPriority {
			v.AddActive(n)
			v.DelPassive(n)
			return n
		}

		// Low priority, a refuse means we move on but keep the peer
		if resp != nil {
			continue
		}

		v.AddActive(n)
		v.DelPassive(n)
		return n
	}
	return nil
}
