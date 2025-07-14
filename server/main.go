package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.StaticFile("/", "../client/pages/index.html")
	r.StaticFile("/bot-chess", "../client/pages/bot-chess.html")

	r.Static("/src", "../client/src")
	r.Static("/assets", "../client/assets")
	r.Static("/img", "../client/assets/img")
	r.Static("/node_modules", "../client/node_modules")

	log.Println("Starting file server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
