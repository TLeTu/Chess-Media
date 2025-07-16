package bot

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/TLeTu/Chess-Media/server/engine"
	"github.com/gin-gonic/gin"
)

type MoveRequest struct {
	CurrentFEN     string `json:"currentFen"`
	PlayerMove     string `json:"playerMove"`
	PromotionPiece string `json:"promotionPiece"`
}

type MoveResponse struct {
	NewFEN     string `json:"newFen"`
	GameStatus string `json:"gameStatus"`
}

func BotMoveHandler(c *gin.Context) {
	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("Received move request: %+v\n", req)

	// Parse the current FEN into a Position object
	currentPos, err := engine.ParseFEN(req.CurrentFEN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid FEN: %v", err)})
		return
	}

	// Generate all legal moves for the current position
	legalMoves := currentPos.GenerateLegalMoves()

	// Find the player's move among the legal moves
	playerMove := engine.Move{}
	foundPlayerMove := false
	for _, move := range legalMoves {
		// For promotion moves, check if the promotion piece matches
		if move.Promotion != engine.NoPieceType && req.PromotionPiece != "" {
			promotedPieceType := engine.PieceType(0)
			switch strings.ToLower(req.PromotionPiece) {
			case "q":
				promotedPieceType = engine.Queen
			case "r":
				promotedPieceType = engine.Rook
			case "b":
				promotedPieceType = engine.Bishop
			case "n":
				promotedPieceType = engine.Knight
			}
			if move.From.String()+move.To.String() == req.PlayerMove && move.Promotion == promotedPieceType {
				playerMove = move
				foundPlayerMove = true
				break
			}
		} else if move.From.String()+move.To.String() == req.PlayerMove && move.Promotion == engine.NoPieceType {
			playerMove = move
			foundPlayerMove = true
			break
		}
	}

	if !foundPlayerMove {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player move"})
		return
	}

	// Apply the player's move
	newPos := engine.ApplyMove(currentPos, playerMove)

	// Check game status after player's move
	gameStatus := "in_progress"
	if len(newPos.GenerateLegalMoves()) == 0 {
		if engine.IsKingInCheck(newPos, newPos.Turn) {
			gameStatus = "checkmate"
		} else {
			gameStatus = "stalemate"
		}
	}

	if gameStatus != "in_progress" {
		c.JSON(http.StatusOK, MoveResponse{
			NewFEN:     newPos.String(),
			GameStatus: gameStatus,
		})
		return
	}

	// Bot's turn: Generate a random legal move for the bot
	botLegalMoves := newPos.GenerateLegalMoves()
	if len(botLegalMoves) > 0 {
		rand.Seed(time.Now().UnixNano())
		botMove := botLegalMoves[rand.Intn(len(botLegalMoves))]
		newPos = engine.ApplyMove(newPos, botMove)

		// Check game status after bot's move
		if len(newPos.GenerateLegalMoves()) == 0 {
			if engine.IsKingInCheck(newPos, newPos.Turn) {
				gameStatus = "checkmate"
			} else {
				gameStatus = "stalemate"
			}
		}
	}

	// Respond with the new FEN and game status
	c.JSON(http.StatusOK, MoveResponse{
		NewFEN:     newPos.String(),
		GameStatus: gameStatus,
	})
}
