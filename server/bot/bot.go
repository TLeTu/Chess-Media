package bot

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"

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

// ChessBot represents an intelligent chess bot
type ChessBot struct {
	maxDepth int
}

// NewChessBot creates a new chess bot with specified search depth
func NewChessBot(depth int) *ChessBot {
	return &ChessBot{maxDepth: depth}
}

// Piece values for evaluation
var pieceValues = map[engine.PieceType]int{
	engine.Pawn:   100,
	engine.Knight: 320,
	engine.Bishop: 330,
	engine.Rook:   500,
	engine.Queen:  900,
	engine.King:   20000,
}

// Piece-square tables for positional evaluation
var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMiddleGameTable = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

var kingEndGameTable = [64]int{
	-50, -40, -30, -20, -20, -30, -40, -50,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-50, -30, -30, -30, -30, -30, -30, -50,
}

// evaluatePosition evaluates the current position from the perspective of the given color
func (bot *ChessBot) evaluatePosition(pos *engine.Position, color engine.Color) int {
	score := 0

	// Material and positional evaluation
	for sq := engine.A1; sq <= engine.H8; sq++ {
		piece := pos.Board[sq]
		if piece == engine.Empty {
			continue
		}

		pieceValue := pieceValues[piece.Type()]
		positionValue := bot.getPositionValue(piece, sq, pos)

		if piece.Color() == color {
			score += pieceValue + positionValue
		} else {
			score -= pieceValue + positionValue
		}
	}

	// Mobility bonus
	legalMoves := pos.GenerateLegalMoves()
	if pos.Turn == color {
		score += len(legalMoves) * 10
	} else {
		score -= len(legalMoves) * 10
	}

	// King safety
	score += bot.evaluateKingSafety(pos, color)
	score -= bot.evaluateKingSafety(pos, oppositeColor(color))

	// Pawn structure
	score += bot.evaluatePawnStructure(pos, color)
	score -= bot.evaluatePawnStructure(pos, oppositeColor(color))

	return score
}

// getPositionValue returns the positional value of a piece on a given square
func (bot *ChessBot) getPositionValue(piece engine.Piece, sq engine.Square, pos *engine.Position) int {
	index := int(sq)

	// Flip the table for black pieces
	if piece.Color() == engine.Black {
		index = 63 - index
	}

	switch piece.Type() {
	case engine.Pawn:
		return pawnTable[index]
	case engine.Knight:
		return knightTable[index]
	case engine.Bishop:
		return bishopTable[index]
	case engine.Rook:
		return rookTable[index]
	case engine.Queen:
		return queenTable[index]
	case engine.King:
		if bot.isEndGame(pos) {
			return kingEndGameTable[index]
		}
		return kingMiddleGameTable[index]
	}
	return 0
}

// isEndGame determines if the position is in the endgame
func (bot *ChessBot) isEndGame(pos *engine.Position) bool {
	queens := 0
	minorPieces := 0

	for sq := engine.A1; sq <= engine.H8; sq++ {
		piece := pos.Board[sq]
		switch piece.Type() {
		case engine.Queen:
			queens++
		case engine.Knight, engine.Bishop:
			minorPieces++
		}
	}

	// Endgame if no queens or very few pieces
	return queens == 0 || (queens == 2 && minorPieces <= 1)
}

// evaluateKingSafety evaluates king safety
func (bot *ChessBot) evaluateKingSafety(pos *engine.Position, color engine.Color) int {
	safety := 0

	// Find the king
	var kingSquare engine.Square = engine.NoSquare
	for sq := engine.A1; sq <= engine.H8; sq++ {
		piece := pos.Board[sq]
		if piece.Type() == engine.King && piece.Color() == color {
			kingSquare = sq
			break
		}
	}

	if kingSquare == engine.NoSquare {
		return -10000 // King not found (shouldn't happen)
	}

	// Check for pawn shield
	if color == engine.White && kingSquare >= engine.A1 && kingSquare <= engine.H2 {
		// White king on back ranks
		safety += bot.countPawnShield(pos, kingSquare, color) * 10
	} else if color == engine.Black && kingSquare >= engine.A7 && kingSquare <= engine.H8 {
		// Black king on back ranks
		safety += bot.countPawnShield(pos, kingSquare, color) * 10
	}

	// Penalty for king in center during middle game
	if !bot.isEndGame(pos) {
		kingFile := int(kingSquare % 8)
		kingRank := int(kingSquare / 8)
		if kingFile >= 2 && kingFile <= 5 && kingRank >= 2 && kingRank <= 5 {
			safety -= 20
		}
	}

	return safety
}

// countPawnShield counts pawns protecting the king
func (bot *ChessBot) countPawnShield(pos *engine.Position, kingSquare engine.Square, color engine.Color) int {
	count := 0
	kingFile := int(kingSquare % 8)
	kingRank := int(kingSquare / 8)

	direction := 1
	if color == engine.Black {
		direction = -1
	}

	// Check pawns in front of king
	for fileOffset := -1; fileOffset <= 1; fileOffset++ {
		file := kingFile + fileOffset
		if file >= 0 && file < 8 {
			rank := kingRank + direction
			if rank >= 0 && rank < 8 {
				sq := engine.Square(rank*8 + file)
				piece := pos.Board[sq]
				if piece.Type() == engine.Pawn && piece.Color() == color {
					count++
				}
			}
		}
	}

	return count
}

// evaluatePawnStructure evaluates pawn structure
func (bot *ChessBot) evaluatePawnStructure(pos *engine.Position, color engine.Color) int {
	score := 0

	// Count pawns per file
	fileCounts := make([]int, 8)
	for sq := engine.A1; sq <= engine.H8; sq++ {
		piece := pos.Board[sq]
		if piece.Type() == engine.Pawn && piece.Color() == color {
			file := int(sq % 8)
			fileCounts[file]++
		}
	}

	// Penalty for doubled pawns
	for _, count := range fileCounts {
		if count > 1 {
			score -= (count - 1) * 10
		}
	}

	// Bonus for passed pawns
	score += bot.countPassedPawns(pos, color) * 20

	return score
}

// countPassedPawns counts passed pawns for a color
func (bot *ChessBot) countPassedPawns(pos *engine.Position, color engine.Color) int {
	count := 0

	for sq := engine.A1; sq <= engine.H8; sq++ {
		piece := pos.Board[sq]
		if piece.Type() == engine.Pawn && piece.Color() == color {
			if bot.isPassedPawn(pos, sq, color) {
				count++
			}
		}
	}

	return count
}

// isPassedPawn checks if a pawn is passed
func (bot *ChessBot) isPassedPawn(pos *engine.Position, pawnSquare engine.Square, color engine.Color) bool {
	file := int(pawnSquare % 8)
	rank := int(pawnSquare / 8)

	direction := 1
	if color == engine.Black {
		direction = -1
	}

	// Check if there are enemy pawns blocking this pawn's path
	for checkRank := rank + direction; checkRank >= 0 && checkRank < 8; checkRank += direction {
		for fileOffset := -1; fileOffset <= 1; fileOffset++ {
			checkFile := file + fileOffset
			if checkFile >= 0 && checkFile < 8 {
				sq := engine.Square(checkRank*8 + checkFile)
				piece := pos.Board[sq]
				if piece.Type() == engine.Pawn && piece.Color() != color {
					return false
				}
			}
		}
	}

	return true
}

// minimax implements the minimax algorithm with alpha-beta pruning
func (bot *ChessBot) minimax(pos *engine.Position, depth int, alpha, beta int, maximizingPlayer bool, botColor engine.Color) int {
	if depth == 0 || pos.GetGameStatus() != engine.InProgress {
		return bot.evaluatePosition(pos, botColor)
	}

	moves := pos.GenerateLegalMoves()

	if maximizingPlayer {
		maxEval := math.MinInt32
		for _, move := range moves {
			newPos := engine.ApplyMove(pos, move)
			eval := bot.minimax(newPos, depth-1, alpha, beta, false, botColor)
			maxEval = max(maxEval, eval)
			alpha = max(alpha, eval)
			if beta <= alpha {
				break // Alpha-beta pruning
			}
		}
		return maxEval
	} else {
		minEval := math.MaxInt32
		for _, move := range moves {
			newPos := engine.ApplyMove(pos, move)
			eval := bot.minimax(newPos, depth-1, alpha, beta, true, botColor)
			minEval = min(minEval, eval)
			beta = min(beta, eval)
			if beta <= alpha {
				break // Alpha-beta pruning
			}
		}
		return minEval
	}
}

// getBestMove finds the best move using minimax algorithm
func (bot *ChessBot) getBestMove(pos *engine.Position) engine.Move {
	moves := pos.GenerateLegalMoves()
	if len(moves) == 0 {
		return engine.Move{} // No legal moves
	}

	bestMove := moves[0]
	bestValue := math.MinInt32
	botColor := pos.Turn

	for _, move := range moves {
		newPos := engine.ApplyMove(pos, move)
		value := bot.minimax(newPos, bot.maxDepth-1, math.MinInt32, math.MaxInt32, false, botColor)

		if value > bestValue {
			bestValue = value
			bestMove = move
		}
	}

	log.Printf("Bot chose move %s with evaluation %d", bestMove.String(), bestValue)
	return bestMove
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func oppositeColor(c engine.Color) engine.Color {
	if c == engine.White {
		return engine.Black
	}
	return engine.White
}

// Global bot instance
var smartBot = NewChessBot(4) // Search depth of 4 moves

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
	status := newPos.GetGameStatus()
	gameStatus := status.String()

	if gameStatus != "in_progress" {
		c.JSON(http.StatusOK, MoveResponse{
			NewFEN:     newPos.String(),
			GameStatus: gameStatus,
		})
		return
	}

	// Bot's turn: Use the smart bot to find the best move
	botMove := smartBot.getBestMove(newPos)
	if botMove.From != botMove.To { // Valid move found
		newPos = engine.ApplyMove(newPos, botMove)

		// Check game status after bot's move
		status := newPos.GetGameStatus()
		gameStatus = status.String()
	}

	// Respond with the new FEN and game status
	c.JSON(http.StatusOK, MoveResponse{
		NewFEN:     newPos.String(),
		GameStatus: gameStatus,
	})
}
