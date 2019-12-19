package simulation

import (
	"fmt"
	"os"

	h "github.com/hashicorp/hyparview"
)

func (w *World) plotPath(file string) string {
	return fmt.Sprintf("../data/%04d-%s", w.iteration, file)
}

func (w *World) isConnected() bool {
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

	fmt.Printf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
	return len(lost) == 0
}

func (w *World) PlotSeed(seed int64) {
	f, _ := os.Create(w.plotPath("seed"))
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", seed))
}

func (w *World) PlotInDegree() {
	plot := func(ns func(*h.Hyparview) []*h.Node, path string) {
		act := map[string]int{}
		for _, v := range w.nodes {
			for _, n := range ns(&v.Hyparview) {
				act[n.ID] += 1
			}
		}

		max := 0
		for _, c := range act {
			if c > max {
				max = c
			}
		}

		deg := make([]int, max+1)
		for _, c := range act {
			deg[c] += 1
		}

		f, _ := os.Create(path)
		defer f.Close()
		for i, c := range deg {
			f.WriteString(fmt.Sprintf("%d %d\n", i, c))
		}
	}
	af := w.plotPath("active")
	pf := w.plotPath("passive")
	plot(func(v *h.Hyparview) []*h.Node { return v.Active.Nodes }, af)
	plot(func(v *h.Hyparview) []*h.Node { return v.Passive.Nodes }, pf)
}

type gossipRound struct {
	app   int
	seen  int
	waste int
}

// Accumulate data about one round of gossip
func (w *World) plotGossipRound() {
	tot := w.gossipTotal
	if tot == nil {
		tot = &gossipRound{}
	}

	app, seen, waste := 0, 0, 0
	for _, c := range w.nodes {
		app += c.app
		seen += c.appSeen
		waste += c.appWaste
	}
	app = app / len(w.nodes)

	rnd := &gossipRound{
		app:   app,
		seen:  seen - tot.seen,
		waste: waste - tot.waste,
	}
	tot.app = rnd.app
	tot.seen += rnd.seen
	tot.waste += rnd.waste
	w.gossipTotal = tot
	w.gossipPlot = append(w.gossipPlot, rnd)
}
