package hyparview

type ViewPart struct {
	Nodes []*Node
	Max   int
}

func CreateViewPart(size int) *ViewPart {
	return &ViewPart{
		Nodes: []*Node{},
		Max:   size,
	}
}

func (v *ViewPart) IsEmpty() bool {
	return len(v.Nodes) == 0
}

func (v *ViewPart) IsFull() bool {
	return len(v.Nodes) >= v.Max
}

func (v *ViewPart) Size() int {
	return len(v.Nodes)
}

func (v *ViewPart) Add(n *Node) {
	if !v.Contains(n) {
		v.Nodes = append(v.Nodes, n)
	}
}

func (v *ViewPart) DelIndex(i int) {
	ns := v.Nodes
	mx := len(ns) - 1
	v.Nodes = append(ns[0:i], ns[i+1:mx]...)
}

func (v *ViewPart) GetIndex(i int) *Node {
	return v.Nodes[i]
}

func (v *ViewPart) Shuffled() []*Node {
	l := len(v.Nodes)
	ns := make([]*Node, l)
	// Start with a copy, fischer-yates needs to operate destructively
	copy(ns, v.Nodes)
	for i := l - 1; i > 0; i-- {
		j := rint(i)
		ns[i], ns[j] = ns[j], ns[i]
	}
	return ns
}

// func (v *ViewPart) RandIndex() int {
// 	return rint(len(v.Nodes) - 1)
// }

// func (v *ViewPart) RandNode() *Node {
// 	return v.Nodes[v.RandIndex()]
// }

func (v *ViewPart) ContainsIndex(n *Node) int {
	for i, m := range v.Nodes {
		if m.Equal(n) {
			return i
		}
	}
	return -1
}

func (v *ViewPart) Contains(n *Node) bool {
	return v.ContainsIndex(n) > 0
}
