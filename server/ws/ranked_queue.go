package ws

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/TLeTu/Chess-Media/server/engine"
)

const (
	ELO_DIFFERENCE_THRESHOLD = 50
	QUEUE_CHECK_INTERVAL     = 2 * time.Second
)

// QueuedPlayer represents a client waiting in the ranked queue
type QueuedPlayer struct {
	Client    *Client
	ELO       int
	JoinedAt  time.Time
	Searching bool // True if actively searching for a match
}

// RankedQueue manages players waiting for ranked games
type RankedQueue struct {
	mu      sync.Mutex
	players map[uint]*QueuedPlayer // map[userID]*QueuedPlayer
	hub     *Hub
}

// NewRankedQueue creates a new instance of RankedQueue
func NewRankedQueue(h *Hub) *RankedQueue {
	return &RankedQueue{
		players: make(map[uint]*QueuedPlayer),
		hub:     h,
	}
}

// AddPlayer adds a client to the ranked queue
func (rq *RankedQueue) AddPlayer(client *Client, elo int) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if _, exists := rq.players[client.UserID]; exists {
		log.Printf("Client %d already in ranked queue", client.UserID)
		return
	}

	rq.players[client.UserID] = &QueuedPlayer{
		Client:    client,
		ELO:       elo,
		JoinedAt:  time.Now(),
		Searching: true,
	}
	log.Printf("Client %d (ELO: %d) added to ranked queue. Current queue size: %d", client.UserID, elo, len(rq.players))
	rq.sendQueueStatus(client, "joined_queue", "Waiting for opponent...")
}

// RemovePlayer removes a client from the ranked queue
func (rq *RankedQueue) RemovePlayer(client *Client) {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if _, exists := rq.players[client.UserID]; !exists {
		return
	}

	delete(rq.players, client.UserID)
	log.Printf("Client %d removed from ranked queue. Current queue size: %d", client.UserID, len(rq.players))
	rq.sendQueueStatus(client, "left_queue", "You left the queue.")
}

// Run starts the matching process
func (rq *RankedQueue) Run() {
	ticker := time.NewTicker(QUEUE_CHECK_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		rq.FindMatches()
	}
}

// FindMatches attempts to find suitable matches for players in the queue
func (rq *RankedQueue) FindMatches() {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	var searchingPlayers []*QueuedPlayer
	for _, p := range rq.players {
		if p.Searching {
			searchingPlayers = append(searchingPlayers, p)
		}
	}

	for i := 0; i < len(searchingPlayers); i++ {
		player1 := searchingPlayers[i]
		if !player1.Searching {
			continue
		}

		for j := i + 1; j < len(searchingPlayers); j++ {
			player2 := searchingPlayers[j]
			if !player2.Searching {
				continue
			}

			if rq.isMatch(player1, player2) {
				log.Printf("Match found: %d (ELO: %d) vs %d (ELO: %d)",
					player1.Client.UserID, player1.ELO, player2.Client.UserID, player2.ELO)

				player1.Searching = false
				player2.Searching = false

				roomID := generateRoomID()
				room := NewRoom(roomID, rq.hub, true)
				rq.hub.Rooms[roomID] = room
				go room.Run()

				rq.assignPlayersToRankedRoom(room, player1.Client, player2.Client)

				delete(rq.players, player1.Client.UserID)
				delete(rq.players, player2.Client.UserID)

				break
			}
		}
	}
}

// isMatch checks if two players are a suitable match based on ELO
func (rq *RankedQueue) isMatch(p1, p2 *QueuedPlayer) bool {
	diff := p1.ELO - p2.ELO
	if diff < 0 {
		diff = -diff
	}
	return diff <= ELO_DIFFERENCE_THRESHOLD
}

// assignPlayersToRankedRoom assigns matched players to a newly created ranked room
func (rq *RankedQueue) assignPlayersToRankedRoom(room *Room, client1, client2 *Client) {
	var whitePlayerClient, blackPlayerClient *Client

	if rand.Intn(2) == 0 {
		whitePlayerClient = client1
		blackPlayerClient = client2
	} else {
		whitePlayerClient = client2
		blackPlayerClient = client1
	}

	room.PendingRankedPlayers[whitePlayerClient.UserID] = engine.White
	room.PendingRankedPlayers[blackPlayerClient.UserID] = engine.Black

	rq.sendMatchFound(whitePlayerClient, room.ID, "white")
	rq.sendMatchFound(blackPlayerClient, room.ID, "black")
}

// sendMatchFound sends a message to the client that a match has been found
func (rq *RankedQueue) sendMatchFound(client *Client, roomID string, color string) {
	log.Printf("Attempting to send match_found to client %d for room %s, color %s", client.UserID, roomID, color)
	payload := struct {
		RoomID string `json:"roomID"`
		Color  string `json:"color"`
	}{
		RoomID: roomID,
		Color:  color,
	}
	msg := Message{Action: "match_found", Payload: payload}
	messageBytes, _ := json.Marshal(msg)
	client.Send <- messageBytes
}

// sendQueueStatus sends a message to the client about their queue status
func (rq *RankedQueue) sendQueueStatus(client *Client, status string, message string) {
	payload := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  status,
		Message: message,
	}
	msg := Message{Action: "queue_status", Payload: payload}
	messageBytes, _ := json.Marshal(msg)
	client.Send <- messageBytes
}
