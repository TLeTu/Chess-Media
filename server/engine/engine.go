package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Color represents the color of a chess piece.
type Color int

const (
	NoColor Color = iota
	White
	Black
)

func (c Color) String() string {
	switch c {
	case White:
		return "w"
	case Black:
		return "b"
	default:
		return "-"
	}
}

func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

// PieceType represents the type of a chess piece.
type PieceType int

const (
	NoPieceType PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

func (pt PieceType) String() string {
	switch pt {
	case Pawn:
		return "p"
	case Knight:
		return "n"
	case Bishop:
		return "b"
	case Rook:
		return "r"
	case Queen:
		return "q"
	case King:
		return "k"
	default:
		return ""
	}
}

// Piece represents a specific piece on the board.
type Piece int

const (
	Empty Piece = iota
	WhitePawn
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

func (p Piece) String() string {
	switch p {
	case WhitePawn:
		return "P"
	case WhiteKnight:
		return "N"
	case WhiteBishop:
		return "B"
	case WhiteRook:
		return "R"
	case WhiteQueen:
		return "Q"
	case WhiteKing:
		return "K"
	case BlackPawn:
		return "p"
	case BlackKnight:
		return "n"
	case BlackBishop:
		return "b"
	case BlackRook:
		return "r"
	case BlackQueen:
		return "q"
	case BlackKing:
		return "k"
	default:
		return " "
	}
}

func (p Piece) Color() Color {
	switch p {
	case WhitePawn, WhiteKnight, WhiteBishop, WhiteRook, WhiteQueen, WhiteKing:
		return White
	case BlackPawn, BlackKnight, BlackBishop, BlackRook, BlackQueen, BlackKing:
		return Black
	default:
		return NoColor
	}
}

func (p Piece) Type() PieceType {
	switch p {
	case WhitePawn, BlackPawn:
		return Pawn
	case WhiteKnight, BlackKnight:
		return Knight
	case WhiteBishop, BlackBishop:
		return Bishop
	case WhiteRook, BlackRook:
		return Rook
	case WhiteQueen, BlackQueen:
		return Queen
	case WhiteKing, BlackKing:
		return King
	default:
		return NoPieceType
	}
}

// Square represents a square on the chessboard (0-63).
type Square int

const (
	A1 Square = iota
	B1
	C1
	D1
	E1
	F1
	G1
	H1
	A2
	B2
	C2
	D2
	E2
	F2
	G2
	H2
	A3
	B3
	C3
	D3
	E3
	F3
	G3
	H3
	A4
	B4
	C4
	D4
	E4
	F4
	G4
	H4
	A5
	B5
	C5
	D5
	E5
	F5
	G5
	H5
	A6
	B6
	C6
	D6
	E6
	F6
	G6
	H6
	A7
	B7
	C7
	D7
	E7
	F7
	G7
	H7
	A8
	B8
	C8
	D8
	E8
	F8
	G8
	H8
	NoSquare Square = -1
)

func (s Square) String() string {
	if s == NoSquare {
		return "-"
	}
	file := string(rune('a' + (s % 8)))
	rank := strconv.Itoa(int(s/8 + 1))
	return file + rank
}

// Board represents the 8x8 chessboard.
type Board [64]Piece

// GameStatus represents the current state of the game
type GameStatus int

const (
	InProgress GameStatus = iota
	Checkmate
	Stalemate
	DrawByRepetition
	DrawByFiftyMoveRule
	DrawByInsufficientMaterial
)

func (gs GameStatus) String() string {
	switch gs {
	case InProgress:
		return "in_progress"
	case Checkmate:
		return "checkmate"
	case Stalemate:
		return "stalemate"
	case DrawByRepetition:
		return "draw_by_repetition"
	case DrawByFiftyMoveRule:
		return "draw_by_fifty_move_rule"
	case DrawByInsufficientMaterial:
		return "draw_by_insufficient_material"
	default:
		return "unknown"
	}
}

// Position encapsulates the entire state of the game.
type Position struct {
	Board          Board
	Turn           Color
	CastlingRights string // KQkq, KQk, etc.
	EnPassant      Square // NoSquare if no en passant square
	HalfMoveClock  int    // For 50-move rule
	FullMoveNumber int    // Increments after Black's move

	// Cache for performance optimization
	whiteKingPos Square
	blackKingPos Square
	positionHash uint64 // For threefold repetition detection
}

// NewGame creates a new game in the starting position.
func NewGame() *Position {
	pos := &Position{
		Board: Board{
			A1: WhiteRook, B1: WhiteKnight, C1: WhiteBishop, D1: WhiteQueen, E1: WhiteKing, F1: WhiteBishop, G1: WhiteKnight, H1: WhiteRook,
			A2: WhitePawn, B2: WhitePawn, C2: WhitePawn, D2: WhitePawn, E2: WhitePawn, F2: WhitePawn, G2: WhitePawn, H2: WhitePawn,
			A8: BlackRook, B8: BlackKnight, C8: BlackBishop, D8: BlackQueen, E8: BlackKing, F8: BlackBishop, G8: BlackKnight, H8: BlackRook,
			A7: BlackPawn, B7: BlackPawn, C7: BlackPawn, D7: BlackPawn, E7: BlackPawn, F7: BlackPawn, G7: BlackPawn, H7: BlackPawn,
		},
		Turn:           White,
		CastlingRights: "KQkq",
		EnPassant:      NoSquare,
		HalfMoveClock:  0,
		FullMoveNumber: 1,
	}

	// Initialize king positions
	pos.updateKingPositions()

	return pos
}

// updateKingPositions finds and caches the positions of both kings
func (pos *Position) updateKingPositions() {
	pos.whiteKingPos = NoSquare
	pos.blackKingPos = NoSquare

	for sq := A1; sq <= H8; sq++ {
		piece := pos.Board[sq]
		if piece.Type() == King {
			if piece.Color() == White {
				pos.whiteKingPos = sq
			} else if piece.Color() == Black {
				pos.blackKingPos = sq
			}
		}
	}
}

// ParseFEN parses a FEN string and returns a Position.
func ParseFEN(fen string) (*Position, error) {
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid FEN string: %s", fen)
	}

	pos := &Position{
		EnPassant: NoSquare, // Default to no en passant square
	}

	// Parse board
	boardStr := parts[0]
	rank := 7 // Start from 8th rank
	file := 0
	for _, r := range boardStr {
		switch {
		case r == '/':
			rank--
			file = 0
		case r >= '1' && r <= '8':
			file += int(r - '0')
		default:
			piece := Empty
			switch r {
			case 'P':
				piece = WhitePawn
			case 'N':
				piece = WhiteKnight
			case 'B':
				piece = WhiteBishop
			case 'R':
				piece = WhiteRook
			case 'Q':
				piece = WhiteQueen
			case 'K':
				piece = WhiteKing
			case 'p':
				piece = BlackPawn
			case 'n':
				piece = BlackKnight
			case 'b':
				piece = BlackBishop
			case 'r':
				piece = BlackRook
			case 'q':
				piece = BlackQueen
			case 'k':
				piece = BlackKing
			default:
				return nil, fmt.Errorf("invalid piece character in FEN: %c", r)
			}
			pos.Board[rank*8+file] = piece
			file++
		}
	}

	// Parse turn
	switch parts[1] {
	case "w":
		pos.Turn = White
	case "b":
		pos.Turn = Black
	default:
		return nil, fmt.Errorf("invalid turn in FEN: %s", parts[1])
	}

	// Parse castling rights
	if parts[2] == "-" {
		pos.CastlingRights = ""
	} else {
		// Validate castling rights
		validChars := "KQkq"
		for _, c := range parts[2] {
			if !strings.ContainsRune(validChars, c) {
				return nil, fmt.Errorf("invalid castling rights in FEN: %s", parts[2])
			}
		}
		pos.CastlingRights = parts[2]
	}

	// Parse en passant square
	if parts[3] != "-" {
		if len(parts[3]) != 2 {
			return nil, fmt.Errorf("invalid en passant square in FEN: %s", parts[3])
		}
		fileChar := parts[3][0]
		rankChar := parts[3][1]

		if fileChar < 'a' || fileChar > 'h' || rankChar < '1' || rankChar > '8' {
			return nil, fmt.Errorf("invalid en passant square in FEN: %s", parts[3])
		}

		pos.EnPassant = Square(int(rankChar-'1')*8 + int(fileChar-'a'))
	}

	// Parse half-move clock
	halfMove, err := strconv.Atoi(parts[4])
	if err != nil || halfMove < 0 {
		return nil, fmt.Errorf("invalid half-move clock in FEN: %s", parts[4])
	}
	pos.HalfMoveClock = halfMove

	// Parse full-move number
	fullMove, err := strconv.Atoi(parts[5])
	if err != nil || fullMove < 1 {
		return nil, fmt.Errorf("invalid full-move number in FEN: %s", parts[5])
	}
	pos.FullMoveNumber = fullMove

	// Update king positions
	pos.updateKingPositions()

	// Validate that both kings are present
	if pos.whiteKingPos == NoSquare || pos.blackKingPos == NoSquare {
		return nil, fmt.Errorf("invalid FEN: missing king(s)")
	}

	return pos, nil
}

// String returns the FEN string representation of the Position.
func (p *Position) String() string {
	var boardStr strings.Builder
	for rank := 7; rank >= 0; rank-- {
		emptyCount := 0
		for file := 0; file < 8; file++ {
			piece := p.Board[rank*8+file]
			if piece == Empty {
				emptyCount++
			} else {
				if emptyCount > 0 {
					boardStr.WriteString(strconv.Itoa(emptyCount))
					emptyCount = 0
				}
				boardStr.WriteString(piece.String())
			}
		}
		if emptyCount > 0 {
			boardStr.WriteString(strconv.Itoa(emptyCount))
		}
		if rank > 0 {
			boardStr.WriteString("/")
		}
	}

	enPassantStr := "-"
	if p.EnPassant != NoSquare {
		enPassantStr = p.EnPassant.String()
	}

	castlingRights := p.CastlingRights
	if castlingRights == "" {
		castlingRights = "-"
	}

	return fmt.Sprintf("%s %s %s %s %d %d",
		boardStr.String(),
		p.Turn.String(),
		castlingRights,
		enPassantStr,
		p.HalfMoveClock,
		p.FullMoveNumber,
	)
}

// Move represents a move from a source square to a destination square.
type Move struct {
	From        Square
	To          Square
	Promotion   PieceType // Only for pawn promotions
	IsCapture   bool
	IsCastling  bool
	IsEnPassant bool
}

func (m Move) String() string {
	s := m.From.String() + m.To.String()
	if m.Promotion != NoPieceType {
		s += m.Promotion.String()
	}
	return s
}

// GenerateLegalMoves generates all legal moves for the current position.
func (pos *Position) GenerateLegalMoves() []Move {
	moves := []Move{}

	for sq := A1; sq <= H8; sq++ {
		piece := pos.Board[sq]
		if piece == Empty || piece.Color() != pos.Turn {
			continue
		}

		switch piece.Type() {
		case Pawn:
			moves = append(moves, generatePawnMoves(pos, sq)...)
		case Knight:
			moves = append(moves, generateKnightMoves(pos, sq)...)
		case Bishop:
			moves = append(moves, generateBishopMoves(pos, sq)...)
		case Rook:
			moves = append(moves, generateRookMoves(pos, sq)...)
		case Queen:
			moves = append(moves, generateQueenMoves(pos, sq)...)
		case King:
			moves = append(moves, generateKingMoves(pos, sq)...)
		}
	}

	// Filter out moves that leave the king in check
	legalMoves := []Move{}
	for _, move := range moves {
		newPos := ApplyMove(pos, move)
		if !IsKingInCheck(newPos, pos.Turn) {
			legalMoves = append(legalMoves, move)
		}
	}

	return legalMoves
}

// ApplyMove applies a move to the position and returns a new position.
func ApplyMove(pos *Position, move Move) *Position {
	// Create a new position with copied board
	newPos := &Position{
		Turn:           oppositeColor(pos.Turn),
		CastlingRights: pos.CastlingRights,
		EnPassant:      NoSquare,              // Default to no en passant square
		HalfMoveClock:  pos.HalfMoveClock + 1, // Increment half-move clock
		FullMoveNumber: pos.FullMoveNumber,
		whiteKingPos:   pos.whiteKingPos,
		blackKingPos:   pos.blackKingPos,
	}

	// Copy the board
	copy(newPos.Board[:], pos.Board[:])

	// Make the move
	movingPiece := newPos.Board[move.From]
	capturedPiece := newPos.Board[move.To] // Store captured piece (if any)

	// Update the board
	newPos.Board[move.To] = movingPiece
	newPos.Board[move.From] = Empty

	// Set the IsCapture flag if there was a piece captured
	if capturedPiece != Empty {
		move.IsCapture = true
	}

	// Handle pawn promotion
	if move.Promotion != NoPieceType {
		color := movingPiece.Color()
		switch move.Promotion {
		case Queen:
			if color == White {
				newPos.Board[move.To] = WhiteQueen
			} else {
				newPos.Board[move.To] = BlackQueen
			}
		case Rook:
			if color == White {
				newPos.Board[move.To] = WhiteRook
			} else {
				newPos.Board[move.To] = BlackRook
			}
		case Bishop:
			if color == White {
				newPos.Board[move.To] = WhiteBishop
			} else {
				newPos.Board[move.To] = BlackBishop
			}
		case Knight:
			if color == White {
				newPos.Board[move.To] = WhiteKnight
			} else {
				newPos.Board[move.To] = BlackKnight
			}
		}
	}

	// Handle en passant capture
	if move.IsEnPassant {
		if pos.Turn == White {
			newPos.Board[move.To-8] = Empty // Captured black pawn
		} else {
			newPos.Board[move.To+8] = Empty // Captured white pawn
		}
	}

	// Update en passant square for next turn
	if movingPiece.Type() == Pawn && abs(int(move.From)-int(move.To)) == 16 {
		if pos.Turn == White {
			newPos.EnPassant = move.From + 8
		} else {
			newPos.EnPassant = move.From - 8
		}
	}

	// Handle castling
	if move.IsCastling {
		switch move.To {
		case G1: // White King-side castling
			newPos.Board[F1] = newPos.Board[H1]
			newPos.Board[H1] = Empty
		case C1: // White Queen-side castling
			newPos.Board[D1] = newPos.Board[A1]
			newPos.Board[A1] = Empty
		case G8: // Black King-side castling
			newPos.Board[F8] = newPos.Board[H8]
			newPos.Board[H8] = Empty
		case C8: // Black Queen-side castling
			newPos.Board[D8] = newPos.Board[A8]
			newPos.Board[A8] = Empty
		}
	}

	// Update castling rights
	newCastlingRights := pos.CastlingRights

	// If king moves, remove both castling rights for that color
	if movingPiece.Type() == King {
		if pos.Turn == White {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "K", "")
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "Q", "")
			newPos.whiteKingPos = move.To // Update king position
		} else {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "k", "")
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "q", "")
			newPos.blackKingPos = move.To // Update king position
		}
	}

	// If rook moves or is captured, remove the corresponding castling right
	if move.From == A1 || move.To == A1 {
		newCastlingRights = strings.ReplaceAll(newCastlingRights, "Q", "")
	}
	if move.From == H1 || move.To == H1 {
		newCastlingRights = strings.ReplaceAll(newCastlingRights, "K", "")
	}
	if move.From == A8 || move.To == A8 {
		newCastlingRights = strings.ReplaceAll(newCastlingRights, "q", "")
	}
	if move.From == H8 || move.To == H8 {
		newCastlingRights = strings.ReplaceAll(newCastlingRights, "k", "")
	}

	newPos.CastlingRights = newCastlingRights

	// Reset half-move clock on pawn move or capture
	if movingPiece.Type() == Pawn || move.IsCapture {
		newPos.HalfMoveClock = 0
	}

	// Update full-move number
	if pos.Turn == Black {
		newPos.FullMoveNumber++
	}

	return newPos
}

// IsKingInCheck checks if the king of the given color is in check.
func IsKingInCheck(pos *Position, color Color) bool {
	// Get the king's position from cache
	kingSquare := pos.whiteKingPos
	if color == Black {
		kingSquare = pos.blackKingPos
	}

	if kingSquare == NoSquare {
		return false // Should not happen in a valid game
	}

	// Check for attacks from all opponent's pieces
	opponentColor := oppositeColor(color)

	// Pawn attacks
	if isPawnAttacking(pos, kingSquare, opponentColor) {
		return true
	}

	// Knight attacks
	if isKnightAttacking(pos, kingSquare, opponentColor) {
		return true
	}

	// Bishop and Queen (diagonal) attacks
	if isSliderAttacking(pos, kingSquare, opponentColor, Bishop) {
		return true
	}

	// Rook and Queen (straight) attacks
	if isSliderAttacking(pos, kingSquare, opponentColor, Rook) {
		return true
	}

	// King attacks (should not happen in a legal position, but for completeness)
	if isKingAttacking(pos, kingSquare, opponentColor) {
		return true
	}

	return false
}

// GetGameStatus returns the current status of the game
func (pos *Position) GetGameStatus() GameStatus {
	// Check for checkmate or stalemate
	legalMoves := pos.GenerateLegalMoves()
	if len(legalMoves) == 0 {
		if IsKingInCheck(pos, pos.Turn) {
			return Checkmate
		}
		return Stalemate
	}

	// Check for 50-move rule
	if pos.HalfMoveClock >= 100 { // 50 moves = 100 half-moves
		return DrawByFiftyMoveRule
	}

	// Check for insufficient material
	if hasInsufficientMaterial(pos) {
		return DrawByInsufficientMaterial
	}

	// Game is still in progress
	return InProgress
}

// hasInsufficientMaterial checks if there is insufficient material for checkmate
func hasInsufficientMaterial(pos *Position) bool {
	// Count pieces
	whitePieces := 0
	blackPieces := 0
	whiteKnights := 0
	blackKnights := 0
	whiteBishops := 0
	blackBishops := 0
	whiteBishopSquareColor := -1 // -1 = not set, 0 = light square, 1 = dark square
	blackBishopSquareColor := -1

	for sq := A1; sq <= H8; sq++ {
		piece := pos.Board[sq]
		if piece == Empty {
			continue
		}

		switch piece {
		case WhitePawn, WhiteRook, WhiteQueen:
			return false // White has material for checkmate
		case BlackPawn, BlackRook, BlackQueen:
			return false // Black has material for checkmate
		case WhiteKnight:
			whiteKnights++
			whitePieces++
		case BlackKnight:
			blackKnights++
			blackPieces++
		case WhiteBishop:
			whiteBishops++
			whitePieces++
			// Determine bishop's square color (light or dark)
			squareColor := (int(sq/8) + int(sq%8)) % 2
			if whiteBishopSquareColor == -1 {
				whiteBishopSquareColor = squareColor
			} else if whiteBishopSquareColor != squareColor {
				return false // Bishops on different colored squares can checkmate
			}
		case BlackBishop:
			blackBishops++
			blackPieces++
			// Determine bishop's square color
			squareColor := (int(sq/8) + int(sq%8)) % 2
			if blackBishopSquareColor == -1 {
				blackBishopSquareColor = squareColor
			} else if blackBishopSquareColor != squareColor {
				return false // Bishops on different colored squares can checkmate
			}
		case WhiteKing:
			whitePieces++
		case BlackKing:
			blackPieces++
		}
	}

	// King vs King
	if whitePieces == 1 && blackPieces == 1 {
		return true
	}

	// King + Bishop vs King or King + Knight vs King
	if (whitePieces == 2 && blackPieces == 1 && (whiteBishops == 1 || whiteKnights == 1)) ||
		(blackPieces == 2 && whitePieces == 1 && (blackBishops == 1 || blackKnights == 1)) {
		return true
	}

	// King + 2 Knights vs King (technically can checkmate but practically a draw)
	if (whitePieces == 3 && blackPieces == 1 && whiteKnights == 2) ||
		(blackPieces == 3 && whitePieces == 1 && blackKnights == 2) {
		return true
	}

	// King + Bishop vs King + Bishop (same colored bishops)
	if whitePieces == 2 && blackPieces == 2 && whiteBishops == 1 && blackBishops == 1 &&
		whiteBishopSquareColor == blackBishopSquareColor {
		return true
	}

	return false
}

// Helper functions for move generation
func generatePawnMoves(pos *Position, sq Square) []Move {
	moves := []Move{}
	piece := pos.Board[sq]
	direction := 0
	startRank := 0
	promotionRank := 0

	if piece.Color() == White {
		direction = 8     // Move up one rank
		startRank = 1     // Pawns start on 2nd rank (index 1)
		promotionRank = 7 // Promote on 8th rank (index 7)
	} else {
		direction = -8    // Move down one rank
		startRank = 6     // Pawns start on 7th rank (index 6)
		promotionRank = 0 // Promote on 1st rank (index 0)
	}

	// Single push
	targetSq := sq + Square(direction)
	if targetSq >= 0 && targetSq < 64 && pos.Board[targetSq] == Empty {
		if (targetSq / 8) == Square(promotionRank) {
			// Pawn promotion
			moves = append(moves, Move{From: sq, To: targetSq, Promotion: Queen})
			moves = append(moves, Move{From: sq, To: targetSq, Promotion: Rook})
			moves = append(moves, Move{From: sq, To: targetSq, Promotion: Bishop})
			moves = append(moves, Move{From: sq, To: targetSq, Promotion: Knight})
		} else {
			moves = append(moves, Move{From: sq, To: targetSq})
		}

		// Double push from starting rank
		if (sq / 8) == Square(startRank) {
			doublePushTargetSq := sq + Square(direction*2)
			if doublePushTargetSq >= 0 && doublePushTargetSq < 64 && pos.Board[doublePushTargetSq] == Empty {
				moves = append(moves, Move{From: sq, To: doublePushTargetSq})
			}
		}
	}

	// Captures (including en passant)
	for _, offset := range []int{-1, 1} {
		captureTargetSq := sq + Square(direction+offset)
		if captureTargetSq >= 0 && captureTargetSq < 64 &&
			abs(int(captureTargetSq%8)-int(sq%8)) == 1 { // Ensure we don't wrap around the board

			// Normal capture
			capturedPiece := pos.Board[captureTargetSq]
			if capturedPiece != Empty && capturedPiece.Color() != piece.Color() {
				if (captureTargetSq / 8) == Square(promotionRank) {
					// Capture with promotion
					moves = append(moves, Move{From: sq, To: captureTargetSq, Promotion: Queen, IsCapture: true})
					moves = append(moves, Move{From: sq, To: captureTargetSq, Promotion: Rook, IsCapture: true})
					moves = append(moves, Move{From: sq, To: captureTargetSq, Promotion: Bishop, IsCapture: true})
					moves = append(moves, Move{From: sq, To: captureTargetSq, Promotion: Knight, IsCapture: true})
				} else {
					moves = append(moves, Move{From: sq, To: captureTargetSq, IsCapture: true})
				}
			}

			// En passant capture
			if captureTargetSq == pos.EnPassant && pos.EnPassant != NoSquare {
				moves = append(moves, Move{From: sq, To: captureTargetSq, IsCapture: true, IsEnPassant: true})
			}
		}
	}

	return moves
}

func generateKnightMoves(pos *Position, sq Square) []Move {
	moves := []Move{}
	piece := pos.Board[sq]

	// Knight's L-shaped moves (row_offset, col_offset)
	knightMoves := []struct{ row, col int }{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for _, km := range knightMoves {
		targetRow := int(sq/8) + km.row
		targetCol := int(sq%8) + km.col

		if targetRow >= 0 && targetRow < 8 && targetCol >= 0 && targetCol < 8 {
			targetSq := Square(targetRow*8 + targetCol)
			targetPiece := pos.Board[targetSq]

			if targetPiece == Empty || targetPiece.Color() != piece.Color() {
				move := Move{From: sq, To: targetSq}
				if targetPiece != Empty {
					move.IsCapture = true
				}
				moves = append(moves, move)
			}
		}
	}
	return moves
}

// generateSliderMoves is a helper for Bishop, Rook, and Queen moves.
func generateSliderMoves(pos *Position, sq Square, deltas [][2]int) []Move {
	moves := []Move{}
	piece := pos.Board[sq]

	for _, delta := range deltas {
		for i := 1; i < 8; i++ { // Max 7 squares in any direction
			targetRow := int(sq/8) + delta[0]*i
			targetCol := int(sq%8) + delta[1]*i

			if targetRow < 0 || targetRow >= 8 || targetCol < 0 || targetCol >= 8 {
				break // Off board
			}

			targetSq := Square(targetRow*8 + targetCol)
			targetPiece := pos.Board[targetSq]

			if targetPiece == Empty {
				moves = append(moves, Move{From: sq, To: targetSq})
			} else {
				if targetPiece.Color() != piece.Color() {
					moves = append(moves, Move{From: sq, To: targetSq, IsCapture: true})
				}
				break // Blocked by own piece or captured opponent's piece
			}
		}
	}
	return moves
}

func generateBishopMoves(pos *Position, sq Square) []Move {
	// Diagonal deltas
	deltas := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	return generateSliderMoves(pos, sq, deltas)
}

func generateRookMoves(pos *Position, sq Square) []Move {
	// Straight deltas
	deltas := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	return generateSliderMoves(pos, sq, deltas)
}

func generateQueenMoves(pos *Position, sq Square) []Move {
	// Queen moves are a combination of Rook and Bishop moves
	deltas := [][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1}, // Rook moves
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1}, // Bishop moves
	}
	return generateSliderMoves(pos, sq, deltas)
}

func generateKingMoves(pos *Position, sq Square) []Move {
	moves := []Move{}
	piece := pos.Board[sq]

	// King moves one square in any direction
	kingMoves := []struct{ row, col int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, km := range kingMoves {
		targetRow := int(sq/8) + km.row
		targetCol := int(sq%8) + km.col

		if targetRow >= 0 && targetRow < 8 && targetCol >= 0 && targetCol < 8 {
			targetSq := Square(targetRow*8 + targetCol)
			targetPiece := pos.Board[targetSq]

			if targetPiece == Empty || targetPiece.Color() != piece.Color() {
				move := Move{From: sq, To: targetSq}
				if targetPiece != Empty {
					move.IsCapture = true
				}
				moves = append(moves, move)
			}
		}
	}

	// Castling moves
	if pos.Turn == White && sq == E1 {
		// King-side castling
		if strings.Contains(pos.CastlingRights, "K") &&
			pos.Board[F1] == Empty && pos.Board[G1] == Empty &&
			!IsKingInCheck(pos, White) {
			// Check if squares are not under attack
			tempPos1 := ApplyMove(pos, Move{From: E1, To: F1})
			tempPos2 := ApplyMove(pos, Move{From: E1, To: G1})
			if !IsKingInCheck(tempPos1, White) && !IsKingInCheck(tempPos2, White) {
				moves = append(moves, Move{From: E1, To: G1, IsCastling: true})
			}
		}

		// Queen-side castling
		if strings.Contains(pos.CastlingRights, "Q") &&
			pos.Board[D1] == Empty && pos.Board[C1] == Empty && pos.Board[B1] == Empty &&
			!IsKingInCheck(pos, White) {
			// Check if squares are not under attack
			tempPos1 := ApplyMove(pos, Move{From: E1, To: D1})
			tempPos2 := ApplyMove(pos, Move{From: E1, To: C1})
			if !IsKingInCheck(tempPos1, White) && !IsKingInCheck(tempPos2, White) {
				moves = append(moves, Move{From: E1, To: C1, IsCastling: true})
			}
		}
	} else if pos.Turn == Black && sq == E8 {
		// King-side castling
		if strings.Contains(pos.CastlingRights, "k") &&
			pos.Board[F8] == Empty && pos.Board[G8] == Empty &&
			!IsKingInCheck(pos, Black) {
			// Check if squares are not under attack
			tempPos1 := ApplyMove(pos, Move{From: E8, To: F8})
			tempPos2 := ApplyMove(pos, Move{From: E8, To: G8})
			if !IsKingInCheck(tempPos1, Black) && !IsKingInCheck(tempPos2, Black) {
				moves = append(moves, Move{From: E8, To: G8, IsCastling: true})
			}
		}

		// Queen-side castling
		if strings.Contains(pos.CastlingRights, "q") &&
			pos.Board[D8] == Empty && pos.Board[C8] == Empty && pos.Board[B8] == Empty &&
			!IsKingInCheck(pos, Black) {
			// Check if squares are not under attack
			tempPos1 := ApplyMove(pos, Move{From: E8, To: D8})
			tempPos2 := ApplyMove(pos, Move{From: E8, To: C8})
			if !IsKingInCheck(tempPos1, Black) && !IsKingInCheck(tempPos2, Black) {
				moves = append(moves, Move{From: E8, To: C8, IsCastling: true})
			}
		}
	}

	return moves
}

// Helper functions for IsKingInCheck
func isPawnAttacking(pos *Position, kingSquare Square, attackerColor Color) bool {
	kingRow, kingCol := int(kingSquare/8), int(kingSquare%8)

	// Direction from attacker's perspective to the king
	direction := 0
	if attackerColor == White {
		direction = 1 // White pawns attack upwards (from rank 1-8 perspective)
	} else {
		direction = -1 // Black pawns attack downwards (from rank 8-1 perspective)
	}

	// Check diagonal attacks - pawns attack diagonally forward
	attackOffsets := []int{-1, 1}
	for _, offset := range attackOffsets {
		// Look for pawns that could attack the king square
		attackerRow, attackerCol := kingRow-direction, kingCol+offset

		if attackerRow >= 0 && attackerRow < 8 && attackerCol >= 0 && attackerCol < 8 {
			attackerPiece := pos.Board[Square(attackerRow*8+attackerCol)]
			if attackerPiece.Type() == Pawn && attackerPiece.Color() == attackerColor {
				return true
			}
		}
	}
	return false
}

func isKnightAttacking(pos *Position, kingSquare Square, attackerColor Color) bool {
	kingRow, kingCol := int(kingSquare/8), int(kingSquare%8)

	knightMoves := []struct{ row, col int }{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}

	for _, km := range knightMoves {
		targetRow, targetCol := kingRow+km.row, kingCol+km.col

		if targetRow >= 0 && targetRow < 8 && targetCol >= 0 && targetCol < 8 {
			attackerPiece := pos.Board[Square(targetRow*8+targetCol)]
			if attackerPiece.Type() == Knight && attackerPiece.Color() == attackerColor {
				return true
			}
		}
	}
	return false
}

func isSliderAttacking(pos *Position, kingSquare Square, attackerColor Color, sliderType PieceType) bool {
	kingRow, kingCol := int(kingSquare/8), int(kingSquare%8)

	deltas := [][2]int{}
	switch sliderType {
	case Bishop:
		deltas = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} // Diagonal
	case Rook:
		deltas = [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} // Straight
	}

	for _, delta := range deltas {
		for i := 1; i < 8; i++ {
			targetRow, targetCol := kingRow+delta[0]*i, kingCol+delta[1]*i

			if targetRow < 0 || targetRow >= 8 || targetCol < 0 || targetCol >= 8 {
				break // Off board
			}

			targetSq := Square(targetRow*8 + targetCol)
			piece := pos.Board[targetSq]

			if piece != Empty {
				if piece.Color() == attackerColor &&
					(piece.Type() == sliderType ||
						(sliderType == Bishop && piece.Type() == Queen) ||
						(sliderType == Rook && piece.Type() == Queen)) {
					return true
				}
				break // Blocked by any piece
			}
		}
	}
	return false
}

func isKingAttacking(pos *Position, kingSquare Square, attackerColor Color) bool {
	kingRow, kingCol := int(kingSquare/8), int(kingSquare%8)

	kingMoves := []struct{ row, col int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, km := range kingMoves {
		targetRow, targetCol := kingRow+km.row, kingCol+km.col

		if targetRow >= 0 && targetRow < 8 && targetCol >= 0 && targetCol < 8 {
			attackerPiece := pos.Board[Square(targetRow*8+targetCol)]
			if attackerPiece.Type() == King && attackerPiece.Color() == attackerColor {
				return true
			}
		}
	}
	return false
}

// FindMove finds a move in the list of legal moves that matches the given from and to squares
func (pos *Position) FindMove(from, to Square, promotionPiece PieceType) (Move, bool) {
	legalMoves := pos.GenerateLegalMoves()

	for _, move := range legalMoves {
		if move.From == from && move.To == to {
			// For non-promotion moves or if promotion piece matches
			if (move.Promotion == NoPieceType && promotionPiece == NoPieceType) ||
				(move.Promotion == promotionPiece) {
				return move, true
			}
		}
	}

	return Move{}, false
}

// ParseMove parses a move string in algebraic notation (e.g., "e2e4")
func ParseMove(pos *Position, moveStr string) (Move, error) {
	if len(moveStr) < 4 {
		return Move{}, fmt.Errorf("invalid move format: %s", moveStr)
	}

	fromFile := moveStr[0] - 'a'
	fromRank := moveStr[1] - '1'
	toFile := moveStr[2] - 'a'
	toRank := moveStr[3] - '1'

	if fromFile < 0 || fromFile > 7 || fromRank < 0 || fromRank > 7 ||
		toFile < 0 || toFile > 7 || toRank < 0 || toRank > 7 {
		return Move{}, fmt.Errorf("invalid move coordinates: %s", moveStr)
	}

	from := Square(int(fromRank)*8 + int(fromFile))
	to := Square(int(toRank)*8 + int(toFile))

	// Check for promotion
	promotionPiece := NoPieceType
	if len(moveStr) > 4 {
		switch moveStr[4] {
		case 'q':
			promotionPiece = Queen
		case 'r':
			promotionPiece = Rook
		case 'b':
			promotionPiece = Bishop
		case 'n':
			promotionPiece = Knight
		default:
			return Move{}, fmt.Errorf("invalid promotion piece: %c", moveStr[4])
		}
	}

	// Find the move in legal moves
	move, found := pos.FindMove(from, to, promotionPiece)
	if !found {
		return Move{}, fmt.Errorf("illegal move: %s", moveStr)
	}

	return move, nil
}

// Utility functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func oppositeColor(c Color) Color {
	if c == White {
		return Black
	}
	return White
}

// InitBoard creates a new game with the starting position.
func InitBoard() *Position {
	return NewGame()
}
