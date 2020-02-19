package hyparview

// testSender needs a struct to keep the pointer to the slice
type SliceSender struct {
	Messages []Message
}

func (s *SliceSender) Send(ms ...Message) {
	s.Messages = append(s.Messages, ms...)
}

func (s *SliceSender) Reset() []Message {
	ms := s.Messages
	s.Messages = []Message{}
	return ms
}

func NewSliceSender() *SliceSender {
	n := &SliceSender{}
	n.Reset()
	return n
}
