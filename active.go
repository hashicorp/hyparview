package hyparview

type ActiveView struct {
	Nodes []Node
	Max   int
}

func (v *ActiveView) IsFull() bool {
	return len(v.Nodes) >= v.Max
}

func (v *ActiveView) Add(n Node) {
	if !v.Contains(n) {
		v.Nodes = append(v.Nodes, n)
	}
}

func (v *ActiveView) DelIndex(i int) {
	ns := v.Nodes
	mx := len(ns) - 1
	v.Nodes = append(ns[0:i], ns[i+1:mx]...)
}

func (v *ActiveView) GetIndex(i int) Node {
	return v.Nodes[i]
}

func (v *ActiveView) Contains(n Node) bool {
	for _, m := range v.Nodes {
		if m.Equal(n) {
			return false
		}
	}
	return true
}
