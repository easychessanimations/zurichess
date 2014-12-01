package main

import (
	"fmt"
)

// Square identifies the location on the board.
type Square int

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

// Piece represents a colorless piece
type Piece int

// Color represents a color.
type Color int

// A birboard 8x8.
type Bitboard uint64

type Position struct {
	board  [ColorMaxValue][PieceMaxValue]Bitboard
	toMove Color
}

func PositionFromFEN(fen string) (*Position, error) {
	pos := &Position{}
	l := 0

	for r := 7; r >= 0; r-- {
		for f := 7; f >= 0; f-- {
			switch fen[l] {
			case 'p':
				pos.PutPiece(ColorBlack, PiecePawn, RankFile(r, f))
			case 'n':
				pos.PutPiece(ColorBlack, PieceKnight, RankFile(r, f))
			case 'b':
				pos.PutPiece(ColorBlack, PieceBishop, RankFile(r, f))
			case 'r':
				pos.PutPiece(ColorBlack, PieceRock, RankFile(r, f))
			case 'q':
				pos.PutPiece(ColorBlack, PieceQueen, RankFile(r, f))
			case 'k':
				pos.PutPiece(ColorBlack, PieceKing, RankFile(r, f))

			case 'P':
				pos.PutPiece(ColorWhite, PiecePawn, RankFile(r, f))
			case 'N':
				pos.PutPiece(ColorWhite, PieceKnight, RankFile(r, f))
			case 'B':
				pos.PutPiece(ColorWhite, PieceBishop, RankFile(r, f))
			case 'R':
				pos.PutPiece(ColorWhite, PieceRock, RankFile(r, f))
			case 'Q':
				pos.PutPiece(ColorWhite, PieceQueen, RankFile(r, f))
			case 'K':
				pos.PutPiece(ColorWhite, PieceKing, RankFile(r, f))

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
		pos.toMove = ColorWhite
	case 'l':
		pos.toMove = ColorBlack
	default:
		return nil, fmt.Errorf("unhandled %c at %d", fen[l], l)
	}
	l++

	// TODO: castling, en passant, halfmove, fullmove

	return pos, nil
}

// PutPiece puts a piece on the board.
// Does not do any checks.
func (pos *Position) PutPiece(co Color, pi Piece, sq Square) {
	pos.board[co][pi] |= sq.Bitboard()
}
