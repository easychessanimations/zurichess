package engine

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	colMask = [3]Square{0x00, 0x00, 0x38}

	// Bonuses and penalties have type int in order to prevent accidental
	// overflows during computation of the position's score.
	// Scores returned directly are int16.

	KnownWinScore  int16 = 25000
	KnownLossScore int16 = -KnownWinScore
	MateScore      int16 = 30000
	InfinityScore  int16 = 32000

	// The original values for PieceSquareTable were suggested by Tomasz Michniewski as an extremely basic
	// evaluation here: https://chessprogramming.wikispaces.com/Simplified+evaluation+function
	//
	// MidGameMaterial defines the material values for mid game.
	MidGameMaterial = Material{
		DoublePawnPenalty: 13,
		BishopPairBonus:   20,
		Mobility:          [FigureArraySize]int{0, 0, 2, 5, 6, 1, -12},
		FigureBonus:       [FigureArraySize]int{0, 17, 332, 308, 385, 941, 10000},
		PieceSquareTable: [FigureArraySize][SquareArraySize]int{
			{}, // NoFigure
			{ // Pawn
				0, 0, 0, 0, 0, 0, 0, 0,
				36, 52, 47, 41, 41, 47, 52, 36,
				39, 53, 46, 54, 54, 46, 53, 39,
				30, 38, 48, 60, 60, 48, 38, 30,
				27, 29, 42, 57, 57, 42, 29, 27,
				51, 80, 41, 82, 82, 41, 80, 51,
				102, 104, 94, 111, 111, 94, 104, 102,
				0, 0, 0, 0, 0, 0, 0, 0,
			},
			{}, // Knight
			{}, // Bishop
			{}, // Rook
			{}, // Queen
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
		DoublePawnPenalty: 49,
		BishopPairBonus:   60,
		Mobility:          [FigureArraySize]int{0, 0, 13, 9, 6, 13, 4},
		FigureBonus:       [FigureArraySize]int{0, 147, 304, 346, 617, 1060, 10000},
		PieceSquareTable: [FigureArraySize][SquareArraySize]int{
			{}, // NoFigure
			{ // Pawn
				0, 0, 0, 0, 0, 0, 0, 0,
				22, 17, 2, -22, -22, 2, 17, 22,
				13, 17, -3, -4, -4, -3, 17, 13,
				23, 21, -13, -4, -4, -13, 21, 23,
				48, 47, 14, 8, 8, 14, 47, 48,
				122, 101, 92, 101, 101, 92, 101, 122,
				124, 108, 119, 100, 100, 119, 108, 124,
				0, 0, 0, 0, 0, 0, 0, 0,
			},
			{}, // Knight
			{}, // Bishop
			{}, // Rook
			{}, // Queen
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
	pawnTable pawnTable

	DoublePawnPenalty int
	BishopPairBonus   int
	Mobility          [FigureArraySize]int

	// FigureBonus stores how much each piece is worth.
	FigureBonus [FigureArraySize]int

	// Piece Square Table from White POV.
	// For black the table is rotated, i.e. black index = 63 - white index.
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable [FigureArraySize][SquareArraySize]int
}

// pawns computes the pawn structure score of side.
// pawns awards chains and penalizes double pawns.
func (m *Material) pawnStructure(pos *Position, side Color) (score int) {
	pawns := pos.ByPiece(side, Pawn)
	mask := colMask[side]
	psqt := m.PieceSquareTable[Pawn][:]

	for bb := pawns; bb != 0; {
		sq := bb.Pop()
		score += m.FigureBonus[Pawn] + psqt[sq^mask]
		fwd := sq.Bitboard().Forward(side)
		if fwd&pawns != 0 {
			score -= m.DoublePawnPenalty
		}
	}

	return score
}

// evaluate position for side.
//
// Pawn features are evaluated part of pawnStructure.
func (m *Material) evaluate(pos *Position, side Color) int {
	// Opposite pawns, one square forward.
	pawns := pos.ByPiece(side.Opposite(), Pawn).Forward(side.Opposite())
	// Squares occupied by current player and those attacked by opposite pawns.
	mask := pos.ByColor[side] | (pawns & ^FileBb(7) << 1) | (pawns & ^FileBb(0) >> 1)
	// All occupied squares.
	all := pos.ByFigure[White] | pos.ByFigure[Black]

	// Award connected bishops.
	score := int(pos.NumPieces[side][Bishop]/2) * m.BishopPairBonus

	for bb := pos.ByPiece(side, Knight); bb != 0; {
		sq := bb.Pop()
		knight := BbKnightAttack[sq] &^ mask
		score += m.FigureBonus[Knight] + knight.Popcnt()*m.Mobility[Knight]
	}
	for bb := pos.ByPiece(side, Bishop); bb != 0; {
		sq := bb.Pop()
		bishop := BishopMagic[sq].Attack(all) &^ mask
		score += m.FigureBonus[Bishop] + bishop.Popcnt()*m.Mobility[Bishop]
	}
	for bb := pos.ByPiece(side, Rook); bb != 0; {
		sq := bb.Pop()
		rook := RookMagic[sq].Attack(all) &^ mask
		score += m.FigureBonus[Rook] + rook.Popcnt()*m.Mobility[Rook]
	}
	for bb := pos.ByPiece(side, Queen); bb != 0; {
		sq := bb.Pop()
		rook := RookMagic[sq].Attack(all) &^ mask
		bishop := BishopMagic[sq].Attack(all) &^ mask
		score += m.FigureBonus[Queen] + (rook|bishop).Popcnt()*m.Mobility[Queen]
	}
	for bb := pos.ByPiece(side, King); bb != 0; {
		sq := bb.Pop()
		king := BbKingAttack[sq] &^ mask
		score += m.FigureBonus[King] + king.Popcnt()*m.Mobility[King]
		score += m.PieceSquareTable[King][sq^colMask[side]]
	}

	return score
}

// Evaluate returns positions score from white's POV.
//
// The returned score is guaranteed to be between -InfinityScore and +InfinityScore.
func (m *Material) Evaluate(pos *Position) int {
	// Evaluate pieces position.
	score := m.evaluate(pos, White)
	score -= m.evaluate(pos, Black)

	// Evaluate pawn structure.
	whitePawns := pos.ByPiece(White, Pawn)
	blackPawns := pos.ByPiece(Black, Pawn)
	pawns, ok := m.pawnTable.get(whitePawns, blackPawns)
	if !ok {
		pawns = m.pawnStructure(pos, White) - m.pawnStructure(pos, Black)
		m.pawnTable.put(whitePawns, blackPawns, pawns)
	}
	score += pawns

	if int(-InfinityScore) > score || score > int(InfinityScore) {
		panic(fmt.Sprintf("score %d should be between %d and %d",
			score, -InfinityScore, +InfinityScore))
	}

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
