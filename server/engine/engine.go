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
)

func (s Square) String() string {
	file := string(rune('a' + (s % 8)))
	rank := strconv.Itoa(int(s/8 + 1))
	return file + rank
}

// Board represents the 8x8 chessboard.
type Board [64]Piece

// Position encapsulates the entire state of the game.
type Position struct {
	Board          Board
	Turn           Color
	CastlingRights string // KQkq, KQk, etc.
	EnPassant      Square // -1 if no en passant square
	HalfMoveClock  int    // For 50-move rule
	FullMoveNumber int    // Increments after Black's move
}

// NewGame creates a new game in the starting position.
func NewGame() *Position {
	return &Position{
		Board: Board{
			A1: WhiteRook, B1: WhiteKnight, C1: WhiteBishop, D1: WhiteQueen, E1: WhiteKing, F1: WhiteBishop, G1: WhiteKnight, H1: WhiteRook,
			A2: WhitePawn, B2: WhitePawn, C2: WhitePawn, D2: WhitePawn, E2: WhitePawn, F2: WhitePawn, G2: WhitePawn, H2: WhitePawn,
			A8: BlackRook, B8: BlackKnight, C8: BlackBishop, D8: BlackQueen, E8: BlackKing, F8: BlackBishop, G8: BlackKnight, H8: BlackRook,
			A7: BlackPawn, B7: BlackPawn, C7: BlackPawn, D7: BlackPawn, E7: BlackPawn, F7: BlackPawn, G7: BlackPawn, H7: BlackPawn,
		},
		Turn:           White,
		CastlingRights: "KQkq",
		EnPassant:      -1, // No en passant square initially
		HalfMoveClock:  0,
		FullMoveNumber: 1,
	}
}

// ParseFEN parses a FEN string and returns a Position.
func ParseFEN(fen string) (*Position, error) {
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		return nil, fmt.Errorf("invalid FEN string: %s", fen)
	}

	pos := &Position{}

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
	pos.CastlingRights = parts[2]

	// Parse en passant square
	if parts[3] != "-" {
		fileChar := parts[3][0]
		rankChar := parts[3][1]
		pos.EnPassant = Square(int(rankChar-'1')*8 + int(fileChar-'a'))
	} else {
		pos.EnPassant = -1
	}

	// Parse half-move clock
	halfMove, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid half-move clock in FEN: %s", parts[4])
	}
	pos.HalfMoveClock = halfMove

	// Parse full-move number
	fullMove, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid full-move number in FEN: %s", parts[5])
	}
	pos.FullMoveNumber = fullMove

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
	if p.EnPassant != -1 {
		enPassantStr = p.EnPassant.String()
	}

	return fmt.Sprintf("%s %s %s %s %d %d",
		boardStr.String(),
		p.Turn.String(),
		p.CastlingRights,
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
	newBoard := pos.Board
	newBoard[move.To] = newBoard[move.From]
	newBoard[move.From] = Empty

	// Handle pawn promotion
	if move.Promotion != NoPieceType {
		color := newBoard[move.To].Color()
		switch move.Promotion {
		case Queen:
			if color == White {
				newBoard[move.To] = WhiteQueen
			} else {
				newBoard[move.To] = BlackQueen
			}
		case Rook:
			if color == White {
				newBoard[move.To] = WhiteRook
			} else {
				newBoard[move.To] = BlackRook
			}
		case Bishop:
			if color == White {
				newBoard[move.To] = WhiteBishop
			} else {
				newBoard[move.To] = BlackBishop
			}
		case Knight:
			if color == White {
				newBoard[move.To] = WhiteKnight
			} else {
				newBoard[move.To] = BlackKnight
			}
		}
	}

	// Handle en passant capture
	if move.IsEnPassant {
		if pos.Turn == White {
			newBoard[move.To-8] = Empty // Captured black pawn
		} else {
			newBoard[move.To+8] = Empty // Captured white pawn
		}
	}

	// Update en passant square for next turn
	newEnPassant := Square(-1)
	if newBoard[move.To].Type() == Pawn && abs(int(move.From)-int(move.To)) == 16 {
		if pos.Turn == White {
			newEnPassant = move.To - 8
		} else {
			newEnPassant = move.To + 8
		}
	}

	// Handle castling
	if move.IsCastling {

		if move.To == G1 { // White King-side castling
			newBoard[F1] = newBoard[H1]
			newBoard[H1] = Empty
		} else if move.To == C1 { // White Queen-side castling
			newBoard[D1] = newBoard[A1]
			newBoard[A1] = Empty
		} else if move.To == G8 { // Black King-side castling
			newBoard[F8] = newBoard[H8]
			newBoard[H8] = Empty
		} else if move.To == C8 { // Black Queen-side castling
			newBoard[D8] = newBoard[A8]
			newBoard[A8] = Empty
		}
	}

	// Update castling rights
	newCastlingRights := pos.CastlingRights
	if pos.Turn == White {
		if move.From == E1 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "K", "")
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "Q", "")
		} else if move.From == A1 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "Q", "")
		} else if move.From == H1 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "K", "")
		}
	} else { // Black
		if move.From == E8 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "k", "")
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "q", "")
		} else if move.From == A8 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "q", "")
		} else if move.From == H8 {
			newCastlingRights = strings.ReplaceAll(newCastlingRights, "k", "")
		}
	}
	if newCastlingRights == "" {
		newCastlingRights = "-"
	}

	// Update half-move clock
	newHalfMoveClock := pos.HalfMoveClock + 1
	if move.IsCapture || newBoard[move.To].Type() == Pawn {
		newHalfMoveClock = 0
	}

	// Update full-move number
	newFullMoveNumber := pos.FullMoveNumber
	if pos.Turn == Black {
		newFullMoveNumber++
	}

	return &Position{
		Board:          newBoard,
		Turn:           oppositeColor(pos.Turn),
		CastlingRights: newCastlingRights,
		EnPassant:      newEnPassant,
		HalfMoveClock:  newHalfMoveClock,
		FullMoveNumber: newFullMoveNumber,
	}
}

// IsKingInCheck checks if the king of the given color is in check.
func IsKingInCheck(pos *Position, color Color) bool {
	kingSquare := Square(-1)
	for sq := A1; sq <= H8; sq++ {
		piece := pos.Board[sq]
		if piece.Type() == King && piece.Color() == color {
			kingSquare = sq
			break
		}
	}

	if kingSquare == -1 {
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

// Helper functions for move generation (to be implemented later)
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
	if pos.Board[targetSq] == Empty {
		moves = append(moves, Move{From: sq, To: targetSq})

		// Double push
		if (sq / 8) == Square(startRank) {
			doublePushTargetSq := sq + Square(direction*2)
			if pos.Board[doublePushTargetSq] == Empty {
				moves = append(moves, Move{From: sq, To: doublePushTargetSq})
			}
		}
	}

	// Captures
	captureTargets := []Square{sq + Square(direction-1), sq + Square(direction+1)}
	for _, target := range captureTargets {
		if target < 0 || target > 63 || (abs(int(target%8)-int(sq%8)) != 1) {
			continue // Check if target is off board or not diagonal
		}
		capturedPiece := pos.Board[target]
		if capturedPiece != Empty && capturedPiece.Color() != piece.Color() {
			moves = append(moves, Move{From: sq, To: target, IsCapture: true})
		}
		// En passant
		if target == pos.EnPassant && pos.EnPassant != -1 {
			moves = append(moves, Move{From: sq, To: target, IsCapture: true, IsEnPassant: true})
		}
	}

	// Handle promotions
	finalMoves := []Move{}
	for _, move := range moves {
		if (move.To / 8) == Square(promotionRank) {
			finalMoves = append(finalMoves, Move{From: move.From, To: move.To, Promotion: Queen, IsCapture: move.IsCapture})
			finalMoves = append(finalMoves, Move{From: move.From, To: move.To, Promotion: Rook, IsCapture: move.IsCapture})
			finalMoves = append(finalMoves, Move{From: move.From, To: move.To, Promotion: Bishop, IsCapture: move.IsCapture})
			finalMoves = append(finalMoves, Move{From: move.From, To: move.To, Promotion: Knight, IsCapture: move.IsCapture})
		} else {
			finalMoves = append(finalMoves, move)
		}
	}

	return finalMoves
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
	// King-side castling
	if (pos.Turn == White && sq == E1 && strings.Contains(pos.CastlingRights, "K")) &&
		pos.Board[F1] == Empty && pos.Board[G1] == Empty &&
		!IsKingInCheck(pos, White) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E1, To: F1}), White) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E1, To: G1}), White) {
		moves = append(moves, Move{From: E1, To: G1, IsCastling: true})
	}
	if (pos.Turn == Black && sq == E8 && strings.Contains(pos.CastlingRights, "k")) &&
		pos.Board[F8] == Empty && pos.Board[G8] == Empty &&
		!IsKingInCheck(pos, Black) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E8, To: F8}), Black) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E8, To: G8}), Black) {
		moves = append(moves, Move{From: E8, To: G8, IsCastling: true})
	}

	// Queen-side castling
	if (pos.Turn == White && sq == E1 && strings.Contains(pos.CastlingRights, "Q")) &&
		pos.Board[B1] == Empty && pos.Board[C1] == Empty && pos.Board[D1] == Empty &&
		!IsKingInCheck(pos, White) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E1, To: D1}), White) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E1, To: C1}), White) {
		moves = append(moves, Move{From: E1, To: C1, IsCastling: true})
	}
	if (pos.Turn == Black && sq == E8 && strings.Contains(pos.CastlingRights, "q")) &&
		pos.Board[B8] == Empty && pos.Board[C8] == Empty && pos.Board[D8] == Empty &&
		!IsKingInCheck(pos, Black) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E8, To: D8}), Black) &&
		!IsKingInCheck(ApplyMove(pos, Move{From: E8, To: C8}), Black) {
		moves = append(moves, Move{From: E8, To: C8, IsCastling: true})
	}

	return moves
}

// Helper functions for IsKingInCheck (to be implemented later)
func isPawnAttacking(pos *Position, kingSquare Square, attackerColor Color) bool {
	kingRow, kingCol := int(kingSquare/8), int(kingSquare%8)

	direction := 0
	if attackerColor == White {
		direction = -1 // White pawns attack downwards relative to their perspective
	} else {
		direction = 1 // Black pawns attack upwards relative to their perspective
	}

	// Check diagonal attacks
	attackOffsets := []int{-1, 1}
	for _, offset := range attackOffsets {
		targetRow, targetCol := kingRow+direction, kingCol+offset

		if targetRow >= 0 && targetRow < 8 && targetCol >= 0 && targetCol < 8 {
			attackerPiece := pos.Board[Square(targetRow*8+targetCol)]
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
	case Queen:
		deltas = [][2]int{
			{-1, -1}, {-1, 0}, {-1, 1},
			{0, -1}, {0, 1},
			{1, -1}, {1, 0}, {1, 1}, // All 8 directions
		}
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
				if piece.Color() == attackerColor && (piece.Type() == sliderType || (sliderType == Bishop && piece.Type() == Queen) || (sliderType == Rook && piece.Type() == Queen)) {
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

// InitBoard is a placeholder for now, will be implemented later.
func InitBoard() *Position {
	return NewGame()
}
