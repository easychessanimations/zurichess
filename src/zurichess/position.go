package main

import (
	"fmt"
	"log"
)

// Square identifies the location on the board.
type Square uint

func RankFile(r, f int) Square {
	return Square(r*8 + f)
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

// PieceType represents a colorless piece
type PieceType uint

// Color represents a color.
type Color uint

func (co Color) String() string {
	switch co {
	case White:
		return "White"
	case Black:
		return "Block"
	case NoColor:
		return "(nocolor)"
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

// String returns the piece as a string.
func (pi Piece) String() string {
	co := pi.Color()
	pt := pi.PieceType()
	return PieceName[co][pt : pt+1]
}

// A birboard 8x8.
type Bitboard uint64

type MoveType int

type Move struct {
	From, To      Square
	MoveType      MoveType
	CapturedColor Color
	CapturedPiece Piece
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

// GetPiece returns the piece at sq.
func (pos *Position) GetPiece(sq Square) Piece {
	var co Color
	var pt PieceType

	for co = ColorMinValue; co < ColorMaxValue; co++ {
		if pos.byColor[co]&sq.Bitboard() != 0 {
			break
		}
	}

	for pt = PieceTypeMinValue; pt < PieceTypeMaxValue; pt++ {
		if pos.byPieceType[pt]&sq.Bitboard() != 0 {
			break
		}
	}

	if co == ColorMaxValue || pt == PieceTypeMaxValue {
		return ColorPiece(NoColor, NoPieceType)
	}
	return ColorPiece(co, pt)
}

// PrettyPrints pretty prints the current position.
func (pos *Position) PrettyPrint() {
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 7; f >= 0; f-- {
			line += pos.GetPiece(RankFile(r, f)).String()
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
