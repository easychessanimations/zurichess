package main

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
	case Rook:
		return "Rook"
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

var (
	pieceSymbol = []string{"       ", " PNBRQK", " pnbrqk "}
)

// Symbol returns the piece as a string.
func (pi Piece) Symbol() string {
	co := pi.Color()
	pt := pi.PieceType()
	return pieceSymbol[co][pt : pt+1]
}

func (pi Piece) String() string {
	return pi.Color().String() + " " + pi.PieceType().String()
}

// An 8x8 bitboard.
type Bitboard uint64

func (bb Bitboard) LSB() Square {
	return Square(bb & (-bb))
}

// Move type.
type MoveType uint

type Move struct {
	MoveType  MoveType
	From, To  Square
	Capture   Piece
	Promotion Piece
}

func (mo *Move) String() string {
	return mo.From.String() + mo.To.String()
}

// Castle type
type Castle uint
