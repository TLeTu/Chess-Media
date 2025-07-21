package ws

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateRoomHandler generates a unique room ID and returns it to the client
func CreateRoomHandler(c *gin.Context) {
	// Generate a random 8-charater hex string for the room ID
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	roomID := hex.EncodeToString(bytes)
	c.JSON(http.StatusOK, gin.H{"roomID": roomID})
}
