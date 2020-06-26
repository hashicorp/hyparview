package simulation

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	h "github.com/hashicorp/hyparview"
)

func (w *World) dir() string {
	return fmt.Sprintf("../data/%05d", w.config.iteration)
}

func (w *World) mkdir() {
	os.MkdirAll(w.dir(), 0777)
}

func (w *World) plotPath(file string) string {
	return w.dir() + "/" + file
}

func (w *World) Connected() error {
	// populate the lost list
	lost := make(map[string]*Client, len(w.nodes))
	nodes := w.nodeKeys()
	for _, n := range nodes {
		lost[n] = w.get(n)
	}

	// recursively remove the active view
	// declare first to allow recursion
	var lp func(h.Node)
	lp = func(n h.Node) {
		if _, ok := lost[n.Addr()]; !ok {
			return
		}

		delete(lost, n.Addr())
		for _, m := range w.get(n.Addr()).Active.Nodes {
			lp(m)
		}
	}

	lp(w.get(nodes[0]).Self)

	if len(lost) == 0 {
		return nil
	}

	// Log the history of lost nodes
	for n := range lost {
		w.plotPeer(n)
	}

	return fmt.Errorf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
}

func (w *World) plotPeer(peer string) {
	// View
	client := w.get(peer)
	if client == nil {
		return
	}

	f, _ := os.Create(w.plotPath("log-peer-" + peer))
	defer f.Close()

	fmt.Fprintf(f, "self:       %s\n", client.Self.Addr())
	fmt.Fprintf(f, "active:     %s\n", strings.Join(nodeAddr(client.Active.Nodes), " "))
	fmt.Fprintf(f, "passive:    %s\n", strings.Join(nodeAddr(client.Passive.Nodes), " "))

	// find and print in-degree
	in := []string{}
	for _, p := range w.randNodes() {
		if p.Active.Contains(client.Self) {
			in = append(in, p.Self.Addr())
		}
	}
	fmt.Fprintf(f, "in active:  %s\n", strings.Join(in, " "))

	in = []string{}
	for _, p := range w.randNodes() {
		if p.Passive.Contains(client.Self) {
			in = append(in, p.Self.Addr())
		}
	}
	fmt.Fprintf(f, "in passive: %s\n", strings.Join(in, " "))

	// History
	wr := bufio.NewWriter(f)
	head := "%s\t%s\t%s\t%s\t%s\n"
	row := "%s\t%s\t%s\t%s\t%d\n"
	fmt.Fprintf(wr, head, "node", "type", "io", "peer", "datum")
	for _, m := range client.history {
		io := "o"
		peer := m.To().Addr()
		if client.Self == m.To() {
			io = "i"
			peer = m.From().Addr()
		}

		fmt.Fprintf(wr, row, client.Self.Addr(), m.Type(), io, peer, datum(m))
	}
	wr.Flush()
}

func datum(m h.Message) int {
	switch v := m.(type) {
	case *gossip:
		return v.app
	case *h.ForwardJoinRequest:
		return v.TTL
	case *h.ShuffleRequest:
		return v.TTL
	case *h.NeighborRequest:
		pri := 0
		if v.Priority {
			pri = 1
		}
		return pri
	default:
		return 0
	}
}

func (w *World) isSymmetric() error {
	count := 0
	for _, n := range w.randNodes() {
		for _, p := range n.Active.Nodes {
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
	max := 0
	h := map[int]int{}
	for _, n := range w.randNodes() {
		h[n.bootstrapCount] += 1
		if n.bootstrapCount > max {
			max = n.bootstrapCount
		}
	}

	f, _ := os.Create(w.plotPath("bootstrap"))
	defer f.Close()

	// go in order to avoid map range
	for i := 0; i <= max; i++ {
		if c, ok := h[i]; ok {
			f.WriteString(fmt.Sprintf("%d %d\n", i, c))
		}
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
		for _, n := range w.randNodes() {
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

	for _, v := range w.randNodes() {
		from := v.Self.Addr()
		for _, n := range part(v).Nodes {
			f.WriteString(fmt.Sprintf(row, from, n.Addr()))
		}
	}
}

func (w *World) plotInDegree() {
	plot := func(part getPart, path string) {
		act := map[string]int{}
		for _, v := range w.randNodes() {
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
	for _, c := range w.randNodes() {
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
