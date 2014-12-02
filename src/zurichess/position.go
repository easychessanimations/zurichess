package main

import (
	"fmt"
	"log"
)

// Square identifies the location on the board.
type Square int

func RankFile(r, f int) Square {
	return Square(r*8 + f)
}

func SquareFromString(s string) Square {
	r := int(s[1] - '1')
	f := int(s[0] - 'a')
	return RankFile(r, f)
}

func (sq Square) Bitboard() Bitboard {
	return 1 << uint(sq)
}

// Rank returns a number 0...7 representing the rank of the square.
func (sq Square) Rank() int {
	return int(sq / 8)
}

// File returns a number 0...7 representing the file of the square.
func (sq Square) File() int {
	return int(sq % 8)
}

func (sq Square) String() string {
	return string([]byte{
		uint8(sq.File() + 'a'),
		uint8(sq.Rank() + '1'),
	})
}

// PieceType represents a colorless piece
type PieceType uint

func (pt PieceType) String() string {
	switch pt {
	case NoPieceType:
		return "(nopiecetype)"
	case Pawn:
		return "Pawn"
	case Knight:
		return "Knight"
	case Bishop:
		return "Bishop"
	case Rock:
		return "Rock"
	case Queen:
		return "Queen"
	case King:
		return "King"
	default:
		return "(badpiecetype)"
	}
}

// Color represents a color.
type Color uint

func (co Color) Other() Color {
	return White + Black - co
}

func (co Color) String() string {
	switch co {
	case NoColor:
		return "(nocolor)"
	case White:
		return "White"
	case Black:
		return "Black"
	default:
		return "(badcolor)"
	}
}

// Piece is a combination of piece type and color
type Piece uint

func ColorPiece(co Color, pt PieceType) Piece {
	return Piece(co<<3) + Piece(pt)
}

func (pi Piece) Color() Color {
	return Color(pi >> 3)
}

func (pi Piece) PieceType() PieceType {
	return PieceType(pi & 7)
}

// Symbol returns the piece as a string.
func (pi Piece) Symbol() string {
	co := pi.Color()
	pt := pi.PieceType()
	return PieceName[co][pt : pt+1]
}

func (pi Piece) String() string {
	return pi.Color().String() + " " + pi.PieceType().String()
}

// A birboard 8x8.
type Bitboard uint64

type MoveType int

type Move struct {
	From, To Square
	Capture  Piece
	MoveType MoveType
}

func (mo *Move) String() string {
	return mo.From.String() + mo.To.String()
}

type Position struct {
	byPieceType [PieceTypeMaxValue]Bitboard
	byColor     [ColorMaxValue]Bitboard
	toMove      Color
}

func PositionFromFEN(fen string) (*Position, error) {
	pos := &Position{}
	l := 0

	for r := 7; r >= 0; r-- {
		for f := 7; f >= 0; f-- {
			switch fen[l] {
			case 'p':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, Pawn))
			case 'n':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, Knight))
			case 'b':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, Bishop))
			case 'r':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, Rock))
			case 'q':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, Queen))
			case 'k':
				pos.PutPiece(RankFile(r, f), ColorPiece(Black, King))

			case 'P':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, Pawn))
			case 'N':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, Knight))
			case 'B':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, Bishop))
			case 'R':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, Rock))
			case 'Q':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, Queen))
			case 'K':
				pos.PutPiece(RankFile(r, f), ColorPiece(White, King))

			case '1', '2', '3', '4', '5', '6', '7', '8':
				f -= int(fen[l]) - int('0') - 1

			default:
				return nil, fmt.Errorf("unhandled '%c' at %d", fen[l], l)
			}
			l++
		}

		expected := uint8('/')
		if r == 0 {
			expected = uint8(' ')
		}
		if fen[l] != expected {
			return nil, fmt.Errorf("expected '%c', got '%c' at %d", expected, fen[l], l)
		}
		l++
	}

	switch fen[l] {
	case 'w':
		pos.toMove = White
	case 'l':
		pos.toMove = Black
	default:
		return nil, fmt.Errorf("unhandled %c at %d", fen[l], l)
	}
	l++

	// TODO: castling, en passant, halfmove, fullmove

	return pos, nil
}

// PutPiece puts a piece on the board.
// Does not do any checks.
func (pos *Position) PutPiece(sq Square, pi Piece) {
	pos.byColor[pi.Color()] |= sq.Bitboard()
	pos.byPieceType[pi.PieceType()] |= sq.Bitboard()
}

func (pos *Position) RemovePiece(sq Square, pi Piece) {
	pos.byColor[pi.Color()] &= ^sq.Bitboard()
	pos.byPieceType[pi.PieceType()] &= ^sq.Bitboard()
}

// GetPiece returns the piece at sq.
func (pos *Position) GetPiece(sq Square) Piece {
	var co Color
	var pt PieceType

	for co = ColorMinValue; co < ColorMaxValue; co++ {
		if pos.byColor[co]&sq.Bitboard() != 0 {
			break
		}
	}
	if co == ColorMaxValue {
		return ColorPiece(NoColor, NoPieceType)
	}

	for pt = PieceTypeMinValue; pt < PieceTypeMaxValue; pt++ {
		if pos.byPieceType[pt]&sq.Bitboard() != 0 {
			break
		}
	}
	if pt == PieceTypeMaxValue {
		panic("expected piece, got nothing")
	}

	return ColorPiece(co, pt)
}

// PrettyPrints pretty prints the current position.
func (pos *Position) PrettyPrint() {
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 7; f >= 0; f-- {
			line += pos.GetPiece(RankFile(r, f)).Symbol()
		}
		if r == 7 && pos.toMove == Black {
			line += " *"
		}
		if r == 0 && pos.toMove == White {
			line += " *"
		}
		log.Println(line)
	}

}

func (pos *Position) ParseMove(s string) Move {
	from := SquareFromString(s[0:2])
	to := SquareFromString(s[2:4])

	return Move{
		From:     from,
		To:       to,
		Capture:  pos.GetPiece(to),
		MoveType: Normal, // TODO
	}
}

// DoMove performs a move.
// Expects the move to be valid.
// TODO: castling, promotion
func (pos *Position) DoMove(mo Move) {
	// log.Println("Playing", mo)
	piece := pos.GetPiece(mo.From)
	pos.RemovePiece(mo.From, piece)
	pos.RemovePiece(mo.To, mo.Capture)
	pos.PutPiece(mo.To, piece)
	pos.toMove = pos.toMove.Other()
}

// UndoMove takes back a move.
// Expects the move to be valid.
// TODO: castling, promotion
func (pos *Position) UndoMove(mo Move) {
	// log.Println("Takeing back", mo)
	piece := pos.GetPiece(mo.To)
	pos.RemovePiece(mo.To, piece)
	pos.PutPiece(mo.From, piece)
	pos.PutPiece(mo.To, mo.Capture)
	pos.toMove = pos.toMove.Other()
}

var (
	knightJump = [8][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
)

func (pos *Position) genKnightMoves(from Square, pi Piece, moves []Move) []Move {
	if pi.PieceType() != Knight {
		panic(fmt.Sprintf("cannot move a %v, expected a %v", pi, Knight))
	}
	for _, e := range knightJump {
		r, f := from.Rank()+e[0], from.File()+e[1]
		if 0 > r || r >= 8 || 0 > f || f >= 8 {
			// Cannot jump out of the table.
			continue
		}
		to := RankFile(r, f)

		capture := pos.GetPiece(to)
		if capture.Color() == pi.Color() {
			// Cannot capture same color.
			continue
		}

		// Found a valid Knight move.
		moves = append(moves, Move{
			From:     from,
			To:       to,
			Capture:  capture,
			MoveType: Normal,
		})
	}
	return moves
}

func (pos *Position) GenerateMoves() []Move {
	moves := make([]Move, 0, 8)
	for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
		pi := pos.GetPiece(sq)
		if pi.Color() != pos.toMove {
			continue
		}

		switch pi.PieceType() {
		case Knight:
			log.Println("Found knight at", sq)
			pos.PrettyPrint()
			moves = pos.genKnightMoves(sq, pi, moves)
		}
	}
	return moves
}
