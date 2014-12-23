package engine

// Square identifies the location on the board.
type Square uint8

func RankFile(r, f int) Square {
	return Square(r*8 + f)
}

func SquareFromString(s string) Square {
	r := int(s[1] - '1')
	f := int(s[0] - 'a')
	return RankFile(r, f)
}

// RookStartSquare returns the rook starts square on castling.
func RookStartSquare(kingEnd Square) Square {
	// How it works for king on E1.
	// if kingEnd == C1 == b010, then rookStart == A1 == b000
	// if kingEnd == G1 == b110, then rookStart == H1 == b111
	// So bit 3 will set bit 2 and bit 1.
	return kingEnd&^3 | (kingEnd & 4 >> 1) | (kingEnd & 4 >> 2)
}

// RookEndSquare returns the rook starts square on castling.
func RookEndSquare(kingEnd Square) Square {
	// How it works for king on E1.
	// if kingEnd == C1 == b010, then rookEnd == D1 == b011
	// if kingEnd == G1 == b110, then rookEnd == F1 == b101
	// So bit 3 will invert bit 2. bit 1 is always set.
	return kingEnd ^ (kingEnd & 4 >> 1) | 1
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

	FigureMaxValue
	FigureMinValue = Pawn
)

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

const (
	NoColor Color = iota
	White
	Black

	ColorMaxValue
	ColorMinValue = White
)

var (
	ColorWeight = [ColorMaxValue]int{0, 1, -1}
	ColorMask   = [ColorMaxValue]Square{0, 0, 63} // ColorMask[color] ^ square rotates the board.
)

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
type Piece uint8

func ColorFigure(co Color, pt Figure) Piece {
	return Piece(pt<<2) + Piece(co)
}

// CastlingRook returns which rook is moved on castling.
func CastlingRook(kingEnd Square) Piece {
	return Piece(Rook<<2) + 1 + Piece(kingEnd>>5)
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
func (bb Bitboard) lsb() Bitboard {
	return bb & (-bb)
	/*
		        // golang is bad at inlining .LSB if it calls LSB
			return Bitboard(LSB(uint64(bb)))
	*/
}

// Pop pops a set square from the bitboard.
func (bb *Bitboard) Pop() Square {
	sq := (*bb).lsb()
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

type Move struct {
	From, To     Square // Source and destination
	Capture      Piece  // Which piece is captured
	Target       Piece  // Target is the piece on To, after the move.
	MoveType     MoveType
	OldEnpassant Square // Old enpassant square
	OldCastle    Castle // Old castle rights
}

func (mo Move) String() string {
	r := mo.From.String() + mo.To.String()
	if mo.MoveType == Promotion {
		s := mo.Target.Figure()
		r += string(pieceSymbol[White][s : s+1])
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

	NoCastle       = Castle(0)
	AnyCastle      = WhiteOO | WhiteOOO | BlackOO | BlackOOO
	CastleMinValue = Castle(0)
	CastleMaxValue = AnyCastle + 1
)

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
