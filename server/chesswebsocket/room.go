package chesswebsocket

import "github.com/TLeTu/Chess-Media/server/engine"

type Room struct {
	ID      string
	Hub     *Hub
	Clients map[*Client]bool // Using a map is an easy way to represent a set
	Game    *engine.Position
}
