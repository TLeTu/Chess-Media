package ws

import "github.com/gin-gonic/gin"

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, c *gin.Context) {
	roomID := c.Param("roomID")
	if roomID == "" {
		c.JSON(400, gin.H{"error": "roomID is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	client := &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: roomID,
	}

	hub.Register <- client

	go client.writePump()
	go client.readPump()
}
