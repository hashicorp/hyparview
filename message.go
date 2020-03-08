package hyparview

// Message allows clients to redefine hyparview messages to carry additional meta information
type Message interface {
	To() *Node
	AssocTo(*Node) Message
}

// Methods that can be generated should be added to message.go.genny, and build by `make
// test`. The genny generator does some funny things around interfaces and type receivers,
// so the template file isn't included in the build for now.
