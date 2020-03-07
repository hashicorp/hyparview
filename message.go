//go:generate genny -in=message.go -out=message-gen.go gen "genericReq=JoinRequest,ForwardJoinRequest,DisconnectRequest,NeighborRequest,NeighborRefuse,ShuffleRequest,ShuffleReply,Gossip"
package hyparview

import "github.com/cheekybits/genny/generic"

type Message interface {
	To() *Node
	AssocTo(*Node) Message
}

type genericReq generic.Type

func (r *genericReq) To() *Node {
	return r.to
}

func (r *genericReq) AssocTo(n *Node) *Node {
	o := *r
	o.to = n
	return &o
}
