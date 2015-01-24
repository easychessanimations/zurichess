package engine

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	MidGame = 0
	EndGame = 1
)

var (
	// Scores returned directly are int16.
	KnownWinScore int16 = 20000
	MateScore     int16 = 30000
	InfinityScore int16 = 32000

	// Bonuses and penalties have type int in order to prevent accidental
	// overflows during computation of the position's score.
	BishopPairBonus int = 40
	KnightPawnBonus int = 6
	RookPawnPenalty int = 12

	// Figure middle and end game bonuses.
	// TODO: Given tapered eval KnightPawnBonus and RookPawnPenalty
	// should be part of FigureBonus.
	FigureBonus = [2][FigureArraySize]int{
		{0, 100, 315, 325, 475, 975, 10000},  // MidGame
		{0, 100, 345, 355, 525, 1000, 10000}, // EndGame
	}

	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// Theses values were suggested by Tomasz Michniewski as an extremely basic
	// evaluation. The original values were copied from:
	// https://chessprogramming.wikispaces.com/Simplified+evaluation+function
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable = [FigureArraySize][64][2]int{
		{ // NoFigure
		},
		{ // Pawn
			{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
			{5, 5}, {10, 10}, {10, 10}, {-20, -20}, {-20, -20}, {10, 10}, {10, 10}, {5, 5},
			{5, 5}, {-5, -5}, {-10, -10}, {0, 0}, {0, 0}, {-10, -10}, {-5, -5}, {5, 5},
			{0, 0}, {0, 0}, {0, 0}, {20, 20}, {20, 20}, {0, 0}, {0, 0}, {0, 0},
			{5, 5}, {5, 5}, {10, 10}, {25, 25}, {25, 25}, {10, 10}, {5, 5}, {5, 5},
			{10, 10}, {10, 10}, {20, 20}, {30, 30}, {30, 30}, {20, 20}, {10, 10}, {10, 10},
			{50, 50}, {50, 50}, {50, 50}, {50, 50}, {50, 50}, {50, 50}, {50, 50}, {50, 50},
			{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
		},
		{ // Knight
			{-50, -50}, {-40, -40}, {-30, -30}, {-30, -30}, {-30, -30}, {-30, -30}, {-40, -40}, {-50, -50},
			{-40, -40}, {-20, -20}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-20, -20}, {-40, -40},
			{-30, -30}, {0, 0}, {10, 10}, {15, 15}, {15, 15}, {10, 10}, {0, 0}, {-30, -30},
			{-30, -30}, {5, 5}, {15, 15}, {20, 20}, {20, 20}, {15, 15}, {5, 5}, {-30, -30},
			{-30, -30}, {0, 0}, {15, 15}, {20, 20}, {20, 20}, {15, 15}, {0, 0}, {-30, -30},
			{-30, -30}, {5, 5}, {10, 10}, {15, 15}, {15, 15}, {10, 10}, {5, 5}, {-30, -30},
			{-40, -40}, {-20, -20}, {0, 0}, {5, 5}, {5, 5}, {0, 0}, {-20, -20}, {-40, -40},
			{-50, -50}, {-40, -40}, {-30, -30}, {-30, -30}, {-30, -30}, {-30, -30}, {-40, -40}, {-50, -50},
		},
		{ // Bishop
			{-20, -20}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-20, -20},
			{-10, -10}, {5, 5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {5, 5}, {-10, -10},
			{-10, -10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {-10, -10},
			{-10, -10}, {0, 0}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {0, 0}, {-10, -10},
			{-10, -10}, {5, 5}, {5, 5}, {10, 10}, {10, 10}, {5, 5}, {5, 5}, {-10, -10},
			{-10, -10}, {0, 0}, {5, 5}, {10, 10}, {10, 10}, {5, 5}, {0, 0}, {-10, -10},
			{-10, -10}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-10, -10},
			{-20, -20}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-10, -10}, {-20, -20},
		},
		{ // Rook
			{0, 0}, {0, 0}, {0, 0}, {5, 5}, {5, 5}, {0, 0}, {0, 0}, {0, 0},
			{-5, -5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-5, -5},
			{-5, -5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-5, -5},
			{-5, -5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-5, -5},
			{-5, -5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-5, -5},
			{-5, -5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-5, -5},
			{5, 5}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {5, 5},
			{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
		},
		{ // Queen
			{-20, -20}, {-10, -10}, {-10, -10}, {-5, -5}, {-5, -5}, {-10, -10}, {-10, -10}, {-20, -20},
			{-10, -10}, {0, 0}, {5, 5}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-10, -10},
			{-10, -10}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {0, 0}, {-10, -10},
			{0, 0}, {0, 0}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {0, 0}, {-5, -5},
			{-5, -5}, {0, 0}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {0, 0}, {-5, -5},
			{-10, -10}, {0, 0}, {5, 5}, {5, 5}, {5, 5}, {5, 5}, {0, 0}, {-10, -10},
			{-10, -10}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {-10, -10},
			{-20, -20}, {-10, -10}, {-10, -10}, {-5, -5}, {-5, -5}, {-10, -10}, {-10, -10}, {-20, -20},
		},
		{ // King
			{20, -50}, {30, -30}, {10, -30}, {0, -30}, {0, -30}, {10, -30}, {30, -30}, {20, -50},
			{20, -30}, {20, -30}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {20, -30}, {20, -30},
			{-10, -30}, {-20, -10}, {-20, 20}, {-20, 30}, {-20, 30}, {-20, 20}, {-20, -10}, {-10, -30},
			{-20, -30}, {-30, -10}, {-30, 30}, {-40, 40}, {-40, 40}, {-30, 30}, {-30, 10}, {-20, -30},
			{-30, -30}, {-40, -10}, {-40, 30}, {-50, 40}, {-50, 40}, {-40, 30}, {-40, -10}, {-30, -30},
			{-30, -30}, {-40, -10}, {-40, 20}, {-50, 30}, {-50, 30}, {-40, 20}, {-40, -10}, {-30, -30},
			{-30, -30}, {-40, -20}, {-40, -10}, {-50, 0}, {-50, 0}, {-40, -10}, {-40, -20}, {-30, -30},
			{-30, -50}, {-40, -40}, {-40, -30}, {-50, -20}, {-50, -20}, {-40, -30}, {-40, -40}, {-30, -50},
		},
	}

	// See MvvLva()
	// mvvLva[attacker * FigureSize + victim]
	mvvLva = [FigureArraySize * FigureArraySize]int{
		242, 64, 85, 341, 341, 668, 20000,
		192, 495, 711, 809, 817, 981, 20000,
		363, 409, 650, 683, 818, 992, 20000,
		423, 441, 468, 663, 784, 900, 20000,
		59, 408, 482, 638, 775, 945, 20000,
		25, 480, 505, 526, 635, 889, 20000,
		250, 458, 701, 714, 904, 929, 20000,
	}
)

// SetMaterialValue parses str and updates array.
//
// str has format "value0,value1,...,valuen-1" (no spaces and no quotes).
// If valuei is empty then array[i] is not modified.
// n, the number of values, must be equal to len(array).
func SetMaterialValue(name string, array []int, str string) error {
	fields := strings.Split(str, ",")
	if len(fields) != len(array) {
		return fmt.Errorf("%s: expected %d elements, got %d",
			name, len(array), len(fields))
	}
	for _, f := range fields {
		if f != "" {
			if _, err := strconv.ParseInt(f, 10, 0); err != nil {
				return fmt.Errorf("%s: %v", name, err)
			}
		}
	}
	for i, f := range fields {
		if f != "" {
			value, _ := strconv.ParseInt(f, 10, 32)
			array[i] = int(value)
		}
	}
	return nil
}

func SetMvvLva(str string) error {
	return SetMaterialValue("MvvLva", mvvLva[:], str)
}

// MvvLva returns a ordering score.
//
// MvvLva stands for "Most valuable victim, Least valuable attacker".
// See https://chessprogramming.wikispaces.com/MVV-LVA.
// In zurichess the MVV/LVA formula is not used,
// but the values are optimized and stored in the mvvLva array.
func MvvLva(att, capt Figure) int {
	return mvvLva[int(att)*FigureArraySize+int(capt)]
}
