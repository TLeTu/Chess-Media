package ws

import "log"

type Hub struct {
	Rooms       map[string]*Room
	Register    chan *Client
	Unregister  chan *Client
	RankedQueue *RankedQueue
}

func NewHub() *Hub {
	hub := &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
	hub.RankedQueue = NewRankedQueue(hub)
	return hub
}

func (h *Hub) Run() {
	go h.RankedQueue.Run() // Start the ranked queue's matching process

	for {
		select {
		case client := <-h.Register:
			if client.RoomID == "ranked" {
				log.Printf("Client %d added to ranked queue", client.UserID)
				client.Hub.RankedQueue.AddPlayer(client, client.UserELO)
			} else {
				if _, ok := h.Rooms[client.RoomID]; !ok {
					// Create a new unranked room if it doesn't exist
					room := NewRoom(client.RoomID, h, false)
					h.Rooms[client.RoomID] = room
					go room.Run()
					log.Printf("New unranked room created: %s", client.RoomID)
				}
				h.Rooms[client.RoomID].Register <- client
			}
		case client := <-h.Unregister:
			if client.RoomID == "ranked" {
				h.RankedQueue.RemovePlayer(client)
			} else if room, ok := h.Rooms[client.RoomID]; ok {
				room.Unregister <- client
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