package engine

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	// Bonuses and penalties have type int in order to prevent accidental
	// overflows during computation of the position's score.
	// Scores returned directly are int16.

	KnownWinScore int16 = 20000
	MateScore     int16 = 30000
	InfinityScore int16 = 32000

	// The original values for PieceSquareTable were suggested by Tomasz Michniewski as an extremely basic
	// evaluation here: https://chessprogramming.wikispaces.com/Simplified+evaluation+function
	//
	// MidGameMaterial defines the material values for mid game.
	MidGameMaterial = Material{
		BishopPairBonus:   40,
		PawnChainBonus:    8,
		DoublePawnPenalty: 13,
		FigureBonus:       [FigureArraySize]int{0, 100, 335, 325, 440, 975, 10000},
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
		BishopPairBonus:   40,
		PawnChainBonus:    8,
		DoublePawnPenalty: 13,
		FigureBonus:       [FigureArraySize]int{0, 115, 315, 355, 590, 1000, 10000},
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

// Material evaluates a position from static point of view,
// i.e. pieces and their position on the table.
type Material struct {
	BishopPairBonus   int
	PawnChainBonus    int
	DoublePawnPenalty int

	// FigureBonus stores how much each piece is worth.
	FigureBonus [FigureArraySize]int

	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable [FigureArraySize][SquareArraySize]int
}

// pawns computes the pawn structure score of side.
// pawns awards chains and penalizes double pawns.
func (m *Material) pawnStructure(pos *Position, side Color) int {
	pawns := pos.ByPiece(side, Pawn)
	forward := pawns
	if side == White {
		forward >>= 8
	} else {
		forward <<= 8
	}

	cs := (pawns & ((forward &^ FileBb(7)) << 1)).Popcnt()
	cs += (pawns & ((forward &^ FileBb(0)) >> 1)).Popcnt()
	ds := (pawns & forward).Popcnt()
	return cs*m.PawnChainBonus - ds*m.DoublePawnPenalty
}

// Evaluate returns positions score from white's POV.
// The returned score is guaranteed to be between -InfinityScore and +InfinityScore.
func (m *Material) Evaluate(pos *Position) int {
	score := 0

	// Award pieces on the table.
	for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
		if pos.NumPieces[NoColor][fig] == 0 {
			continue
		}

		score += int(pos.NumPieces[White][fig]-pos.NumPieces[Black][fig]) * m.FigureBonus[fig]
		psqt := m.PieceSquareTable[fig][:]
		for bb := pos.ByPiece(White, fig); bb != 0; {
			sq := bb.Pop()
			score += psqt[sq^0x00]
		}
		for bb := pos.ByPiece(Black, fig); bb != 0; {
			sq := bb.Pop()
			score -= psqt[sq^0x38]
		}
	}

	// Award connected bishops.
	score += int(pos.NumPieces[White][Bishop]/2-pos.NumPieces[Black][Bishop]/2) * m.BishopPairBonus

	// Award pawn structure.
	score += m.pawnStructure(pos, White)
	score -= m.pawnStructure(pos, Black)

	return score
}

// phase returns a current, total pair which is the linear progress
// between mid game and end game.
//
// phase is determined by the number of pieces left in the game where
// pawn has score 0, knight and bishop 1, rook 2, queen 2.
// See tapered eval: // https://chessprogramming.wikispaces.com/Tapered+Eval
func phase(pos *Position) (int, int) {
	totalPhase := 16*0 + 4*1 + 4*1 + 4*2 + 2*4
	currPhase := totalPhase
	currPhase -= int(pos.NumPieces[NoColor][Pawn]) * 0
	currPhase -= int(pos.NumPieces[NoColor][Knight]) * 1
	currPhase -= int(pos.NumPieces[NoColor][Bishop]) * 1
	currPhase -= int(pos.NumPieces[NoColor][Rook]) * 2
	currPhase -= int(pos.NumPieces[NoColor][Queen]) * 4
	currPhase = (currPhase*256 + totalPhase/2) / totalPhase
	return currPhase, 256
}

// Evaluate evaluates position.
func Evaluate(pos *Position) int16 {
	midGame := MidGameMaterial.Evaluate(pos)
	endGame := EndGameMaterial.Evaluate(pos)
	curr, total := phase(pos)
	score := (midGame*(total-curr) + endGame*curr) / total
	return int16(score)
}

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
