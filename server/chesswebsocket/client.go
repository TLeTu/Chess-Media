package chesswebsocket

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   uint64
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
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %d: %v", c.ID, err)
			}
			break
		}

		// Parse incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(messageBytes, &incomingMsg); err != nil {
			log.Printf("Invalid JSON from client %d: %v", c.ID, err)

			// Send error response
			errorMsg := OutgoingMessage{
				Type: "error",
				Payload: map[string]string{
					"message": "Invalid message format",
				},
			}
			c.Hub.sendToClient(c, errorMsg)
			continue
		}

		// Send to hub for processing
		clientMessage := ClientMessage{
			Client:  c,
			Message: incomingMsg,
		}

		select {
		case c.Hub.ClientMessages <- clientMessage:
		default:
			log.Printf("Hub message channel full, dropping message from client %d", c.ID)
		}
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
