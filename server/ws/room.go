package ws

import (
	"encoding/json"
	"log"

	"github.com/TLeTu/Chess-Media/server/engine"
)

type Room struct {
	ID         string
	Players    map[engine.Color]*Client
	Spectators map[*Client]bool

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
		Players:    make(map[engine.Color]*Client),
		Spectators: make(map[*Client]bool),
		Broadcast:  make(chan *ClientMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Hub:        hub,
		Game:       engine.NewGame(), // Initialize a new game
	}
}

// assignColorAndNotify assigns a color to a client and notifies them.
func (r *Room) assignColorAndNotify(client *Client) {
	var color engine.Color
	var colorStr string

	// Find an available player slot
	if _, ok := r.Players[engine.White]; !ok {
		color = engine.White
		colorStr = "white"
	} else if _, ok := r.Players[engine.Black]; !ok {
		color = engine.Black
		colorStr = "black"
	} else {
		// Room is full, assign as spectator
		r.Spectators[client] = true
		colorStr = "spectator"
		// Notify the client of their role
		payload := PlayerAssignmentPayload{Color: colorStr}
		message := Message{
			Action:  "player_assigned",
			Payload: payload,
		}
		messageBytes, _ := json.Marshal(message)
		client.Send <- messageBytes
		return // Exit early for spectators
	}

	r.Players[color] = client
	client.PlayerColor = color

	// Notify the client of their color
	payload := PlayerAssignmentPayload{Color: colorStr}
	message := Message{
		Action:  "player_assigned",
		Payload: payload,
	}
	messageBytes, _ := json.Marshal(message)
	client.Send <- messageBytes
}

// broadcastGameState sends the current game state to all clients in the room
func (r *Room) broadcastGameState() {
	payload := GameStatePayload{
		FEN:        r.Game.String(),
		GameStatus: r.Game.GetGameStatus().String(),
	}

	message := Message{
		Action:  "game_state",
		Payload: payload,
	}
	messageBytes, _ := json.Marshal(message)
	// send to players
	for _, client := range r.Players {
		client.Send <- messageBytes
	}
	// send to spectators
	for spectator := range r.Spectators {
		spectator.Send <- messageBytes
	}

}

// sendErrorMessage sends an error message to a specific client
func (r *Room) sendErrorMessage(client *Client, message string) {
	payload := ErrorPayload{
		Message: message,
	}
	messageBytes, _ := json.Marshal(Message{
		Action:  "error",
		Payload: payload,
	})
	client.Send <- messageBytes
}

// handleMove attempts to apply a game and broadcasts the new state
func (r *Room) handleMove(sender *Client, payload interface{}) {
	// Check if the sender is a player
	if sender.PlayerColor == engine.NoColor {
		r.sendErrorMessage(sender, "You are not a player")
		return
	}

	// Check if it is the sender's turn
	if sender.PlayerColor != r.Game.Turn {
		r.sendErrorMessage(sender, "It is not your turn")
		return
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		r.sendErrorMessage(sender, "Invalid move format")
		return
	}

	var movePayload MovePayLoad
	if err := json.Unmarshal(payloadBytes, &movePayload); err != nil {
		r.sendErrorMessage(sender, "Invalid move payload")
		return
	}

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
			client.Room = r
			r.assignColorAndNotify(client)
			log.Printf("Client registered to room %s", r.ID)
			r.broadcastGameState()
		case client := <-r.Unregister:
			// remove from the spectators map
			if _, ok := r.Spectators[client]; ok {
				delete(r.Spectators, client)
			}
			// remove from the players map
			if client.PlayerColor != engine.NoColor {
				delete(r.Players, client.PlayerColor)
			}
			// close the client's send channel
			close(client.Send)
			log.Printf("Client unregistered from room %s", r.ID)
			if (len(r.Players) == 0) && len(r.Spectators) == 0 {
				r.Hub.deleteRoom(r.ID)
				log.Printf("Room %s deleted", r.ID)
				return
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
