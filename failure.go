package hyparview

type SendCallback func(Message, error)

type Send interface {
	// Send sends one message at a time to a peer. TODO simplify batching?
	// send should use a timeout to detect blocking as failure
	Send(Message, SendCallback)
	// Failed is called after hyparview has handled the failure, to handle e.g.
	// connection cleanup
	Failed(Node)
	// Bootstrap sends a join to some server, discovered externally
	Bootstrap()
}

// Send wraps the S.Send sender in appropriate error handling
// Send -> network -> callback -> recover, resend
func (v *Hyparview) Send(ms ...Message) {
	for _, m := range ms {
		// RecvDisconnect -> PromotePassiveBut may send a nil message
		// shouldn't happen otherwise, but we don't want to panic
		if m == nil {
			continue
		}

		n := m.To()

		v.S.Send(m, func(_ Message, err error) {
			if err == nil {
				return
			}

			v.Active.DelNode(n)
			v.S.Failed(n)
			v.PromotePassive(m)
		})
	}
}

func (v *Hyparview) Bootstrap() {
	v.S.Bootstrap()
}

func (v *Hyparview) PromotePassive(m Message) {
	v.PromotePassiveBut(nil, m)
}

// PromotePassiveBut chooses any passive node except peer, and sends it a neighbor message.
// If the request is accepted, the peer is promoted and message is retried with the new
// active peer.
func (v *Hyparview) PromotePassiveBut(peer Node, message Message) {
	n := v.Passive.RandNodeBut(peer)
	if n == nil {
		// We'll drop the message here, but we've got bigger problems
		v.Bootstrap()
		return
	}

	pri := v.Active.IsEmpty()
	m := NewNeighbor(n, v.Self, pri)

	v.S.Send(m, func(resp Message, err error) {
		if err != nil {
			v.Passive.DelNode(n)
			return
		}

		// Low priority, a refuse means we move on and leave the node in our passive view
		// A correct peer should never return a refusal if we sent a high priority
		// request, so checking pri here is probably unecessary
		if resp != nil && pri == LowPriority {
			v.PromotePassiveBut(peer, message)
			return
		}

		v.AddActive(n)
		v.DelPassive(n)
		v.Send(message)
		return
	})
}

// greedyShuffle tries to populate our active view on RecvShuffle
// func (v *Hyparview) greedyShuffle() {
// 	if !v.Active.IsFull() {
// 		v.PromotePassive()
// 	}
// }

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
