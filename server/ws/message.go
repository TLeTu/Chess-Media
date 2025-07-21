package ws

import "encoding/json"

// Message defines the structure for messages sent over WebSocket
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// MovePayLoad defines the payload for a "move" action
type MovePayLoad struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GameStatePayload defines the payload for a "game_state" update
type GameStatePayload struct {
	FEN        string `json:"fen"`
	GameStatus string `json:"game_status"`
}

// ErrorPayload defines the payload for an "error" message
type ErrorPayload struct {
	Message string `json:"message"`
}
