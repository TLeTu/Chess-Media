package chesswebsocket

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active rooms and broadcasts messages.
type Hub struct {
	// Registered clients.
	Rooms map[string]*Room
	// Inbound messages from the clients.
	Broadcast chan []byte
	// Register requests from the clients.
	Register chan *Client
	// Unregister requests from clients.
	Unregister     chan *Client
	Clients        map[*Client]bool
	ClientID       uint64
	ClientMessages chan ClientMessage
	Mutex          sync.RWMutex // To protect concurrent access (prevent things like client get the same id or smthing)

}

type ClientMessage struct {
	Client  *Client
	Message IncomingMessage
}

func newHub() *Hub {
	return &Hub{
		Broadcast:      make(chan []byte),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Rooms:          make(map[string]*Room),
		Clients:        make(map[*Client]bool),
		ClientID:       0,
		ClientMessages: make(chan ClientMessage),
	}
}

// The Hub's main event loop
func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case message := <-h.ClientMessages:
			h.routeMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	// Generate a unique ID for the client
	client.ID = h.ClientID
	h.ClientID++

	// Add the client to the list of clients
	h.Clients[client] = true
	log.Printf("Client %d connected, total clients: %d", client.ID, len(h.Clients))
	// Send the client their ID
	// client.Send <- []byte("Your ID is " + string(client.ID))
	welcomeMsg := OutgoingMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"clientId": client.ID,
			"message":  "Connected to chess server",
		},
	}

	h.sendToClient(client, welcomeMsg)
}

func (h *Hub) sendToClient(client *Client, message OutgoingMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	select {
	case client.Send <- messageBytes:
	default:
		log.Printf("Client %d send channel full, removing", client.ID)
		h.unregisterClient(client)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	// Check if client exists
	if _, exists := h.Clients[client]; !exists {
		return
	}

	// Remove from clients map
	delete(h.Clients, client)

	// Close send channel safely
	select {
	case <-client.Send:
	default:
		close(client.Send)
	}

	log.Printf("Client %d disconnected, total clients: %d", client.ID, len(h.Clients))

	// Remove from any room they're in
	h.RemoveClientFromRoom(client)
}

func (h *Hub) RemoveClientFromRoom(client *Client) {
	for roomID, room := range h.Rooms {
		if _, inRoom := room.Clients[client]; inRoom {
			delete(room.Clients, client)

			// Notify remaining players
			if len(room.Clients) == 1 {
				// One player left, notify them
				for remainingClient := range room.Clients {
					disconnectMsg := OutgoingMessage{
						Type: "opponent_disconnected",
						Payload: map[string]interface{}{
							"message": "Your opponent has disconnected",
						},
					}
					h.sendToClient(remainingClient, disconnectMsg)
					remainingClient.Room = nil
				}
			}

			// Clean up empty rooms
			if len(room.Clients) == 0 {
				delete(h.Rooms, roomID)
				log.Printf("Room %s deleted (empty)", roomID)
			}

			break // Client can only be in one room
		}
	}
}

// NEW METHOD: Add this to hub.go
func (h *Hub) routeMessage(clientMsg ClientMessage) {
	switch clientMsg.Message.Type {
	case "create_room":
		h.handleCreateRoom(clientMsg.Client, clientMsg.Message.Payload)
	case "join_room":
		h.handleJoinRoom(clientMsg.Client, clientMsg.Message.Payload)
	case "make_move":
		h.handleMakeMove(clientMsg.Client, clientMsg.Message.Payload)
	default:
		log.Printf("Unknown message type from client %d: %s", clientMsg.Client.ID, clientMsg.Message.Type)

		errorMsg := OutgoingMessage{
			Type: "error",
			Payload: map[string]string{
				"message": "Unknown message type: " + clientMsg.Message.Type,
			},
		}
		h.sendToClient(clientMsg.Client, errorMsg)
	}
}

// PLACEHOLDER METHODS: Add these to hub.go (will implement in Phase 2 & 3)
func (h *Hub) handleCreateRoom(client *Client, payload json.RawMessage) {
	log.Printf("Create room request from client %d", client.ID)
	// Implementation in Phase 2
}

func (h *Hub) handleJoinRoom(client *Client, payload json.RawMessage) {
	log.Printf("Join room request from client %d", client.ID)
	// Implementation in Phase 2
}

func (h *Hub) handleMakeMove(client *Client, payload json.RawMessage) {
	log.Printf("Make move request from client %d", client.ID)
	// Implementation in Phase 3
}
