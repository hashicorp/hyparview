package simulation

import (
	"fmt"
	"log"
	"strings"
)

// symCheck is an adhoc debugging tool
func (w *World) symCheck(m h.Message) {
	// if w.spinCountM == nil {
	// 	w.spinCountM = map[string]int{}
	// }

	// n := w.get(m.FromNode().ID)
	// p := w.get(m.To().ID)
	// if n.Active.Contains(p.Self) != p.Active.Contains(n.Self) {
	// 	pretty.Log("asymmetric", m)
	// }
	// return

	switch m1 := m.(type) {
	// case *h.JoinRequest:
	// 	fmt.Printf("%s %s\n", m1.To().ID, m1.From.ID)

	case *h.ForwardJoinRequest:
		if w.spinCount >= 1000000 {
			w.spinCount = 0
			log.Printf("fwd  1m %s %d", m1.Join.ID, m1.TTL)
		} else {
			w.spinCount += 1
		}
		if m1.From.ID == m1.To().ID {
			log.Printf("fwd  dup")
		}
		if m1.TTL < 0 {
			log.Printf("fwd  ttl0")
		}

	case *h.DisconnectRequest:
		if m1.From.ID == m1.To().ID {
			log.Printf("diss dup")
		}
		n := w.get(m1.From.ID)
		m := w.get(m.To().ID)
		if n.Active.Contains(m.Self) {
			log.Printf("diss %s %s", m1.From.ID, m1.To().ID)
		}

		if m.Active.Contains(n.Self) {
			log.Printf("disr %s %s", m1.From.ID, m1.To().ID)
		}

	case *h.NeighborRequest:
		if !m1.Join || m1.From.ID == m1.To().ID {
			return
		}
		n := w.get(m1.From.ID)
		m := w.get(m.To().ID)
		if !(n.Active.Contains(m.Self) && m.Active.Contains(n.Self)) {
			log.Printf("nei %s %s", m1.From.ID, m1.To().ID)
		}
	default:
	}

	// w.spinCountM[fmt.Sprintf("%T", m)] += 1
	// w.spinPrint()
}

func (w *World) spinPrint() {
	if h.Rint(100000) == 1 {
		var ss []string
		for k, v := range w.spinCountM {
			ss = append(ss, fmt.Sprintf("%s:%d", k, v))
		}
		log.Println(strings.Join(ss, " "))
	}
}
