package hyparview

type ViewPart struct {
	nodes []Node
	max   int
}

func CreateViewPart(size int) *ViewPart {
	return &ViewPart{
		nodes: make([]Node, 0, size),
		max:   size,
	}
}

func (v *ViewPart) IsEmpty() bool {
	return len(v.nodes) <= 1
}

func (v *ViewPart) IsEmptyBut(peer Node) bool {
	return v.IsEmpty() ||
		(len(v.nodes) == 1 &&
			EqualNode(peer, v.nodes[0]))
}

func (v *ViewPart) IsFull() bool {
	return len(v.nodes) >= v.max
}

func (v *ViewPart) Size() int {
	return len(v.nodes)
}

func (v *ViewPart) Copy() *ViewPart {
	w := *v
	nodes := make([]Node, v.max)
	copy(nodes, v.nodes)
	w.nodes = nodes
	return &w
}

func (v *ViewPart) Equal(w *ViewPart) bool {
	if w == nil {
		return v == w
	}
	if v.Size() != w.Size() {
		return false
	}
setwise:
	for _, n := range v.nodes {
		for _, m := range w.nodes {
			if EqualNode(n, m) {
				continue setwise
			}
		}
		return false
	}
	return true
}

func (v *ViewPart) Add(n Node) {
	if !v.Contains(n) {
		v.nodes = append(v.nodes, n)
	}
}

func (v *ViewPart) DelIndex(i int) {
	v.nodes = append(v.nodes[:i], v.nodes[i+1:]...)
}

func (v *ViewPart) DelNode(n Node) bool {
	idx := v.ContainsIndex(n)
	if idx >= 0 {
		v.DelIndex(idx)
		return true
	}
	return false
}

func (v *ViewPart) GetIndex(i int) Node {
	return v.nodes[i]
}

func (v *ViewPart) Shuffled() []Node {
	l := len(v.nodes)
	ns := make([]Node, l)
	// Start with a copy, fischer-yates needs to operate destructively
	copy(ns, v.nodes)
	for i := l - 1; i > 0; i-- {
		j := Rint(i)
		ns[i], ns[j] = ns[j], ns[i]
	}
	return ns
}

func (v *ViewPart) RandIndex() int {
	return Rint(len(v.nodes) - 1)
}

func (v *ViewPart) RandNode() Node {
	return v.nodes[v.RandIndex()]
}

func (v *ViewPart) ContainsIndex(n Node) int {
	for i, m := range v.nodes {
		if EqualNode(n, m) {
			return i
		}
	}
	return -1
}

func (v *ViewPart) Contains(n Node) bool {
	return v.ContainsIndex(n) >= 0
}
