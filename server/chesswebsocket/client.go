package chesswebsocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte // A buffered channel of outbound messages
	Room *Room
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// Handle error (e.g., client disconnected)
			break
		}
		// Here, you will parse the message and decide what to do.
		// Then you'll forward it to the hub for processing.
		// // For example: hub.broadcast <- processedMessage
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	defer c.Conn.Close()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.

				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
