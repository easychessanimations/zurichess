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

	KnownWinScore  int16 = 25000          // All scores strictly greater than KnownWinScore are sure wins.
	KnownLossScore int16 = -KnownWinScore // All scores strictly lower than KnownLossScore are sure losses.
	MateScore      int16 = 30000          // MateScore - N is mate in N plies.
	MatedScore     int16 = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  int16 = 32000          // Maximum possible score. -InfinityScore is the minimum possible score.

	Evaluation = Material{
		DoublePawnPenalty: Score{11, 53},
		BishopPairBonus:   Score{24, 58},
		Mobility:          [FigureArraySize]Score{{0, 0}, {0, 0}, {2, 13}, {5, 9}, {6, 6}, {1, 13}, {-10, -1}},
		FigureBonus:       [FigureArraySize]Score{{0, 0}, {17, 147}, {337, 304}, {313, 346}, {389, 618}, {941, 1059}, {10000, 10000}},

		PieceSquareTable: [FigureArraySize][SquareArraySize]Score{
			{}, // NoFigure
			{ // Pawn
				{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
				{36, 22}, {52, 17}, {47, 2}, {41, -22}, {41, -22}, {47, 2}, {52, 17}, {36, 22},
				{39, 13}, {53, 17}, {46, -3}, {54, -4}, {54, -4}, {46, -3}, {53, 17}, {39, 13},
				{30, 23}, {38, 21}, {48, -13}, {60, -4}, {60, -4}, {48, -13}, {38, 21}, {30, 23},
				{27, 48}, {29, 47}, {42, 14}, {57, 8}, {57, 8}, {42, 14}, {29, 47}, {27, 48},
				{51, 122}, {80, 101}, {41, 92}, {82, 101}, {82, 101}, {41, 92}, {80, 101}, {51, 122},
				{102, 124}, {104, 108}, {94, 119}, {111, 100}, {111, 100}, {94, 119}, {104, 108}, {102, 124},
				{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
			},
			{}, // Knight
			{ // Bishop
				{-7, -1}, {2, -1}, {-5, -13}, {0, -3}, {0, -3}, {-5, -13}, {2, -1}, {-7, -1},
				{1, -1}, {-1, 2}, {-6, 1}, {-4, -2}, {-4, -2}, {-6, 1}, {-1, 2}, {1, -1},
				{-1, 16}, {4, 14}, {18, 1}, {1, 7}, {1, 7}, {18, 1}, {4, 14}, {-1, 16},
				{-12, 5}, {2, 6}, {1, 5}, {26, -9}, {26, -9}, {1, 5}, {2, 6}, {-12, 5},
				{-12, 5}, {2, 6}, {1, 5}, {26, -9}, {26, -9}, {1, 5}, {2, 6}, {-12, 5},
				{-1, 16}, {4, 14}, {18, 1}, {1, 7}, {1, 7}, {18, 1}, {4, 14}, {-1, 16},
				{1, -1}, {-1, 2}, {-6, 1}, {-4, -2}, {-4, -2}, {-6, 1}, {-1, 2}, {1, -1},
				{-7, -1}, {2, -1}, {-5, -13}, {0, -3}, {0, -3}, {-5, -13}, {2, -1}, {-7, -1},
			},
			{}, // Rook
			{}, // Queen
			{ // King
				// The values for King were suggested by Tomasz Michniewski.
				// Numbers were copied from: https://chessprogramming.wikispaces.com/Simplified+evaluation+function
				{20, -50}, {30, -30}, {10, -30}, {0, -30}, {0, -30}, {10, -30}, {30, -30}, {20, -50},
				{20, -30}, {20, -30}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {20, -30}, {20, -30},
				{-10, -30}, {-20, -10}, {-20, 20}, {-20, 30}, {-20, 30}, {-20, 20}, {-20, -10}, {-10, -30},
				{-20, -30}, {-30, -10}, {-30, 30}, {-40, 40}, {-40, 40}, {-30, 30}, {-30, 10}, {-20, -30},
				{-30, -30}, {-40, -10}, {-40, 30}, {-50, 40}, {-50, 40}, {-40, 30}, {-40, -10}, {-30, -30},
				{-30, -30}, {-40, -10}, {-40, 20}, {-50, 30}, {-50, 30}, {-40, 20}, {-40, -10}, {-30, -30},
				{-30, -30}, {-40, -20}, {-40, -10}, {-50, 0}, {-50, 0}, {-40, -10}, {-40, -20}, {-30, -30},
				{-30, -50}, {-40, -40}, {-40, -30}, {-50, -20}, {-50, -20}, {-40, -30}, {-40, -40}, {-30, -50},
			},
		},
	}
)

// Score represents a pair of mid game and end game scores.
type Score struct {
	M, E int
}

func (s Score) Plus(o Score) Score {
	return Score{s.M + o.M, s.E + o.E}
}

func (s Score) Minus(o Score) Score {
	return Score{s.M - o.M, s.E - o.E}
}

func (s Score) Times(t int) Score {
	return Score{s.M * t, s.E * t}
}

// Material evaluates a position from static point of view,
// i.e. pieces and their position on the table.
type Material struct {
	pawnTable pawnTable // a cache for pawn evaluation

	DoublePawnPenalty Score
	BishopPairBonus   Score
	Mobility          [FigureArraySize]Score // how much each piece's mobility is worth
	FigureBonus       [FigureArraySize]Score // how much each piece is worth

	// Piece Square Table from White POV.
	// For black the table is flipped, i.e. black index = 0x38 ^ white index.
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable [FigureArraySize][SquareArraySize]Score
}

// pawns computes the pawn structure score of side.
// pawns awards chains and penalizes double pawns.
func (m *Material) pawnStructure(pos *Position, side Color) (score Score) {
	pawns := pos.ByPiece(side, Pawn)
	mask := colMask[side]
	psqt := m.PieceSquareTable[Pawn][:]

	for bb := pawns; bb != 0; {
		sq := bb.Pop()
		score = score.Plus(m.FigureBonus[Pawn])
		score = score.Plus(psqt[sq^mask])
		fwd := sq.Bitboard().Forward(side)
		if fwd&pawns != 0 {
			score = score.Minus(m.DoublePawnPenalty)
		}
	}

	return score
}

// evaluate position for side.
//
// Pawn features are evaluated part of pawnStructure.
func (m *Material) evaluate(pos *Position, side Color) Score {
	// Exclude squares attacked by enemy pawns from calculating mobility.
	excl := pos.ByColor[side] | pos.PawnThreats(side.Opposite())
	all := pos.ByFigure[Pawn] | pos.ByFigure[Knight]
	mask := colMask[side]

	// Award connected bishops.
	score := m.BishopPairBonus.Times(int(pos.NumPieces[side][Bishop] / 2))

	for bb := pos.ByPiece(side, Knight); bb != 0; {
		sq := bb.Pop()
		knight := pos.KnightMobility(sq) &^ excl
		score = score.Plus(m.FigureBonus[Knight])
		score = score.Plus(m.Mobility[Knight].Times(knight.Popcnt()))
	}
	for bb := pos.ByPiece(side, Bishop); bb != 0; {
		sq := bb.Pop()
		bishop := pos.BishopMobility(sq, all) &^ excl
		score = score.Plus(m.FigureBonus[Bishop])
		score = score.Plus(m.Mobility[Bishop].Times(bishop.Popcnt()))
		score = score.Plus(m.PieceSquareTable[Bishop][sq^mask])
	}
	for bb := pos.ByPiece(side, Rook); bb != 0; {
		sq := bb.Pop()
		rook := pos.RookMobility(sq, all) &^ excl
		score = score.Plus(m.FigureBonus[Rook])
		score = score.Plus(m.Mobility[Rook].Times(rook.Popcnt()))
	}
	for bb := pos.ByPiece(side, Queen); bb != 0; {
		sq := bb.Pop()
		queen := pos.QueenMobility(sq, all) &^ excl
		score = score.Plus(m.FigureBonus[Queen])
		score = score.Plus(m.Mobility[Queen].Times(queen.Popcnt()))
	}
	for bb := pos.ByPiece(side, King); bb != 0; {
		sq := bb.Pop()
		king := pos.KingMobility(sq) &^ excl
		score = score.Plus(m.FigureBonus[King])
		score = score.Plus(m.Mobility[King].Times(king.Popcnt()))
		score = score.Plus(m.PieceSquareTable[King][sq^mask])
	}

	return score
}

// Evaluate returns positions score from white's POV.
//
// The returned score is guaranteed to be between -InfinityScore and +InfinityScore.
func (m *Material) Evaluate(pos *Position) Score {
	// Evaluate pieces position.
	score := m.evaluate(pos, White).Minus(m.evaluate(pos, Black))

	// Evaluate pawn structure.
	whitePawns := pos.ByPiece(White, Pawn)
	blackPawns := pos.ByPiece(Black, Pawn)
	pawns, ok := m.pawnTable.get(whitePawns, blackPawns)
	if !ok {
		pawns = m.pawnStructure(pos, White).Minus(m.pawnStructure(pos, Black))
		m.pawnTable.put(whitePawns, blackPawns, pawns)
	}
	score = score.Plus(pawns)

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
//
// Returned score is a tapered between MidGameMaterial and EndGameMaterial.
func Evaluate(pos *Position) int16 {
	score := Evaluation.Evaluate(pos)
	curr, total := phase(pos)
	phased := (score.M*(total-curr) + score.E*curr) / total

	if int(-InfinityScore) > phased || phased > int(InfinityScore) {
		panic(fmt.Sprintf("score %d should be between %d and %d",
			score, -InfinityScore, +InfinityScore))
	}

	return int16(phased)
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
