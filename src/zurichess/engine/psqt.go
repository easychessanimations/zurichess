// psqt.go stores the Piece Square Tables.
// Every piece gets a bonus depending on the position on the table.
package engine

var (
	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// Theses values were suggested by Tomasz Michniewski as an extremely basic
	// evaluation. The original values were copied from:
	// https://chessprogramming.wikispaces.com/Simplified+evaluation+function
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable = [FigureMaxValue][64]int{
		{ // NoFigure
		},
		{ // Pawn
			0, 0, 0, 0, 0, 0, 0, 0,
			5, 10, 10, -20, -20, 10, 10, 5,
			5, -5, -10, 0, 0, -10, -5, 5,
			0, 0, 0, 20, 20, 0, 0, 0,
			5, 5, 10, 25, 25, 10, 5, 5,
			10, 10, 20, 30, 30, 20, 10, 10,
			50, 50, 50, 50, 50, 50, 50, 50,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{ // Knight
			-50, -40, -30, -30, -30, -30, -40, -50,
			-40, -20, 0, 0, 0, 0, -20, -40,
			-30, 0, 10, 15, 15, 10, 0, -30,
			-30, 5, 15, 20, 20, 15, 5, -30,
			-30, 0, 15, 20, 20, 15, 0, -30,
			-30, 5, 10, 15, 15, 10, 5, -30,
			-40, -20, 0, 5, 5, 0, -20, -40,
			-50, -40, -30, -30, -30, -30, -40, -50,
		},
		{ // Bishop
			-20, -10, -10, -10, -10, -10, -10, -20,
			-10, 5, 0, 0, 0, 0, 5, -10,
			-10, 10, 10, 10, 10, 10, 10, -10,
			-10, 0, 10, 10, 10, 10, 0, -10,
			-10, 5, 5, 10, 10, 5, 5, -10,
			-10, 0, 5, 10, 10, 5, 0, -10,
			-10, 0, 0, 0, 0, 0, 0, -10,
			-20, -10, -10, -10, -10, -10, -10, -20,
		},
		{ // Rook
			0, 0, 0, 5, 5, 0, 0, 0,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			-5, 0, 0, 0, 0, 0, 0, -5,
			5, 10, 10, 10, 10, 10, 10, 5,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
		{ // Queen
			-20, -10, -10, -5, -5, -10, -10, -20,
			-10, 0, 5, 0, 0, 0, 0, -10,
			-10, 5, 5, 5, 5, 5, 0, -10,
			0, 0, 5, 5, 5, 5, 0, -5,
			-5, 0, 5, 5, 5, 5, 0, -5,
			-10, 0, 5, 5, 5, 5, 0, -10,
			-10, 0, 0, 0, 0, 0, 0, -10,
			-20, -10, -10, -5, -5, -10, -10, -20,
		},
		{}, // King
	}
)
