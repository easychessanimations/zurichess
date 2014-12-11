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

// Figure represents a colorless piece
type Figure uint

func (pt Figure) String() string {
	switch pt {
	case NoFigure:
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

func ColorFigure(co Color, pt Figure) Piece {
	return Piece(pt<<2) + Piece(co)
}

func (pi Piece) Color() Color {
	return Color(pi & 3)
}

func (pi Piece) Figure() Figure {
	return Figure(pi >> 2)
}

var pieceSymbol = []string{"       ", " PNBRQK", " pnbrqk "}

// Symbol returns the piece as a string.
func (pi Piece) Symbol() string {
	co := pi.Color()
	pt := pi.Figure()
	return pieceSymbol[co][pt : pt+1]
}

func (pi Piece) String() string {
	if pi == NoPiece {
		return "(nopiece)"
	}
	return pi.Color().String() + " " + pi.Figure().String()
}

// An 8x8 bitboard.
type Bitboard uint64

// LSB picks a square in the board.
func (bb Bitboard) LSB() Bitboard {
	return Bitboard(LSB(uint64(bb)))
}

// If the bitboard has a single piece, returns the occupied square.
func (bb Bitboard) AsSquare() Square {
	return Square(LogN(uint64(bb)))
}

// Pop pops a set square from the bitboard.
func (bb *Bitboard) Pop() Square {
	sq := (*bb).LSB()
	*bb -= sq
	return sq.AsSquare()
}

// Move type.
type MoveType uint

type Move struct {
	MoveType     MoveType
	From, To     Square
	Capture      Piece
	Promotion    Piece
	OldCastle    Castle
	OldEnpassant Square
}

func (mo Move) String() string {
	r := mo.From.String() + mo.To.String()
	if mo.MoveType == Promotion {
		s := mo.Promotion.Figure()
		r += string(pieceSymbol[White][s : s+1])
	}
	return r
}

// Castle type
type Castle uint

var castleSymbol = map[Castle]byte{
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
	for k, v := range castleSymbol {
		if ca&k != 0 {
			r = append(r, v)
		}
	}
	return string(r)
}

func RankBb(rank int) Bitboard {
	return BbRank1 << uint(8*rank)
}

func FileBb(file int) Bitboard {
	return BbFileA << uint(file)
}
