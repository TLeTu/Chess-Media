package ws

import (
	"encoding/json"
	"log"

	"github.com/TLeTu/Chess-Media/server/engine"
)

type Room struct {
	ID      string
	Clients map[*Client]bool
	// Inbound messages from the clients
	Broadcast chan *ClientMessage
	// Register requests from the clients
	Register chan *Client
	// Unregister requests from clients
	Unregister chan *Client
	// Hub that manages this room
	Hub *Hub
	// Game state
	Game *engine.Position
}

func NewRoom(id string, hub *Hub) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *ClientMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Hub:        hub,
		Game:       engine.NewGame(), // Initialize a new game
	}
}

// broadcastGameState sends the current game state to all clients in the room
func (r *Room) broadcastGameState() {
	payload := GameStatePayload{
		FEN:        r.Game.String(),
		GameStatus: r.Game.GetGameStatus().String(),
	}
	payloadBytes, _ := json.Marshal(payload)
	message := Message{
		Action:  "game_state",
		Payload: payloadBytes,
	}
	messageBytes, _ := json.Marshal(message)
	for client := range r.Clients {
		select {
		case client.Send <- messageBytes:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}

}

// sendErrorMessage sends an error message to a specific client
func (r *Room) sendErrorMessage(client *Client, message string) {
	payload := ErrorPayload{
		Message: message,
	}
	payloadBytes, _ := json.Marshal(payload)
	messageBytes, _ := json.Marshal(Message{
		Action:  "error",
		Payload: payloadBytes,
	})
	messageBytes = append(messageBytes, '\n')
	client.Send <- messageBytes
}

// handleMove attempts to apply a game and broadcasts the new state
func (r *Room) handleMove(sender *Client, payload json.RawMessage) {
	var movePayload MovePayLoad
	if err := json.Unmarshal(payload, &movePayload); err != nil {
		r.sendErrorMessage(sender, "Invalid move payload")
		return
	}
	// TODO: check if it's the sender's turn. For now, we allow anyone to move
	move, err := engine.ParseMove(r.Game, movePayload.From+movePayload.To)
	if err != nil {
		r.sendErrorMessage(sender, err.Error())
		return
	}
	r.Game = engine.ApplyMove(r.Game, move)
	r.broadcastGameState()
}

// Start the room's event loop
func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
			client.Room = r
			log.Printf("Client registered to room %s", r.ID)
			r.broadcastGameState()
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
				log.Printf("Client unregistered from room %s", r.ID)
				if len(r.Clients) == 0 {
					r.Hub.deleteRoom(r.ID)
					return
				}
			}
		case clientMessage := <-r.Broadcast:
			sender := clientMessage.Client
			message := clientMessage.Message

			switch message.Action {
			case "move":
				r.handleMove(sender, message.Payload)
			default:
				// r.sendErrorMessage(sender, "Invalid action")
				log.Printf("Invalid action: %s", message.Action)
			}

		}
	}
}
