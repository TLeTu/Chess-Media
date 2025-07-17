package chesswebsocket

import "encoding/json"

// IncomingMessage represents a message sent from a client to the server.
// It uses a generic structure with a Type to determine how to process the payload.
type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// OutgoingMessage represents a message sent from the server to a client.
// The payload can be any data structure, which will be serialized to JSON.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
