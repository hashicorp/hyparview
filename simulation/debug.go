package simulation

import (
	"fmt"
	"log"
	"strings"

	h "github.com/hashicorp/hyparview"
)

// symCheck is an adhoc debugging tool
func (w *World) symCheck(m h.Message) {
	// if w.spinCountM == nil {
	// 	w.spinCountM = map[string]int{}
	// }

	// n := w.get(m.FromNode().Addr())
	// p := w.get(m.To().Addr())
	// if n.Active.Contains(p.Self) != p.Active.Contains(n.Self) {
	// 	pretty.Log("asymmetric", m)
	// }
	// return

	switch m1 := m.(type) {
	// case *h.JoinRequest:
	// 	fmt.Printf("%s %s\n", m1.To().Addr(), m1.From().Addr())

	case *h.ForwardJoinRequest:
		if w.spinCount >= 1000000 {
			w.spinCount = 0
			log.Printf("fwd  1m %s %d", m1.Join.Addr(), m1.TTL)
		} else {
			w.spinCount += 1
		}
		if m1.From().Addr() == m1.To().Addr() {
			log.Printf("fwd  dup")
		}
		if m1.TTL < 0 {
			log.Printf("fwd  ttl0")
		}

	case *h.DisconnectRequest:
		if m1.From().Addr() == m1.To().Addr() {
			log.Printf("diss dup")
		}
		n := w.get(m1.From().Addr())
		m := w.get(m.To().Addr())
		if n.Active.Contains(m.Self) {
			log.Printf("diss %s %s", m1.From().Addr(), m1.To().Addr())
		}

		if m.Active.Contains(n.Self) {
			log.Printf("disr %s %s", m1.From().Addr(), m1.To().Addr())
		}

	case *h.NeighborRequest:
		if !m1.Join || m1.From().Addr() == m1.To().Addr() {
			return
		}
		n := w.get(m1.From().Addr())
		m := w.get(m.To().Addr())
		if !(n.Active.Contains(m.Self) && m.Active.Contains(n.Self)) {
			log.Printf("nei %s %s", m1.From().Addr(), m1.To().Addr())
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
