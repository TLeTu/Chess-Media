package ws

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/TLeTu/Chess-Media/server/database"
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

	IsRanked bool

	PendingRankedPlayers map[uint]engine.Color // map[userID]assignedColor
}

func NewRoom(id string, hub *Hub, isRanked bool) *Room {
	return &Room{
		ID:                   id,
		Players:              make(map[engine.Color]*Client),
		Spectators:           make(map[*Client]bool),
		Host:                 nil,
		GameState:            "waiting",
		ReadyState:           make(map[*Client]bool),
		Broadcast:            make(chan *ClientMessage),
		Register:             make(chan *Client),
		Unregister:           make(chan *Client),
		Hub:                  hub,
		Game:                 engine.NewGame(),
		IsRanked:             isRanked,
		PendingRankedPlayers: make(map[uint]engine.Color),
	}
}

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

func (r *Room) broadcastLobbyState() {
	if r.IsRanked {
		return
	}
	var hostReady, guestReady bool
	var hostColor string
	guest := r.getGuest()

	if r.Host != nil {
		hostReady = r.ReadyState[r.Host]
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
			GameType:    "unranked",
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

func (r *Room) handleAssignColor(sender *Client, payload interface{}) {
	if sender != r.Host {
		r.sendErrorMessage(sender, "Only the host can assign colors.")
		return
	}

	payloadBytes, _ := json.Marshal(payload)
	var colorPayload AssignColorPayload
	json.Unmarshal(payloadBytes, &colorPayload)

	guest := r.getGuest()
	var hostC, guestC engine.Color
	switch colorPayload.Color {
	case "white":
		hostC, guestC = engine.White, engine.Black
	case "black":
		hostC, guestC = engine.Black, engine.White
	case "random":
		if rand.Intn(2) == 0 {
			hostC, guestC = engine.White, engine.Black
		} else {
			hostC, guestC = engine.Black, engine.White
		}
	default:
		r.sendErrorMessage(sender, "Invalid color selection.")
		return
	}

	r.Host.PlayerColor = hostC
	if guest != nil {
		guest.PlayerColor = guestC
	}

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
	r.ReadyState[sender] = !r.ReadyState[sender]
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
			}
			payload := PlayerAssignmentPayload{Color: colorStr}
			message := Message{Action: "player_assigned", Payload: payload}
			messageBytes, _ := json.Marshal(message)
			client.Send <- messageBytes
		}
	}

	r.broadcastGameState()
}

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

			if r.GameState == "waiting" && !r.IsRanked {
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
				if message.Action == "move" {
					r.handleMove(sender, message.Payload)
				} else {
					log.Printf("Action '%s' not allowed during 'in_progress' state.", message.Action)
				}
			}
		}
	}
}

func (r *Room) handleClientRegistration(client *Client) {
	client.Room = r

	if r.IsRanked {
		assignedColor, ok := r.PendingRankedPlayers[client.UserID]
		if !ok {
			log.Printf("Error: Ranked client %d connected without pending data.", client.UserID)
			r.sendErrorMessage(client, "Error: Could not join ranked game. Missing player data.")
			close(client.Send)
			return
		}

		client.PlayerColor = assignedColor
		r.Players[assignedColor] = client
		delete(r.PendingRankedPlayers, client.UserID)

		log.Printf("Client %d registered to ranked room %s. Color: %s", client.UserID, r.ID, client.PlayerColor.String())

		if len(r.Players) == 2 {
			r.GameState = "in_progress"
			for color, p := range r.Players {
				if p != nil {
					var colorStr string
					if color == engine.White {
						colorStr = "white"
					} else if color == engine.Black {
						colorStr = "black"
					}
					payload := PlayerAssignmentPayload{Color: colorStr}
					message := Message{Action: "player_assigned", Payload: payload}
					messageBytes, _ := json.Marshal(message)
					p.Send <- messageBytes
				}
			}
			r.broadcastGameState()
		}
		return
	}

	// Unranked game logic
	if r.Host == nil {
		r.Host = client
	}

	if len(r.Players) < 2 {
		tempKey := engine.Color(10 + len(r.Players))
		r.Players[tempKey] = client
	} else {
		r.Spectators[client] = true
	}

	r.ReadyState[client] = false
	log.Printf("Client registered to room %s. Players: %d, Spectators: %d", r.ID, len(r.Players), len(r.Spectators))
	r.broadcastLobbyState()
}

func (r *Room) handleClientUnregistration(client *Client) {
	if r.IsRanked {
		log.Printf("Client %d unregistered from ranked room %s.", client.UserID, r.ID)
		for _, p := range r.Players {
			if p != nil && p != client {
				r.sendErrorMessage(p, "Opponent disconnected. Game ended.")
				close(p.Send)
			}
		}
		for s := range r.Spectators {
			r.sendErrorMessage(s, "Game ended due to player disconnection.")
			close(s.Send)
		}
		r.Hub.deleteRoom(r.ID)
		return
	}

	// Unranked game logic
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

	gameStatus := r.Game.GetGameStatus()
	if gameStatus != engine.InProgress {
		log.Printf("Game %s ended with status: %s", r.ID, gameStatus.String())

		if r.IsRanked {
			var winner, loser *Client
			if gameStatus == engine.Checkmate {
				winner = r.Players[r.Game.Turn.Opponent()]
				loser = r.Players[r.Game.Turn]
			} else {
				log.Printf("Ranked game %s ended in a draw. No ELO changes.", r.ID)
			}

			if winner != nil && loser != nil {
				winner.UserELO += 100
				loser.UserELO -= 50

				if err := database.UpdateUserELO(winner.User.ID, winner.UserELO); err != nil {
					log.Printf("Error updating ELO for winner %d: %v", winner.UserID, err)
				}
				if err := database.UpdateUserELO(loser.User.ID, loser.UserELO); err != nil {
					log.Printf("Error updating ELO for loser %d: %v", loser.UserID, err)
				}

				log.Printf("ELO updated: Winner %d (New ELO: %d), Loser %d (New ELO: %d)",
					winner.UserID, winner.UserELO, loser.UserID, loser.UserELO)
			}
		}

		for _, p := range r.Players {
			if p != nil {
				r.sendErrorMessage(p, "Game Over: "+gameStatus.String())
				close(p.Send)
			}
		}
		for s := range r.Spectators {
			r.sendErrorMessage(s, "Game Over: "+gameStatus.String())
			close(s.Send)
		}
		r.Hub.deleteRoom(r.ID)
		return
	}

	r.broadcastGameState()
}
