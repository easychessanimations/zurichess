package main

import (
	"fmt"
	"strings"
)

// Parses position from a FEN string.
// See http://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
func PositionFromFEN(fen string) (*Position, error) {
	fld := strings.Fields(fen)
	if len(fld) < 4 {
		return nil, fmt.Errorf("expected at least 4 fields, got %d", len(fld))
	}

	pos := &Position{}

	// Parse position.
	ranks := strings.Split(fld[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("expected 8 rows, got %d", len(ranks))
	}
	for r := range ranks {
		sq := RankFile(7-r, 0) // FEN describes the table from 8th rank.
		for _, p := range ranks[r] {
			pi := NoPiece
			switch p {
			case 'p':
				pi = ColorPiece(Black, Pawn)
			case 'n':
				pi = ColorPiece(Black, Knight)
			case 'b':
				pi = ColorPiece(Black, Bishop)
			case 'r':
				pi = ColorPiece(Black, Rook)
			case 'q':
				pi = ColorPiece(Black, Queen)
			case 'k':
				pi = ColorPiece(Black, King)

			case 'P':
				pi = ColorPiece(White, Pawn)
			case 'N':
				pi = ColorPiece(White, Knight)
			case 'B':
				pi = ColorPiece(White, Bishop)
			case 'R':
				pi = ColorPiece(White, Rook)
			case 'Q':
				pi = ColorPiece(White, Queen)
			case 'K':
				pi = ColorPiece(White, King)

			case '1', '2', '3', '4', '5', '6', '7', '8':
				sq = sq.Relative(0, int(p)-int('0')-1)

			default:
				return nil, fmt.Errorf("unhandled '%c'", p)
			}
			pos.PutPiece(sq, pi)
			sq = sq.Relative(0, 1)
		}
	}

	// Parse next to move.
	switch fld[1] {
	case "w":
		pos.toMove = White
	case "b":
		pos.toMove = Black
	default:
		return nil, fmt.Errorf("unknown color %s", fld[1])
	}

	// Parse castling rights.
	for _, p := range fld[2] {
		switch p {
		case 'K':
			pos.castle |= WhiteOO
		case 'Q':
			pos.castle |= WhiteOOO
		case 'k':
			pos.castle |= BlackOO
		case 'q':
			pos.castle |= BlackOOO
		}
	}

	// Parse enpassant.
	// TODO: handle error
	if fld[3] != "-" {
		pos.enpassant = SquareFromString(fld[3])
	}

	// TODO: halfmove, fullmove
	return pos, nil
}
