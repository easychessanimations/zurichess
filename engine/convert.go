package engine

import (
	"fmt"
	"strings"
)

type castleInfo struct {
	Castle Castle
	Piece  [2]Piece
	Square [2]Square
}

var (
	itoa               = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} // shortcut for Itoa
	colorToSymbol      = []string{"", "w", "b"}
	pieceToSymbol      = []string{".", "?", "P", "p", "N", "n", "B", "b", "R", "r", "Q", "q", "K", "k"}
	symbolToCastleInfo = map[rune]castleInfo{
		'K': castleInfo{
			Castle: WhiteOO,
			Piece:  [2]Piece{WhiteKing, WhiteRook},
			Square: [2]Square{SquareE1, SquareH1},
		},
		'k': castleInfo{
			Castle: BlackOO,
			Piece:  [2]Piece{BlackKing, BlackRook},
			Square: [2]Square{SquareE8, SquareH8},
		},
		'Q': castleInfo{
			Castle: WhiteOOO,
			Piece:  [2]Piece{WhiteKing, WhiteRook},
			Square: [2]Square{SquareE1, SquareA1},
		},
		'q': castleInfo{
			Castle: BlackOOO,
			Piece:  [2]Piece{BlackKing, BlackRook},
			Square: [2]Square{SquareE8, SquareA8},
		},
	}
	symbolToColor = map[string]Color{
		"w": White,
		"b": Black,
	}
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
)

// ParsePiecePlacement parse pieces from str (FEN like) into pos.
func ParsePiecePlacement(str string, pos *Position) error {
	ranks := strings.Split(str, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("expected 8 ranks, got %d", len(ranks))
	}
	for r := range ranks {
		f := 0
		for _, p := range ranks[r] {
			pi := symbolToPiece[p]
			if pi == NoPiece {
				if '1' <= p && p <= '8' {
					f += int(p) - int('0') - 1
				} else {
					return fmt.Errorf("expected rank or number, got %s", string(p))
				}
			}
			if f >= 8 {
				return fmt.Errorf("rank %d too long (%d cells)", 8-r, f)
			}
			// 7-r because FEN describes the table from 8th rank.
			pos.Put(RankFile(7-r, f), pi)
			f++
		}
		if f < 8 {
			return fmt.Errorf("rank %d too short (%d cells)", r+1, f)
		}
	}
	return nil
}

// FormatPiecePlacement converts a position to FEN piece placement.
func FormatPiecePlacement(pos *Position) string {
	s := ""
	for r := 7; r >= 0; r-- {
		space := 0
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			pi := pos.Get(sq)
			if pi == NoPiece {
				space++
			} else {
				if space != 0 {
					s += itoa[space]
					space = 0
				}
				s += pieceToSymbol[pi]
			}
		}

		if space != 0 {
			s += itoa[space]
		}
		if r != 0 {
			s += "/"
		}
	}
	return s
}

func ParseEnpassantSquare(str string, pos *Position) error {
	if str[:1] == "-" {
		pos.SetEnpassantSquare(SquareA1)
		return nil
	}
	sq, err := SquareFromString(str)
	if err != nil {
		return err
	}
	pos.SetEnpassantSquare(sq)
	return nil
}

// FormatEnpassantSquare converts position's castling ability to string.
func FormatEnpassantSquare(pos *Position) string {
	if pos.EnpassantSquare() != SquareA1 {
		return pos.EnpassantSquare().String()
	}
	return "-"
}

func ParseSideToMove(str string, pos *Position) error {
	if col, ok := symbolToColor[str]; ok {
		pos.SetSideToMove(col)
		return nil
	}
	return fmt.Errorf("invalid color %s", str)
}

func FormatSideToMove(pos *Position) string {
	return colorToSymbol[pos.SideToMove]
}

func ParseCastlingAbility(str string, pos *Position) error {
	if str == "-" {
		pos.SetCastlingAbility(NoCastle)
		return nil
	}

	ability := NoCastle
	for _, p := range str {
		info, ok := symbolToCastleInfo[p]
		if !ok {
			return fmt.Errorf("invalid castling ability %s", str)
		}
		ability |= info.Castle
		for i := 0; i < 2; i++ {
			if info.Piece[i] != pos.Get(info.Square[i]) {
				return fmt.Errorf("expected %v at %v, got %v",
					info.Piece[i], info.Square[i], pos.Get(info.Square[i]))
			}
		}
	}
	pos.SetCastlingAbility(ability)
	return nil
}

func FormatCastlingAbility(pos *Position) string {
	return pos.CastlingAbility().String()
}