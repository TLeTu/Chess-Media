package main

import (
	"log"

	"github.com/TLeTu/Chess-Media/server/authentication"
	"github.com/TLeTu/Chess-Media/server/database"
	"github.com/TLeTu/Chess-Media/server/models"
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

	r := gin.Default()

	r.StaticFile("/", "../client/pages/index.html")
	r.StaticFile("/bot-chess", "../client/pages/bot-chess.html")

	r.Static("/src", "../client/src")
	r.Static("/assets", "../client/assets")
	r.Static("/img", "../client/assets/img")
	r.Static("/node_modules", "../client/node_modules")

	r.POST("/login", authentication.LoginHandler)
	r.GET("/protected", authentication.ProtectedHandler)

	log.Println("Starting file server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
