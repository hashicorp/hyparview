package hyparview

// testSender needs a struct to keep the pointer to the slice
type sliceSender struct {
	ms []Message
}

func (s *sliceSender) Send(ms ...Message) error {
	s.ms = append(s.ms, ms...)
	return nil
}

func (s *sliceSender) Fail(n *Node) {
}

func (s *sliceSender) reset() []Message {
	ms := s.ms
	s.ms = []Message{}
	return ms
}

func newSliceSender() *sliceSender {
	n := &sliceSender{}
	n.reset()
	return n
}
