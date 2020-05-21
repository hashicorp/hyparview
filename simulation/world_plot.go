package simulation

import (
	"bufio"
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

	var lp func(h.Node)
	lp = func(n h.Node) {
		if _, ok := lost[n.Addr()]; !ok {
			return
		}

		delete(lost, n.Addr())
		for _, m := range w.get(n.Addr()).Active.Shuffled() {
			lp(m)
		}
	}

	// I hate that this is lp(first(nodes))
	var start h.Node
	for _, v := range w.nodes {
		start = v.Self
		break
	}
	lp(start)

	if len(lost) == 0 {
		return nil
	}

	// Log the history of lost nodes
	f, _ := os.Create(w.plotPath("lost.log"))
	defer f.Close()
	wr := bufio.NewWriter(f)
	for _, n := range lost {
		pretty.Fprintf(wr, "%# v\n%# v\n", n.Self, n.history)
		break
	}

	return fmt.Errorf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
}

func (w *World) isSymmetric() error {
	count := 0
	for _, n := range w.nodes {
		for _, p := range n.Active.Shuffled() {
			if !w.get(p.Addr()).Active.Contains(n.Self) {
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

func (w *World) plotBootstrapCount() {
	h := map[int]int{}
	for _, n := range w.nodes {
		h[n.bootstrapCount] += 1
	}

	f, _ := os.Create(w.plotPath("bootstrap"))
	defer f.Close()

	for boots, nodes := range h {
		f.WriteString(fmt.Sprintf("%d %d\n", boots, nodes))
	}
}

type getPart func(c *Client) *h.ViewPart

func activePart(c *Client) *h.ViewPart {
	return c.Active
}

func passivePart(c *Client) *h.ViewPart {
	return c.Passive
}

func (w *World) plotOutDegree() {
	plot := func(p getPart, path string) {
		act := map[string]int{}
		for _, n := range w.nodes {
			act[n.Self.Addr()] = p(n).Size()
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

	plot(activePart, w.plotPath("out-active"))
	plot(passivePart, w.plotPath("out-passive"))
}

func (w *World) plotGraphs() {
	w.plotGraph("graph-active", activePart)
	w.plotGraph("graph-passive", passivePart)
}

func (w *World) plotGraph(plotName string, part getPart) {
	path := w.plotPath(plotName)
	f, _ := os.Create(path)
	defer f.Close()

	row := "%s\t%s\n"

	for _, v := range w.nodes {
		from := v.Self.Addr()
		for _, n := range part(v).Nodes {
			f.WriteString(fmt.Sprintf(row, from, n.Addr()))
		}
	}
}

func (w *World) plotInDegree() {
	plot := func(part getPart, path string) {
		act := map[string]int{}
		for _, v := range w.nodes {
			for _, n := range part(v).Nodes {
				// Count in-degree
				act[n.Addr()] += 1
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

	plot(activePart, w.plotPath("in-active"))
	plot(passivePart, w.plotPath("in-passive"))
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
