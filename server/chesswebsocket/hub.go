package chesswebsocket

// Hub maintains the set of active rooms and broadcasts messages.
type Hub struct {
	// Registered clients.
	Rooms map[string]*Room
	// Inbound messages from the clients.
	Broadcast chan []byte
	// Register requests from the clients.
	Register chan *Client
	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[string]*Room),
	}
}

// The Hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Handle new client registration
		case client := <-h.Unregister:
			// Handle client unregistration
		case message := <-h.Broadcast:
			// Handle incoming messages and broadcast to the correct room
		}
	}
}
