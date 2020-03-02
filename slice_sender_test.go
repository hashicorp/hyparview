package hyparview

// testSender needs a struct to keep the pointer to the slice
type SliceSender struct {
	ms []Message
}

func (s *SliceSender) Send(ms ...Message) error {
	s.ms = append(s.ms, ms...)
	return nil
}

func (s *SliceSender) Fail(n *Node) {
}

func (s *SliceSender) reset() []Message {
	ms := s.ms
	s.ms = []Message{}
	return ms
}

func NewSliceSender() *SliceSender {
	n := &SliceSender{}
	n.reset()
	return n
}
