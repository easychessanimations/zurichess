//go:generate stringer -type Figure
//go:generate stringer -type Color
//go:generate stringer -type Piece
//go:generate stringer -type MoveType
package engine

import (
	"fmt"
)

var (
	errorInvalidSquare = fmt.Errorf("invalid square")
)

// Square identifies the location on the board.
type Square uint8

func RankFile(r, f int) Square {
	return Square(r*8 + f)
}

func SquareFromString(s string) (Square, error) {
	if len(s) != 2 {
		return SquareA1, errorInvalidSquare
	}

	f, r := -1, -1
	if 'a' <= s[0] && s[0] <= 'h' {
		f = int(s[0] - 'a')
	}
	if 'A' <= s[0] && s[0] <= 'H' {
		f = int(s[0] - 'A')
	}
	if '1' <= s[1] && s[1] <= '8' {
		r = int(s[1] - '1')
	}
	if f == -1 || r == -1 {
		return SquareA1, errorInvalidSquare
	}

	return RankFile(r, f), nil
}

// Bitboard returns a bitboard that has sq set.
func (sq Square) Bitboard() Bitboard {
	return 1 << uint(sq)
}

func (sq Square) Relative(dr, df int) Square {
	return sq + Square(dr*8+df)
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

// Figure represents a colorless piece
type Figure uint

const (
	NoFigure Figure = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King

	FigureArraySize = int(iota)
	FigureMinValue  = Pawn
	FigureMaxValue  = King
)

// Color represents a color.
type Color uint

const (
	NoColor Color = iota
	White
	Black

	ColorArraySize = int(iota)
	ColorMinValue  = White
	ColorMaxValue  = Black
)

var (
	ColorWeight = [ColorArraySize]int{0, 1, -1}
	ColorMask   = [ColorArraySize]Square{0, 0, 63} // ColorMask[color] ^ square rotates the board.
)

func (co Color) Other() Color {
	return White + Black - co
}

// Piece is a combination of piece type and color
type Piece uint8

func ColorFigure(co Color, pt Figure) Piece {
	return Piece(pt<<2) + Piece(co)
}

func (pi Piece) Color() Color {
	return Color(pi & 3)
}

func (pi Piece) Figure() Figure {
	return Figure(pi >> 2)
}

// An 8x8 bitboard.
type Bitboard uint64

func RankBb(rank int) Bitboard {
	return BbRank1 << uint(8*rank)
}

func FileBb(file int) Bitboard {
	return BbFileA << uint(file)
}

// If the bitboard has a single piece, returns the occupied square.
func (bb Bitboard) AsSquare() Square {
	return Square(debrujin64[bb*debrujinMul>>debrujinShift])
	/*
		        // golang is bad at inlining .AsSquare if it calls LogN
			return Square(LogN(uint64(bb)))
	*/
}

// LSB picks a square in the board.
// Returns empty board for empty board.
func (bb Bitboard) LSB() Bitboard {
	return bb & (-bb)
}

// Pop pops a set square from the bitboard.
func (bb *Bitboard) Pop() Square {
	sq := (*bb).LSB()
	*bb -= sq
	return sq.AsSquare()
}

// Move type.
type MoveType uint8

const (
	NoMove MoveType = iota
	Normal
	Promotion
	Castling
	Enpassant
)

// Move stores a position dependent move.
type Move struct {
	From, To       Square // Source and destination
	Capture        Piece  // Which piece is captured
	Target         Piece  // Target is the piece on To, after the move.
	MoveType       MoveType
	SavedEnpassant Square // Old enpassant square
	SavedCastle    Castle // Old castle rights
}

// Piece returns the piece moved.
func (m *Move) Piece() Piece {
	if m.MoveType != Promotion {
		return m.Target
	}
	return Piece(Pawn<<2) + m.Target&3
}

// Promotion return the promovated piece if any.
func (m *Move) Promotion() Piece {
	if m.MoveType != Promotion {
		return NoPiece
	}
	return m.Target
}

func (m Move) String() string {
	r := m.From.String() + m.To.String()
	if m.MoveType == Promotion {
		r += string(pieceToSymbol[m.Target])
	}
	return r
}

// Castle type
type Castle uint16

const (
	WhiteOO Castle = 1 << iota
	WhiteOOO
	BlackOO
	BlackOOO

	NoCastle  Castle = 0
	AnyCastle Castle = WhiteOO | WhiteOOO | BlackOO | BlackOOO

	CastleArraySize = int(AnyCastle + 1)
	CastleMinValue  = NoCastle
	CastleMaxValue  = AnyCastle
)

var castleToSymbol = map[Castle]byte{
	WhiteOO:  'K',
	WhiteOOO: 'Q',
	BlackOO:  'k',
	BlackOOO: 'q',
}

func (ca Castle) String() string {
	if ca == 0 {
		return "-"
	}

	var r []byte
	for k, v := range castleToSymbol {
		if ca&k != 0 {
			r = append(r, v)
		}
	}
	return string(r)
}

// CastlingRook returns which rook is moved on castling.
//
// How rookStart it works for king on E1.
// if kingEnd == C1 == b010, then rookStart == A1 == b000
// if kingEnd == G1 == b110, then rookStart == H1 == b111
// So bit 3 will set bit 2 and bit 1.
//
// How rookEnd works for king on E1.
// if kingEnd == C1 == b010, then rookEnd == D1 == b011
// if kingEnd == G1 == b110, then rookEnd == F1 == b101
// So bit 3 will invert bit 2. bit 1 is always set.
func CastlingRook(kingEnd Square) (Piece, Square, Square) {
	piece := Piece(Rook<<2) + 1 + Piece(kingEnd>>5)
	rookStart := kingEnd&^3 | (kingEnd & 4 >> 1) | (kingEnd & 4 >> 2)
	rookEnd := kingEnd ^ (kingEnd & 4 >> 1) | 1
	return piece, rookStart, rookEnd
}
