// psqt.go stores the Piece Square Tables.
// Every piece gets a bonus depending on the position on the table.
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

	// Bonuses and penalties enter score calculation and are ints
	// to prevent accidental overflows during computation of the final
	// score.
	BishopPairBonus int = 40
	KnightPawnBonus int = 6
	RookPawnPenalty int = 12

	// Figure middle and end game bonuses.
	FigureBonus = [FigureArraySize][2]int{
		{0, 0},         // NoFigure
		{100, 100},     // Pawn
		{345, 345},     // Knight
		{355, 355},     // Bishop
		{475, 525},     // Rook
		{975, 1000},    // Queen
		{10000, 10000}, // King
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
		0, 369, 902, 1432, 2102, 2534, 20000, // Promotion
		0, 1017, 1151, 1735, 2093, 2146, 20000, // Pawn
		0, 454, 1213, 1602, 2410, 2973, 20000, // Knight
		0, 447, 641, 1340, 1906, 2740, 20000, // Bishop
		0, 24, 599, 1174, 1737, 2565, 20000, // Rook
		0, 81, 521, 1074, 1604, 1972, 20000, // Queen
		0, 981, 1815, 1839, 2673, 3391, 20000, // King
	}
)

// SetMaterialValue updates array from a string.
// str has format "value,value,...,value" (no spaces and no quotes).
// value can be empty to let the value intact.
// The number of values has to match the array size.
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
// MvvLva stands for "Most valuable victim, Least valuable attacker".
// See https://chessprogramming.wikispaces.com/MVV-LVA
// In zurichess the MVV/LVA formula is not used,
// but the values are optimized and stored in mvvLva array.
func MvvLva(att, capt Figure) int {
	return mvvLva[int(att)*FigureArraySize+int(capt)]
}
