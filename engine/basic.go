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

	figureToSymbol = map[Figure]string{
		Knight: "N",
		Bishop: "B",
		Rook:   "R",
		Queen:  "Q",
		King:   "K",
	}
)

// Square identifies the location on the board.
type Square uint8

// RankFile returns a square with rank r and file f.
// r and f should be between 0 and 7.
func RankFile(r, f int) Square {
	return Square(r*8 + f)
}

// SquareFromString parses a square from a string.
// The string has standard chess format [a-h][1-8].
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

// Rank returns a number from 0 to 7 representing the rank of the square.
func (sq Square) Rank() int {
	return int(sq / 8)
}

// File returns a number from 0 to 7 representing the file of the square.
func (sq Square) File() int {
	return int(sq % 8)
}

func (sq Square) String() string {
	return string([]byte{
		uint8(sq.File() + 'a'),
		uint8(sq.Rank() + '1'),
	})
}

// Figure represents a piece without a color.
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

// Color represents a side.
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
	colorWeight  = [ColorArraySize]int{0, 1, -1}
	colorMask    = [ColorArraySize]Square{0, 0, 63} // colorMask[color] ^ square rotates the board.
	kingHomeRank = [ColorArraySize]int{0, 0, 7}
)

// Opposite returns the reversed color.
// Result is undefined if c is not White or Black.
func (c Color) Opposite() Color {
	return White + Black - c
}

// KingHomeRank return king's rank on starting position.
// Result is undefined if c is not White or Black.
func (c Color) KingHomeRank() int {
	return kingHomeRank[c]
}

// Piece is a figure owned by one side.
type Piece uint8

// ColorFigure returns a piece with col and fig.
func ColorFigure(col Color, fig Figure) Piece {
	return Piece(fig<<2) + Piece(col)
}

// Color returns piece's color.
func (pi Piece) Color() Color {
	return Color(pi & 3)
}

// Figure returns piece's figure.
func (pi Piece) Figure() Figure {
	return Figure(pi >> 2)
}

// An 8x8 bitboard.
type Bitboard uint64

// RankBb returns a bitboard with all bits on rank set.
func RankBb(rank int) Bitboard {
	rank1 := Bitboard(0x00000000000000ff)
	return rank1 << uint(8*rank)
}

// FileBb returns a bitboard with all bits on file set.
func FileBb(file int) Bitboard {
	fileA := Bitboard(0x0101010101010101)
	return fileA << uint(file)
}

// As square returns the occupied square if the bitboard has a single piece.
// If the board has more then one piece the result is undefined.
func (bb Bitboard) AsSquare() Square {
	return Square(logN(uint64(bb)))
}

// LSB picks a square in the board.
// Returns empty board for empty board.
func (bb Bitboard) LSB() Bitboard {
	return bb & (-bb)
}

// Popcnt counts number of squares set in bb.
func (bb Bitboard) Popcnt() int {
	return popcnt(uint64(bb))
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
	From, To       Square // source and destination squares
	Capture        Piece  // which piece is captured
	Target         Piece  // the piece on To, after the move
	MoveType       MoveType
	Data           int8   // some data, unrelated to move or position
	SavedEnpassant Square // old enpassant square
	SavedCastle    Castle // old castle rights
}

// SideToMove returns which player is moving.
func (m *Move) SideToMove() Color {
	return Color(m.Target & 3)
}

// CaptureSquare returns the captured piece square.
// If no piece is captured, the result is undefined.
func (m *Move) CaptureSquare() Square {
	if m.MoveType == Enpassant {
		return m.From&0x38 + m.To&0x7
	}
	return m.To
}

// Piece returns the piece moved.
func (m *Move) Piece() Piece {
	if m.MoveType != Promotion {
		return m.Target
	}
	return Piece(Pawn<<2) + m.Target&3
}

// Promotion returns the promoted piece if any.
func (m *Move) Promotion() Piece {
	if m.MoveType != Promotion {
		return NoPiece
	}
	return m.Target
}

// IsViolent returns true if the move can change the position's score
// significantly.
func (m *Move) IsViolent() bool {
	return m.Capture != NoPiece || m.MoveType == Promotion
}

// UCI converts a move to UCI format.
// The protocol specification at http://wbec-ridderkerk.nl/html/UCIProtocol.html
// incorrectly states that this is the long algebraic notation (LAN).
func (m *Move) UCI() string {
	return m.From.String() + m.To.String() + figureToSymbol[m.Promotion().Figure()]
}

// LAN converts a move to Long Algebraic Notation.
// http://en.wikipedia.org/wiki/Algebraic_notation_%28chess%29#Long_algebraic_notation
func (m *Move) LAN() string {
	r := figureToSymbol[m.Piece().Figure()] + m.From.String()
	if m.Capture != NoPiece {
		r += "-"
	} else {
		r += "x"
	}
	r += m.To.String() + figureToSymbol[m.Promotion().Figure()]
	return r
}

func (m *Move) String() string {
	return m.UCI()
}

// Castling rights mask.
type Castle uint8

const (
	// White can castle on King side.
	WhiteOO Castle = 1 << iota
	// White can castle on Queen side.
	WhiteOOO
	// Black can castle on King side.
	BlackOO
	// Black can castle on Queen side.
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

func (c Castle) String() string {
	if c == 0 {
		return "-"
	}

	var r []byte
	for c > 0 {
		k := c & (-c)
		r = append(r, castleToSymbol[k])
		c -= k
	}
	return string(r)
}

// CastlingRook returns the rook moved during castling
// together with starting and stopping squares.
func CastlingRook(kingEnd Square) (Piece, Square, Square) {
	// Explanation how rookStart works for king on E1.
	// if kingEnd == C1 == b010, then rookStart == A1 == b000
	// if kingEnd == G1 == b110, then rookStart == H1 == b111
	// So bit 3 will set bit 2 and bit 1.
	//
	// Explanation how rookEnd works for king on E1.
	// if kingEnd == C1 == b010, then rookEnd == D1 == b011
	// if kingEnd == G1 == b110, then rookEnd == F1 == b101
	// So bit 3 will invert bit 2. bit 1 is always set.
	piece := Piece(Rook<<2) + 1 + Piece(kingEnd>>5)
	rookStart := kingEnd&^3 | (kingEnd & 4 >> 1) | (kingEnd & 4 >> 2)
	rookEnd := kingEnd ^ (kingEnd & 4 >> 1) | 1
	return piece, rookStart, rookEnd
}
