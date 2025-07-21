package main

import (
	"log"

	"github.com/TLeTu/Chess-Media/server/authentication"
	"github.com/TLeTu/Chess-Media/server/bot"
	"github.com/TLeTu/Chess-Media/server/database"
	"github.com/TLeTu/Chess-Media/server/models"
	"github.com/TLeTu/Chess-Media/server/ws"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables if set")
	}
	database.Connect()
	database.DB.AutoMigrate(&models.User{})

	// Create and run the WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()

	r.StaticFile("/", "../client/pages/index.html")
	r.StaticFile("/bot", "../client/pages/bot-chess.html")
	r.StaticFile("login", "../client/pages/login.html")
	r.StaticFile("/game", "../client/pages/game.html")

	r.Static("/src", "../client/src")
	r.Static("/assets", "../client/assets")
	r.Static("/img", "../client/assets/img")
	r.Static("/node_modules", "../client/node_modules")

	r.POST("/api/login", authentication.LoginHandler)
	r.POST("/api/register", authentication.RegisterHandler)
	r.GET("/api/validate", authentication.ValidateHandler)

	r.POST("/api/bot/move", bot.BotMoveHandler)

	r.POST("/api/rooms/create", ws.CreateRoomHandler)

	// WebSocket endpoint
	r.GET("/ws/game/:roomID", func(c *gin.Context) {
		ws.ServeWs(hub, c)
	})

	// For websocket testing
	r.StaticFile("/wstest", "../client/pages/wstest.html")

	log.Println("Starting file server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
