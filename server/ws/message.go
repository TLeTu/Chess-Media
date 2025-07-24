package ws

// Message defines the structure for messages sent over WebSocket
type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

// MovePayLoad defines the payload for a "move" action
type MovePayload struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion,omitempty"`
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

// PlayerAssignmentPayload defines the payload for a "player_assigned" action
type PlayerAssignmentPayload struct {
	Color string `json:"color"` //white black spectator
}

// AssignColorPayload is sent by the host to choose a color
type AssignColorPayload struct {
	Color string `json:"color"` //white black
}

// LobbyStatePayload is broadcast by the server to update all clients on the lobby status
type LobbyStatePayload struct {
	HostReady   bool   `json:"host_ready"`
	GuestReady  bool   `json:"guest_ready"`
	HostColor   string `json:"host_color"`
	IsHost      bool   `json:"is_host"`
	GameState   string `json:"game_state"`
	PlayerCount int    `json:"player_count"`
	GameType    string `json:"game_type"` // "ranked" or "unranked"
}
