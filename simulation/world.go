package simulation

type World struct {
	nodes map[string]*Client
	// morgue map[string]*Client
	queue         []Message
	totalMessages int
	totalPayloads int
}

type WorldFailureRate struct {
	active  int
	shuffle int
	reply   int
}

type WorldConfig struct {
	rounds     int
	peers      int
	mortality  int
	drainDepth int
	payloads   int
	fail       WorldFailureRate
}

func (w *World) get(id string) *Client {
	return w.nodes[id]
}

func (w *World) nodeKeys() []string {
	m := w.nodes
	ks := make([]string, len(m))
	i := 0
	for k, _ := range m {
		ks[i] = k
		i++
	}
	return ks
}

func (w *World) randNodes() (ns []*Client) {
	for _, k := range w.nodeKeys() {
		ns = append(ns, w.get(k))
	}
	return ns
}

func (w *World) send(ms []Message) {
	w.totalMessages += len(ms)
	w.queue = append(w.queue, ms...)
}

func (w *World) sendOne(m Message) {
	w.send([]Message{m})
}

// drain the queue, appending resulting messages back onto the queue
func (w *World) drain(depth int) {
	for depth != 0 {
		if len(w.queue) == 0 {
			return
		}

		m := w.queue[0]
		w.queue = w.queue[1:]

		v := w.get(m.To().ID)
		if v != nil {
			ms := v.Recv(m)
			w.send(ms)
		}

		depth--
	}
}

func (w *World) drainAll() {
	// This should be -1, but it keeps blowing the stack. Suggests I've got a bug where
	// I never stop sending messages

	// w.drain(-1)
	w.drain(100)
}
