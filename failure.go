package hyparview

type Send interface {
	// Send sends one message at a time to a peer. TODO simplify batching?
	// send should use a timeout to detect blocking as failure
	Send(Message) (*NeighborRefuse, error)
	// Failed is called after hyparview has handled the failure, to handle e.g.
	// connection cleanup
	Failed(Node)
	// Bootstrap sends a join to some server, discovered by some external consideration
	Bootstrap() Node
}

// Send wraps the S.Send sender in appropriate error handling
func (v *Hyparview) Send(ms ...Message) {
	subs := map[string]Node{}

	for i := 0; i < len(ms); {
		m := ms[i]
		n := m.To()

		// Send messages to the replacement node, if we have one
		if sub, ok := subs[n.Addr()]; ok {
			m = m.AssocTo(sub)
		}

		_, err := v.S.Send(m)
		if err != nil {
			v.Active.DelNode(n)
			v.S.Failed(n)
			sub := v.PromotePassive()

			if sub == nil {
				// FIXME re-Join
				// log.Printf("WARN empty passive view, fail %d", len(ms)-i)
				// return
				sub = v.S.Bootstrap()
			}

			subs[n.Addr()] = sub
		} else {
			// On failure, retry the failed message with the replacement server
			i++
		}
	}
}

func (v *Hyparview) Bootstrap() Node {
	return v.S.Bootstrap()
}

func (v *Hyparview) PromotePassive() Node {
	return v.PromotePassiveBut(nil)
}

func (v *Hyparview) PromotePassiveBut(peer Node) Node {
	pri := v.Active.IsEmpty()

	for _, n := range v.Passive.Shuffled() {
		if EqualNode(n, peer) {
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

// greedyShuffle tries to populate our active view on RecvShuffle
func (v *Hyparview) greedyShuffle() {
	if !v.Active.IsFull() {
		v.PromotePassive()
	}
}

// repairAsymmetry handles a message from an unexpected sender
func (v *Hyparview) repairAsymmetry(m Message) {
	peer := m.From()
	if EqualNode(v.Self, peer) || v.Active.Contains(peer) {
		return
	}
	if v.Active.IsFull() {
		v.Send(NewDisconnect(peer, v.Self))
		return
	}
	v.Active.Add(peer)
}

// SendKeepalives actively repairs the active view
func (v *Hyparview) SendKeepalives() {
	for _, n := range v.Active.Nodes {
		v.Send(NewNeighborKeepalive(n, v.Self))
	}
}
