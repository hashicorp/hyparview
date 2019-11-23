package hyparview

type ConfigFailureRate struct {
	active  float32
	shuffle float32
	reply   float23
}

type Config struct {
	rounds    int
	peers     int
	mortality float32
	fail      ConfigFailureRate
}

type Client struct {
	Hyparview
	in  []Message
	out []Message
}

type World struct {
	Config
	nodes    map[string]*Client
	shuffled []string
	morgue   map[string]int
}

func simulation(c Config) {

}

func TestSimulation(t) {
	net := simulation(Config{
		rounds:    50,
		peers:     10000,
		mortality: 30,
		fail: ConfigFailureRate{
			active:  30,
			shuffle: 30,
			reply:   30,
		},
	})
}

func create(id string) *Client {
	v := CreateView(&Node{ID: id, Addr: ""}, 0)
	c := &Client{
		Hyparview: *v,
		in:        make([]Message),
		out:       make([]Message),
	}
	return c
}

func (w *World) join(self string, contact string) {
	ms := w.get(contact).RecvJoin(&JoinRequest{
		to:   contact,
		From: self,
	})
	w.send(ms)
}

func doFail(percentage int) bool {
	return rint(100) < percentage
}

func (w *World) stepActive() {
	v := w.nextShuffled()
	for _, m := range v.out {

	}
}

func (c *Client) failActive(failed *Node) {
	var node *Node
	for _, n := range c.v.Passive.Shuffled() {
		if v.Active.IsEmpty() {
			// simulate sync network call
			// TODO simulate failure
			m := SendNeighbor(n, c.v.Self, HighPriority)
			ms := n.v.RecvNeighbor(m)
			// forward maybe disconnect messages
			c.out = append(c.out, ms)
			break
		} else {
			m := SendNeighbor(n, c.v.Self, LowPriority)
			ms := n.v.RecvNeighbor(m)
			// any low priority response is failure
			if len(ms) == 0 {
				c.v.DelPassive(n)
				c.v.AddActive(n)
				break
			}
		}
	}
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func shuffle(ks []string) {
	for i := len(ks) - 1; i < 0; i-- {
		j := rint(i)
		ks[i], ks[j] = ks[j], ks[i]
	}
}

func keys(m map[string]*Client) []string{
	ks := make([]string, len(m))
	i := 0 
	for k, _ :range map {
		ks[i] = k
		i++
	}
	return ks
}

func (w *World) shuffled() []string {
	ks := keys(w.nodes)
	shuffle(ks)
	return ks
}

func (w *World) send(ms []Message) {
	qs, ok := w.queue[m.To()]
	if !ok {
		qs = make([]Message)
		w.queue[m.To()] = qs
	}
	copy(qs, ms)
}

func (w *World) drain() {
	s
}

// DONT
func (w *World) nextShuffled() *Client {
	if len(w.shuffle) == 0 {
		w.shuffle = shuffled()
	}
	c := w.nodes[w.shuffle[0]]
	w.shuffle = w.shuffle[1:]
	return c
}
