package hyparview

type ActiveView struct {
	Nodes []*Node
	Max   int
}

func CreateActiveView(size int) *ActiveView {
	return &ActiveView{
		Nodes: []*Node{},
		Max:   size,
	}
}

func (v *ActiveView) IsEmpty() bool {
	return len(v.Nodes) == 0
}

func (v *ActiveView) IsFull() bool {
	return len(v.Nodes) >= v.Max
}

func (v *ActiveView) Size() int {
	return len(v.Nodes)
}

func (v *ActiveView) Add(n *Node) {
	if !v.Contains(n) {
		v.Nodes = append(v.Nodes, n)
	}
}

func (v *ActiveView) DelIndex(i int) {
	ns := v.Nodes
	mx := len(ns) - 1
	v.Nodes = append(ns[0:i], ns[i+1:mx]...)
}

func (v *ActiveView) GetIndex(i int) *Node {
	return v.Nodes[i]
}

func (v *ActiveView) Shuffled() []*Node {
	l := len(v.Nodes)
	ns := make([]*Node, l)
	for i := l - 1; i > 0; i-- {
		j := rint(i)
		ns[i], ns[j] = v.Nodes[j], v.Nodes[i]
	}
	return ns
}

// func (v *ActiveView) RandIndex() int {
// 	return rint(len(v.Nodes) - 1)
// }

// func (v *ActiveView) RandNode() *Node {
// 	return v.Nodes[v.RandIndex()]
// }

func (v *ActiveView) ContainsIndex(n *Node) int {
	for i, m := range v.Nodes {
		if m.Equal(n) {
			return i
		}
	}
	return -1
}

func (v *ActiveView) Contains(n *Node) bool {
	return v.ContainsIndex(n) > 0
}
