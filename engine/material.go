package engine

import (
	"fmt"
	"strconv"
	"strings"
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

	// See sorterByMvvLva. These bonuses are multiplied by 32.
	moveBonus = [...]int8{32, 16}
)

const (
	// Bonus indexes in moveBonus array.

	hashMove   = iota
	killerMove = iota
)

// Material evaluates a position from static point of view,
// i.e. pieces and their position on the table.
type Material struct {
	// FigureBonus stores how much each piece is worth.
	FigureBonus [FigureArraySize]int

	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable [FigureArraySize][SquareArraySize]int
}

// EvaluatePosition returns positions score from white's POV.
// The returned score is guaranteed to be between -InfinityScore and +InfinityScore.
func (m *Material) EvaluatePosition(pos *Position) int {
	score := 0
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		colScore := 0
		colMask := ColorMask[col]
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			for bb := pos.ByPiece(col, fig); bb != 0; {
				sq := bb.Pop()
				colScore += m.FigureBonus[fig]
				colScore += m.PieceSquareTable[fig][sq^colMask]
			}
		}
		score += ColorWeight[col] * colScore
	}
	return score
}

// EvaluateMove returns move score from white's POV.
// Move score is defined as the difference between position's score
// after and before the move is executed.
func (m *Material) EvaluateMove(move Move) int {
	score := 0
	mask := ColorMask[move.SideToMove()]
	otherMask := ColorMask[move.SideToMove().Opposite()]

	if move.MoveType == Promotion {
		fig := move.Promotion().Figure()
		score -= m.FigureBonus[Pawn]
		score += m.FigureBonus[fig]
		score -= m.PieceSquareTable[Pawn][move.From^mask]
		score += m.PieceSquareTable[fig][move.To^mask]
	} else {
		if move.MoveType == Castling {
			_, start, end := CastlingRook(move.To)
			score -= m.PieceSquareTable[Rook][start^mask]
			score += m.PieceSquareTable[Rook][end^mask]
		}

		fig := move.Piece().Figure()
		score -= m.PieceSquareTable[fig][move.From^mask]
		score += m.PieceSquareTable[fig][move.To^mask]
	}

	if move.Capture != NoPiece {
		fig := move.Capture.Figure()
		score += m.FigureBonus[fig]
		score += m.PieceSquareTable[fig][move.CaptureSquare()^otherMask]
	}
	return ColorWeight[move.SideToMove()] * score
}

var (
	// Theses values were suggested by Tomasz Michniewski as an extremely basic
	// evaluation.  See the the original values were copied from:
	// https://chessprogramming.wikispaces.com/Simplified+evaluation+function

	// MidGameMaterial defines the material values for mid game.
	MidGameMaterial = Material{
		FigureBonus: [FigureArraySize]int{
			0, 100, 315, 325, 475, 975, 10000,
		},
		PieceSquareTable: [FigureArraySize][SquareArraySize]int{
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
				-20, -10, -10, -5, -5, -10, -10, -20},
			{ // King
				20, 30, 10, 0, 0, 10, 30, 20,
				20, 20, 0, 0, 0, 0, 20, 20,
				-10, -20, -20, -20, -20, -20, -20, -10,
				-20, -30, -30, -40, -40, -30, -30, -20,
				-30, -40, -40, -50, -50, -40, -40, -30,
				-30, -40, -40, -50, -50, -40, -40, -30,
				-30, -40, -40, -50, -50, -40, -40, -30,
				-30, -40, -40, -50, -50, -40, -40, -30,
			},
		},
	}

	// EndGameMaterial defines the material values for end game.
	EndGameMaterial = Material{
		FigureBonus: [FigureArraySize]int{
			0, 100, 345, 355, 525, 1000, 10000,
		},
		PieceSquareTable: [FigureArraySize][SquareArraySize]int{

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
			{ // King
				-50, -30, -30, -30, -30, -30, -30, -50,
				-30, -30, 0, 0, 0, 0, -30, -30,
				-30, -10, 20, 30, 30, 20, -10, -30,
				-30, -10, 30, 40, 40, 30, 10, -30,
				-30, -10, 30, 40, 40, 30, -10, -30,
				-30, -10, 20, 30, 30, 20, -10, -30,
				-30, -20, -10, 0, 0, -10, -20, -30,
				-50, -40, -30, -20, -20, -30, -40, -50,
			},
		},
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
