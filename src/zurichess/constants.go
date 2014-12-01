package main

const (
	SquareA1 Square = iota
	SquareA2
	SquareA3
	SquareA4
	SquareA5
	SquareA6
	SquareA7
	SquareA8
	SquareB1
	SquareB2
	SquareB3
	SquareB4
	SquareB5
	SquareB6
	SquareB7
	SquareB8
	SquareC1
	SquareC2
	SquareC3
	SquareC4
	SquareC5
	SquareC6
	SquareC7
	SquareC8
	SquareD1
	SquareD2
	SquareD3
	SquareD4
	SquareD5
	SquareD6
	SquareD7
	SquareD8
	SquareE1
	SquareE2
	SquareE3
	SquareE4
	SquareE5
	SquareE6
	SquareE7
	SquareE8
	SquareF1
	SquareF2
	SquareF3
	SquareF4
	SquareF5
	SquareF6
	SquareF7
	SquareF8
	SquareG1
	SquareG2
	SquareG3
	SquareG4
	SquareG5
	SquareG6
	SquareG7
	SquareG8
	SquareH1
	SquareH2
	SquareH3
	SquareH4
	SquareH5
	SquareH6
	SquareH7
	SquareH8

	SquareMaxValue
	SquareMinValue = SquareA1
)

const (
	NoPieceType PieceType = iota
	Pawn
	Knight
	Bishop
	Rock
	Queen
	King

	PieceTypeMaxValue
	PieceTypeMinValue = Pawn
)

const (
	NoColor Color = iota
	White
	Black

	ColorMaxValue
	ColorMinValue = White
)

const (
	NoPiece = iota
)

const (
	Normal MoveType = iota
	Promotion
	Castling
	Enpassant
)

var (
	FENStartPos string   = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	PieceName   []string = []string{"       ", " PNBRQK", " pnbrqk "}
)
