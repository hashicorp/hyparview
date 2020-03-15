package simulation

import (
	"fmt"
	"os"

	h "github.com/hashicorp/hyparview"
	"github.com/kr/pretty"
)

func (w *World) plotPath(file string) string {
	return fmt.Sprintf("../data/%04d-%s", w.config.iteration, file)
}

func (w *World) Connected() error {
	lost := make(map[string]*Client, len(w.nodes))
	for k, v := range w.nodes {
		lost[k] = v
	}

	var lp func(*h.Node)
	lp = func(n *h.Node) {
		if _, ok := lost[n.ID]; !ok {
			return
		}

		delete(lost, n.ID)
		for _, m := range w.get(n.ID).Active.Shuffled() {
			lp(m)
		}
	}

	// I hate that this is lp(first(nodes))
	var start *h.Node
	for _, v := range w.nodes {
		start = v.Self
		break
	}
	lp(start)

	if len(lost) == 0 {
		return nil
	}

	for _, n := range lost {
		pretty.Log(n.Self, n.history)
		break
	}

	return fmt.Errorf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
}

func (w *World) isSymmetric() error {
	count := 0
	for _, n := range w.nodes {
		for _, p := range n.Active.Shuffled() {
			if !w.get(p.ID).Active.Contains(n.Self) {
				count++
				break
			}
		}
	}

	if count == 0 {
		return nil
	}
	return fmt.Errorf("asymmetric: %d", count)
}

func (w *World) plotSeed(seed int64) {
	f, _ := os.Create(w.plotPath("seed"))
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", seed))
}

func (w *World) plotOutDegree() {
	plot := func(ns func(*h.Hyparview) int, path string) {
		act := map[string]int{}
		for _, n := range w.nodes {
			act[n.Self.ID] = ns(&n.Hyparview)
		}

		max := 0
		for _, outDegree := range act {
			if outDegree > max {
				max = outDegree
			}
		}

		deg := make([]int, max+1)
		for _, outDegree := range act {
			deg[outDegree] += 1
		}

		f, _ := os.Create(path)
		defer f.Close()
		for outDegree, peers := range deg {
			if peers == 0 {
				continue
			}
			f.WriteString(fmt.Sprintf("%d %d\n", outDegree, peers))
		}
	}

	af := w.plotPath("out-active")
	pf := w.plotPath("out-passive")
	plot(func(v *h.Hyparview) int { return v.Active.Size() }, af)
	plot(func(v *h.Hyparview) int { return v.Passive.Size() }, pf)
}

func (w *World) plotInDegree() {
	plot := func(ns func(*h.Hyparview) []*h.Node, path string) {
		act := map[string]int{}
		for _, v := range w.nodes {
			for _, n := range ns(&v.Hyparview) {
				// Count in-degree
				act[n.ID] += 1
			}
		}

		max := 0
		for _, inDegree := range act {
			if inDegree > max {
				max = inDegree
			}
		}

		deg := make([]int, max+1)
		for _, inDegree := range act {
			deg[inDegree] += 1
		}

		f, _ := os.Create(path)
		defer f.Close()
		for inDegree, peers := range deg {
			if peers == 0 {
				continue
			}
			f.WriteString(fmt.Sprintf("%d %d\n", inDegree, peers))
		}
	}
	af := w.plotPath("in-active")
	pf := w.plotPath("in-passive")
	plot(func(v *h.Hyparview) []*h.Node { return v.Active.Nodes }, af)
	plot(func(v *h.Hyparview) []*h.Node { return v.Passive.Nodes }, pf)
}

type gossipRound struct {
	miss  int
	seen  int
	waste int
	maint int
}

// Accumulate data about one round of gossip
func (w *World) traceRound(app int) {
	tot := w.gossipTotal
	if tot == nil {
		tot = &gossipRound{maint: w.totalMessages}
	}

	miss, seen, waste := 0, 0, 0
	for _, c := range w.nodes {
		if c.app < app {
			miss += 1
		}
		seen += c.appSeen
		waste += c.appWaste
	}

	rnd := &gossipRound{
		miss:  miss,
		seen:  seen - tot.seen,
		waste: waste - tot.waste,
		maint: w.totalMessages - tot.maint,
	}
	tot.miss = rnd.miss
	tot.seen += rnd.seen
	tot.waste += rnd.waste
	tot.maint += rnd.maint
	w.gossipTotal = tot
	w.gossipRound = append(w.gossipRound, rnd)
}

func (w *World) plotGossip() {
	f, _ := os.Create(w.plotPath("gossip"))
	defer f.Close()

	f.WriteString(fmt.Sprintf("%s %s %s %s\n", "Round", "Gossip", "Waste", "Hyparview"))
	for i, r := range w.gossipRound {
		f.WriteString(fmt.Sprintf("%d %d %d %d\n", i+1, r.seen, r.waste, r.maint))
	}
}
