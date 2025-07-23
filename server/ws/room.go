package ws

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/TLeTu/Chess-Media/server/engine"
)

type Room struct {
	ID         string
	Players    map[engine.Color]*Client
	Spectators map[*Client]bool
	Host       *Client
	GameState  string // "waiting", "in_progress", "finished"
	ReadyState map[*Client]bool

	Broadcast  chan *ClientMessage
	Register   chan *Client
	Unregister chan *Client
	Hub        *Hub
	Game       *engine.Position
}

func NewRoom(id string, hub *Hub) *Room {
	return &Room{
		ID:         id,
		Players:    make(map[engine.Color]*Client),
		Spectators: make(map[*Client]bool),
		Host:       nil,
		GameState:  "waiting",
		ReadyState: make(map[*Client]bool),
		Broadcast:  make(chan *ClientMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Hub:        hub,
		Game:       engine.NewGame(),
	}
}

// --- Helper functions ---

func (r *Room) getGuest() *Client {
	for _, p := range r.Players {
		if p != r.Host {
			return p
		}
	}
	return nil
}

func (r *Room) getAllClients() map[*Client]bool {
	all := make(map[*Client]bool)
	for _, c := range r.Players {
		if c != nil {
			all[c] = true
		}
	}
	for c := range r.Spectators {
		all[c] = true
	}
	return all
}

// --- Broadcasting functions ---

func (r *Room) broadcastLobbyState() {
	var hostReady, guestReady bool
	var hostColor string
	guest := r.getGuest()

	if r.Host != nil {
		hostReady = r.ReadyState[r.Host]
		// Find the color of the host
		for color, client := range r.Players {
			if client == r.Host {
				hostColor = color.String()
				break
			}
		}
	}
	if guest != nil {
		guestReady = r.ReadyState[guest]
	}

	for c := range r.getAllClients() {
		isHost := (c == r.Host)
		payload := LobbyStatePayload{
			HostReady:   hostReady,
			GuestReady:  guestReady,
			HostColor:   hostColor,
			IsHost:      isHost,
			GameState:   r.GameState,
			PlayerCount: len(r.Players),
		}
		message := Message{Action: "lobby_state", Payload: payload}
		messageBytes, _ := json.Marshal(message)
		c.Send <- messageBytes
	}
}

func (r *Room) broadcastGameState() {
	payload := GameStatePayload{
		FEN:        r.Game.String(),
		GameStatus: r.Game.GetGameStatus().String(),
	}
	message := Message{Action: "game_state", Payload: payload}
	messageBytes, _ := json.Marshal(message)

	for client := range r.getAllClients() {
		client.Send <- messageBytes
	}
}

func (r *Room) sendErrorMessage(client *Client, message string) {
	payload := ErrorPayload{Message: message}
	msg := Message{Action: "error", Payload: payload}
	messageBytes, _ := json.Marshal(msg)
	client.Send <- messageBytes
}

// --- Logic for handling lobby actions ---

func (r *Room) handleAssignColor(sender *Client, payload interface{}) {
	if sender != r.Host {
		r.sendErrorMessage(sender, "Only the host can assign colors.")
		return
	}

	payloadBytes, _ := json.Marshal(payload)
	var colorPayload AssignColorPayload
	json.Unmarshal(payloadBytes, &colorPayload)

	guest := r.getGuest()
	// Clear existing player assignments to be safe
	// r.Players = make(map[engine.Color]*Client)

	// Log out the colorPayload.color
	log.Printf("Color host chose: %s", colorPayload.Color)

	var hostC, guestC engine.Color
	switch colorPayload.Color {
	case "white":
		hostC, guestC = engine.White, engine.Black
	case "black":
		hostC, guestC = engine.Black, engine.White
	case "random":
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(2) == 0 {
			hostC, guestC = engine.White, engine.Black
		} else {
			hostC, guestC = engine.Black, engine.White
		}
	default:
		r.sendErrorMessage(sender, "Invalid color selection.")
		return
	}
	// Log out the host and guess color
	log.Printf("Host color: %s", hostC.String())
	log.Printf("Guest color: %s", guestC.String())

	r.Host.PlayerColor = hostC
	if guest != nil {
		guest.PlayerColor = guestC
	}
	// Log out the color assigned to host
	log.Printf("Host color assigned: %s", r.Host.PlayerColor.String())
	newPlayersMap := make(map[engine.Color]*Client)
	newPlayersMap[hostC] = r.Host
	if guest != nil {
		newPlayersMap[guestC] = guest
	}
	r.Players = newPlayersMap

	r.broadcastLobbyState()
}

func (r *Room) handlePlayerReady(sender *Client) {
	if r.GameState != "waiting" {
		return
	}
	r.ReadyState[sender] = !r.ReadyState[sender] // Toggle readiness
	r.broadcastLobbyState()
}

func (r *Room) handleStartGame(sender *Client) {
	if sender != r.Host {
		r.sendErrorMessage(sender, "Only the host can start the game.")
		return
	}
	guest := r.getGuest()
	if guest == nil {
		r.sendErrorMessage(sender, "Two players are required to start.")
		return
	}
	if !r.ReadyState[guest] {
		r.sendErrorMessage(sender, "Guest must be ready.")
		return
	}
	// Check if colors have been assigned
	if r.Host.PlayerColor == engine.NoColor || guest.PlayerColor == engine.NoColor {
		r.sendErrorMessage(sender, "The host must select a color first.")
		return
	}

	r.GameState = "in_progress"
	for color, client := range r.Players {
		if client != nil {
			var colorStr string
			if color == engine.White {
				colorStr = "white"
			} else if color == engine.Black {
				colorStr = "black"
			} else {
				continue
			}
			payload := PlayerAssignmentPayload{Color: colorStr}
			message := Message{
				Action:  "player_assigned",
				Payload: payload,
			}
			messageBytes, _ := json.Marshal(message)
			client.Send <- messageBytes
		}
	}

	r.broadcastGameState()
}

// --- Main Room Logic ---

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.handleClientRegistration(client)

		case client := <-r.Unregister:
			r.handleClientUnregistration(client)

		case clientMessage := <-r.Broadcast:
			sender := clientMessage.Client
			message := clientMessage.Message

			if r.GameState == "waiting" {
				switch message.Action {
				case "assign_color":
					r.handleAssignColor(sender, message.Payload)
				case "player_ready":
					r.handlePlayerReady(sender)
				case "start_game":
					r.handleStartGame(sender)
				default:
					log.Printf("Action '%s' not allowed during 'waiting' state.", message.Action)
				}
			} else if r.GameState == "in_progress" {
				switch message.Action {
				case "move":
					r.handleMove(sender, message.Payload)
				default:
					log.Printf("Action '%s' not allowed during 'in_progress' state.", message.Action)
				}
			}
		}
	}
}

// --- Utility and Registration/Unregistration handlers ---

func (r *Room) handleClientRegistration(client *Client) {
	client.Room = r

	if r.Host == nil {
		r.Host = client
	}

	// Assign to players or spectators
	if len(r.Players) < 2 {
		// Use a temporary, unique key for each player in the lobby to avoid collision
		tempKey := engine.Color(10 + len(r.Players)) // 10 for host, 11 for guest
		r.Players[tempKey] = client
	} else {
		r.Spectators[client] = true
	}

	r.ReadyState[client] = false // All players start as not ready
	log.Printf("Client registered to room %s. Players: %d, Spectators: %d", r.ID, len(r.Players), len(r.Spectators))
	r.broadcastLobbyState()
}

func (r *Room) handleClientUnregistration(client *Client) {
	if client == r.Host {
		log.Printf("Host disconnected from room %s. Closing room.", r.ID)
		for c := range r.getAllClients() {
			if c != client {
				r.sendErrorMessage(c, "The host has disconnected. The game has ended.")
				close(c.Send)
			}
		}
		r.Hub.deleteRoom(r.ID)
		return
	}

	delete(r.Spectators, client)
	var colorToDelete engine.Color = -1
	for color, p := range r.Players {
		if p == client {
			colorToDelete = color
			break
		}
	}
	if colorToDelete != -1 {
		delete(r.Players, colorToDelete)
	}

	delete(r.ReadyState, client)
	close(client.Send)

	log.Printf("Client unregistered from room %s. Players: %d, Spectators: %d", r.ID, len(r.Players), len(r.Spectators))
	if len(r.Players) == 0 && len(r.Spectators) == 0 {
		r.Hub.deleteRoom(r.ID)
		return
	}
	r.broadcastLobbyState()
}

// handleMove needs to be included as well
func (r *Room) handleMove(sender *Client, payload interface{}) {
	if sender.PlayerColor == engine.NoColor {
		r.sendErrorMessage(sender, "Spectators cannot make moves.")
		return
	}
	if sender.PlayerColor != r.Game.Turn {
		r.sendErrorMessage(sender, "It's not your turn.")
		return
	}
	payloadBytes, _ := json.Marshal(payload)
	var movePayload MovePayload
	json.Unmarshal(payloadBytes, &movePayload)
	moveStr := movePayload.From + movePayload.To
	if movePayload.Promotion != "" {
		moveStr += movePayload.Promotion
	}
	move, err := engine.ParseMove(r.Game, moveStr)
	if err != nil {
		r.sendErrorMessage(sender, "Invalid move: "+err.Error())
		return
	}
	r.Game = engine.ApplyMove(r.Game, move)
	r.broadcastGameState()
}
