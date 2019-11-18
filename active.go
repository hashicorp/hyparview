package hyparview

type ActiveView struct {
	Nodes []Node
	MAX   int
}

func (v *ActiveView) IsFull() bool {
	return len(v.Nodes) >= v.MAX
}

func (v *ActiveView) Add(n Node) {
	if !v.Contains(n) {
		v.Nodes = append(v.Nodes, n)
	}
}

func (v *ActiveView) Contains(n Node) bool {
	for _, m := range v.Nodes {
		if m.Equal(n) {
			return false
		}
	}
	return true
}
