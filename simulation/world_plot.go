package simulation

func (w *World) isConnected() bool {
	lost := make(map[string]*Client, len(w.nodes))
	for k, v := range w.nodes {
		lost[k] = v
	}

	var lp func(*Node)
	lp = func(n *Node) {
		if _, ok := lost[n.ID]; !ok {
			return
		}

		delete(lost, n.ID)
		for _, m := range w.get(n.ID).Active.Shuffled() {
			lp(m)
		}
	}

	// I hate that this is lp(first(nodes))
	var start *Node
	for _, v := range w.nodes {
		start = v.Self
		break
	}
	lp(start)

	fmt.Printf("%d connected, %d lost\n", len(w.nodes)-len(lost), len(lost))
	return len(lost) == 0
}

func (w *World) PlotInDegree() {
	plot := func(ns func(*Hyparview) []*Node, path string) {
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

	plot(func(v *Hyparview) []*Node { return v.Active.Nodes }, "active.data")
	plot(func(v *Hyparview) []*Node { return v.Passive.Nodes }, "passive.data")
}
