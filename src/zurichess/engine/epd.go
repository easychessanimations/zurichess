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

// Extended Position Description
type EPD struct {
	Position *Position
}

func (epd *EPD) parsePosition(str string) error {
	// Parse position.
	ranks := strings.Split(str, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("expected 8 rows, got %d", len(ranks))
	}
	for r := range ranks {
		sq := RankFile(7-r, 0) // FEN describes the table from 8th rank.
		for _, p := range ranks[r] {
			pi := symbolToPiece[p]
			if pi == NoPiece {
				if '1' <= p && p <= '8' {
					sq = sq.Relative(0, int(p)-int('0')-1)
				} else {
					return fmt.Errorf("unhandled '%c'", p)
				}
			}
			epd.Position.Put(sq, pi)
			sq = sq.Relative(0, 1)
		}
	}
	return nil
}

func (epd *EPD) parseToMove(str string) error {
	switch str {
	case "w":
		pos.SetToMove(White)
	case "b":
		pos.SetToMove(Black)
	default:
		return fmt.Errorf("unknown color %s", str)
	}
	return nil
}

func (epd *EPD) parseCastlingRights(str string) error {
	if str == "-" {
		return nil
	}
	for _, p := range str {
		castle := symbolToCastle[p]
		if castle == NoCastle {
			return fmt.Errorf("unknown castle rights %s", str)
		}
		epd.Position.Castle |= castle
	}
	return nil
}

func (epd *EPD) parseEnpassant(str string) error {
	// TODO: handle error
        if str[:1] != "-" {
		pos.SetEnpassant(SquareFromString(field[3]))
        }
	return nil
}

// Parses position from a FEN string.
// See http://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
func PositionFromFEN(fen string) (*Position, error) {
	field := strings.Fields(fen)
	if len(field) < 6 {
		return nil, fmt.Errorf("expected at least 6 fields, got %d", len(field))
	}

	epd := &EPD{Position: &Position{}}
	if err := epd.parsePosition(field[0]); err != nil {
		return nil, err
	}
	if err := epd.parseToMove(field[1]); err != nil {
		return nil, err
	}
	if err := epd.parseCastlingRights(field[2]); err != nil {
		return nil, err
	}
	if err := epd.parseEnpassant(field[3]); err != nil {
		return nil, err
	}

	return epd.Position, nil
}
