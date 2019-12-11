package simulation

type Client struct {
	Hyparview
	app      int
	appSeen  int
	appWaste int
	in       []Message
	out      []Message
}

func create(id string) *Client {
	v := CreateView(&Node{ID: id, Addr: ""}, 0)
	c := &Client{
		Hyparview: *v,
		in:        make([]Message, 0),
		out:       make([]Message, 0),
	}
	return c
}
