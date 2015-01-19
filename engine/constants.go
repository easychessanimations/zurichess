package engine

const (
	SquareA1 Square = iota
	SquareB1
	SquareC1
	SquareD1
	SquareE1
	SquareF1
	SquareG1
	SquareH1
	SquareA2
	SquareB2
	SquareC2
	SquareD2
	SquareE2
	SquareF2
	SquareG2
	SquareH2
	SquareA3
	SquareB3
	SquareC3
	SquareD3
	SquareE3
	SquareF3
	SquareG3
	SquareH3
	SquareA4
	SquareB4
	SquareC4
	SquareD4
	SquareE4
	SquareF4
	SquareG4
	SquareH4
	SquareA5
	SquareB5
	SquareC5
	SquareD5
	SquareE5
	SquareF5
	SquareG5
	SquareH5
	SquareA6
	SquareB6
	SquareC6
	SquareD6
	SquareE6
	SquareF6
	SquareG6
	SquareH6
	SquareA7
	SquareB7
	SquareC7
	SquareD7
	SquareE7
	SquareF7
	SquareG7
	SquareH7
	SquareA8
	SquareB8
	SquareC8
	SquareD8
	SquareE8
	SquareF8
	SquareG8
	SquareH8

	SquareArraySize = int(iota)
	SquareMinValue  = SquareA1
	SquareMaxValue  = SquareH8
)

// Piece constants must stay in sync with ColorFigure
const (
	NoPiece        = Piece(0)
	PieceArraySize = Piece(FigureArraySize << 2)
)

const (
	WhitePawn Piece = Piece(iota+Pawn)<<2 + Piece(White)
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
)

const (
	BlackPawn Piece = Piece(iota+Pawn)<<2 + Piece(Black)
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

const (
	BbEmpty           Bitboard = 0x0000000000000000
	BbFull            Bitboard = 0xffffffffffffffff
	BbBorder          Bitboard = 0xff818181818181ff
	BbPawnLeftAttack  Bitboard = 0x00fefefefefefe00
	BbPawnRightAttack Bitboard = 0x007f7f7f7f7f7f00
	BbPawnStartRank   Bitboard = 0x00ff00000000ff00
	BbPawnDoubleRank  Bitboard = 0x000000ffff000000
	BbBlackSquares    Bitboard = 0xaa55aa552a55aa55
	BbWhiteSquares    Bitboard = 0xd5aa55aad5aa55aa
)

var (
	// Some positions commonly used for testing.
	FENStartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	FENKiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	FENDuplain  = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
)
