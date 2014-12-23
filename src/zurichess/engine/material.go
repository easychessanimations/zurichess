// psqt.go stores the Piece Square Tables.
// Every piece gets a bonus depending on the position on the table.
package engine

const (
	MidGame = 0
	EndGame = 1
)

var (
	KnownWinScore = 20000
	MateScore     = 30000
	InfinityScore = 32000

	BishopPairBonus = 40
	KnightPawnBonus = 6
	RookPawnPenalty = 12

	// Figure middle and end game bonuses.
	FigureBonus = [FigureMaxValue][2]int{
		{0, 0},         // NoFigure
		{100, 100},     // Pawn
		{345, 345},     // Knight
		{355, 355},     // Bishop
		{525, 525},     // Rook
		{1000, 1000},   // Queen
		{10000, 10000}, // King
	}

	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// Theses values were suggested by Tomasz Michniewski as an extremely basic
	// evaluation. The original values were copied from:
	// https://chessprogramming.wikispaces.com/Simplified+evaluation+function
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable = [FigureMaxValue][64][2]int{
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
)