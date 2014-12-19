package engine

import (
	"fmt"
	"strings"
)

var (
	symbolToPiece = map[rune]Piece{
		'p': BlackPawn,
		'n': BlackKnight,
		'b': BlackBishop,
		'r': BlackRook,
		'q': BlackQueen,
		'k': BlackKing,

		'P': WhitePawn,
		'N': WhiteKnight,
		'B': WhiteBishop,
		'R': WhiteRook,
		'Q': WhiteQueen,
		'K': WhiteKing,
	}

	symbolToCastle = map[rune]Castle{
		'K': WhiteOO,
		'Q': WhiteOOO,
		'k': BlackOO,
		'q': BlackOOO,
	}
)

// Parses position from a FEN string.
// See http://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
func PositionFromFEN(fen string) (*Position, error) {
	field := strings.Fields(fen)
	if len(field) < 4 {
		return nil, fmt.Errorf("expected at least 4 fields, got %d", len(field))
	}

	pos := &Position{}

	// Parse position.
	ranks := strings.Split(field[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("expected 8 rows, got %d", len(ranks))
	}
	for r := range ranks {
		sq := RankFile(7-r, 0) // FEN describes the table from 8th rank.
		for _, p := range ranks[r] {
			pi := symbolToPiece[p]
			if pi == NoPiece {
				if '1' <= p && p <= '8' {
					sq = sq.Relative(0, int(p)-int('0')-1)
				} else {
					return nil, fmt.Errorf("unhandled '%c'", p)
				}
			}
			pos.Put(sq, pi)
			sq = sq.Relative(0, 1)
		}
	}

	// Parse next to move.
	switch field[1] {
	case "w":
		pos.SetToMove(White)
	case "b":
		pos.SetToMove(Black)
	default:
		return nil, fmt.Errorf("unknown color %s", field[1])
	}

	// Parse castling rights.
	castle := NoCastle
	for _, p := range field[2] {
		vastle |= symbolToCastle[p]
	}
	pos.SetCastle(castle)

	// Parse Enpassant.
	// TODO: handle error
        if field[3][:1] != "-" {
		pos.SetEnpassant(SquareFromString(field[3]))
	}

	// TODO: halfmove, fullmove
	return pos, nil
}
