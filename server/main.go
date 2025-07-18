package main

import (
	"log"
	"net/http"

	"github.com/TLeTu/Chess-Media/server/authentication"
	"github.com/TLeTu/Chess-Media/server/bot"
	"github.com/TLeTu/Chess-Media/server/chesswebsocket"
	"github.com/TLeTu/Chess-Media/server/database"
	"github.com/TLeTu/Chess-Media/server/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub = chesswebsocket.newHub() // Create a single hub instance

func init() {
	go hub.Run() // start the hub's event loop in the bg
}

func WsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &chesswebsocket.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256), // Buffer of 256 messages
	}

	client.Hub.Register <- client // Register the new client

	// Start goroutines to handle reading and writting for this client
	go client.writePump()
	go client.readPump()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables if set")
	}
	database.Connect()
	database.DB.AutoMigrate(&models.User{})

	r := gin.Default()

	r.StaticFile("/", "../client/pages/index.html")
	r.StaticFile("/bot", "../client/pages/bot-chess.html")
	r.StaticFile("login", "../client/pages/login.html")

	r.Static("/src", "../client/src")
	r.Static("/assets", "../client/assets")
	r.Static("/img", "../client/assets/img")
	r.Static("/node_modules", "../client/node_modules")

	r.POST("/api/login", authentication.LoginHandler)
	r.POST("/api/register", authentication.RegisterHandler)
	r.GET("/api/validate", authentication.ValidateHandler)

	r.POST("/api/bot/move", bot.BotMoveHandler)

	log.Println("Starting file server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
