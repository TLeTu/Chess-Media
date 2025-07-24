package ws

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// generateRoomID generates a unique 8-character hex string for a room ID.
func generateRoomID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// In a real application, you might want to handle this error more robustly
		// For now, we'll just log it and return a less random ID
		return hex.EncodeToString([]byte("fallback")) // Fallback ID
	}
	return hex.EncodeToString(bytes)
}

// CreateRoomHandler generates a unique room ID and returns it to the client
func CreateRoomHandler(c *gin.Context) {
	roomID := generateRoomID()
	c.JSON(http.StatusOK, gin.H{"roomID": roomID})
}
