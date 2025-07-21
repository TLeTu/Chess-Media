package ws

import "log"

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Check if the room already exists
			if _, ok := h.Rooms[client.RoomID]; ok {
				// Room exists, register the client to it
				h.Rooms[client.RoomID].Register <- client
			} else {
				// Room doesn't exist, create it
				room := NewRoom(client.RoomID, h)
				h.Rooms[client.RoomID] = room
				go room.Run()
				room.Register <- client
				log.Printf("New room created: %s", client.RoomID)
			}
		case client := <-h.Unregister:
			// Unregister the client from its room
			if _, ok := h.Rooms[client.RoomID]; ok {
				h.Rooms[client.RoomID].Unregister <- client
			}

		}
	}
}

// Called from a Room when it becomes empty
func (h *Hub) deleteRoom(roomID string) {
	if _, ok := h.Rooms[roomID]; ok {
		delete(h.Rooms, roomID)
		log.Printf("Room %s deleted", roomID)
	}
}
