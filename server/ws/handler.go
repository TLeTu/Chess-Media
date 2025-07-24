package ws

import (
	"log"

	"github.com/TLeTu/Chess-Media/server/authentication"
	"github.com/TLeTu/Chess-Media/server/database"
	"github.com/TLeTu/Chess-Media/server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, c *gin.Context) {
	roomID := c.Param("roomID")
	if roomID == "" {
		c.JSON(400, gin.H{"error": "roomID is required"})
		return
	}

	// All WebSocket connections must be authenticated.
	tokenString := c.Query("token")
	if tokenString == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "missing authentication token"})
		return
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return authentication.GetSecretKey(), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		return
	}

	email, ok := (*claims)["username"].(string)
	if !ok {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid token claims"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "user not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection for room %s: %v", roomID, err)
		return
	}

	client := &Client{
		Hub:     hub,
		UserID:  user.ID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		RoomID:  roomID,
		UserELO: user.ELO,
		User:    &user,
	}

	hub.Register <- client

	go client.writePump()
	go client.readPump()
}