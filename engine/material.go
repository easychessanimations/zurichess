package engine

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	KnownWinScore  = 25000          // All scores strictly greater than KnownWinScore are sure wins.
	KnownLossScore = -KnownWinScore // All scores strictly lower than KnownLossScore are sure losses.
	MateScore      = 30000          // MateScore - N is mate in N plies.
	MatedScore     = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  = 32000          // Maximum possible score. -InfinityScore is the minimum possible score.
)

var (
	// sq ^ colorMask[col] is sq from col's POV.
	// Used with PieceSquareTables which are from White's POV.
	colorMask = [3]Square{0x00, 0x38, 0x00}

	// Bonuses and penalties have type int in order to prevent accidental
	// overflows during computation of the position's score.
	GlobalMaterial = Material{
		DoublePawnPenalty: Score{23, 37},
		BishopPairBonus:   Score{27, 57},
		Mobility:          [FigureArraySize]Score{{0, 0}, {0, 0}, {7, 9}, {3, 8}, {5, 7}, {2, 14}, {-11, 0}},
		FigureBonus:       [FigureArraySize]Score{{0, 0}, {22, 148}, {328, 314}, {342, 332}, {455, 594}, {1020, 1041}, {10000, 10000}},

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
	M, E int32
}

func pov(s Score, col Color) Score {
	switch col {
	case White:
		return s
	case Black:
		return s.Neg()
	default:
		return Score{}
	}
}

// Neg returns -s.
func (s Score) Neg() Score {
	return Score{-s.M, -s.E}
}

// Plus returns s + o.
func (s Score) Plus(o Score) Score {
	return Score{s.M + o.M, s.E + o.E}
}

// Minus returns s - o.
func (s Score) Minus(o Score) Score {
	return Score{s.M - o.M, s.E - o.E}
}

// Times return s scaled by t.
func (s Score) Times(t int32) Score {
	return Score{s.M * t, s.E * t}
}

// Material stores the evaluation parameters.
type Material struct {
	DoublePawnPenalty Score
	BishopPairBonus   Score
	Mobility          [FigureArraySize]Score // how much each piece's mobility is worth
	FigureBonus       [FigureArraySize]Score // how much each piece is worth
	// Piece Square Table from White POV.
	// For black the table is flipped, i.e. black index = 0x38 ^ white index.
	// The tables are indexed from SquareA1 to SquareH8.
	PieceSquareTable [FigureArraySize][SquareArraySize]Score
}

// Evaluation evaluates a position.
//
// Evaluation has two parts:
//  - a primitive static score that is incrementally updated every move.
//  - a dynamic score, a more refined score of the position.
type Evaluation struct {
	Static    Score     // static score, i.e. only the figure bonus
	position  *Position // position to evaluate
	material  *Material // evaluation parameters
	pawnTable pawnTable // a cache for pawn evaluation

	piece [PieceArraySize]Score // cached scores for piece
	promo [PieceArraySize]Score // cached scores for promotion
}

// MakeEvaluation returns a new Evaluation object which evaluates
// pos using parameters in mat.
func MakeEvaluation(pos *Position, mat *Material) Evaluation {
	static := Score{}
	for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
		pi := pos.Get(sq)
		static = static.Plus(pov(mat.FigureBonus[pi.Figure()], pi.Color()))
	}
	var piece, promo [PieceArraySize]Score
	for pi := PieceMinValue; pi <= PieceMaxValue; pi++ {
		piece[pi] = pov(mat.FigureBonus[pi.Figure()], pi.Color())
		promo[pi] = pov(mat.FigureBonus[pi.Figure()].Minus(mat.FigureBonus[Pawn]), pi.Color())
	}
	return Evaluation{
		Static:   static,
		position: pos,
		material: mat,
		piece:    piece,
		promo:    promo,
	}
}

// DoMove executes a move and updates the static score.
func (e *Evaluation) DoMove(m Move) {
	e.position.DoMove(m)
	e.Static = e.Static.Minus(e.piece[m.Capture()])
	if m.MoveType() == Promotion {
		e.Static = e.Static.Plus(e.promo[m.Target()])
	}
}

// UndoMove takes back the latest move and updates the static score.
func (e *Evaluation) UndoMove(m Move) {
	e.position.UndoMove(m)
	e.Static = e.Static.Plus(e.piece[m.Capture()])
	if m.MoveType() == Promotion {
		e.Static = e.Static.Minus(e.promo[m.Target()])
	}
}

// pawns computes the pawn structure score of side.
func (e *Evaluation) pawnStructure(us Color) (score Score) {
	// FigureBonus is included in the static score, and thus not added here.
	pos, mat := e.position, e.material // shortcut
	pawns := pos.ByPiece(us, Pawn)
	mask := colorMask[us]
	psqt := mat.PieceSquareTable[Pawn][:]

	for bb := pawns; bb != 0; {
		sq := bb.Pop()
		score = score.Plus(psqt[sq^mask])
		// Award advanced pawns.
		// Penalize double pawns.
		fwd := sq.Bitboard().Forward(us)
		if fwd&pawns != 0 {
			score = score.Minus(mat.DoublePawnPenalty)
		}
	}

	return score
}

// evaluate position for a single side.
//
// The returned score is from White's POV. Pawn features are evaluated part of pawnStructure.
func (e *Evaluation) evaluateSide(us Color) Score {
	// FigureBonus is included in the static score, and thus not added here.
	pos, mat := e.position, e.material // shortcut
	// Exclude squares attacked by enemy pawns from calculating mobility.
	excl := pos.ByColor[us] | pos.PawnThreats(us.Opposite())
	mask := colorMask[us]

	// Award connected bishops.
	score := mat.BishopPairBonus.Times(int32(pos.NumPieces[us][Bishop] / 2))

	all := pos.ByFigure[Pawn]
	for bb := pos.ByPiece(us, Knight); bb != 0; {
		sq := bb.Pop()
		knight := pos.KnightMobility(sq) &^ excl
		score = score.Plus(mat.Mobility[Knight].Times(knight.Popcnt()))
	}
	for bb := pos.ByPiece(us, Bishop); bb != 0; {
		sq := bb.Pop()
		bishop := pos.BishopMobility(sq, all) &^ excl
		score = score.Plus(mat.Mobility[Bishop].Times(bishop.Popcnt()))
		score = score.Plus(mat.PieceSquareTable[Bishop][sq^mask])
	}
	all = pos.ByFigure[Pawn] | pos.ByFigure[Knight] | pos.ByFigure[Bishop]
	for bb := pos.ByPiece(us, Rook); bb != 0; {
		sq := bb.Pop()
		rook := pos.RookMobility(sq, all) &^ excl
		score = score.Plus(mat.Mobility[Rook].Times(rook.Popcnt()))
	}
	for bb := pos.ByPiece(us, Queen); bb != 0; {
		sq := bb.Pop()
		queen := pos.QueenMobility(sq, all) &^ excl
		score = score.Plus(mat.Mobility[Queen].Times(queen.Popcnt()))
	}
	for bb := pos.ByPiece(us, King); bb != 0; {
		sq := bb.Pop()
		king := pos.KingMobility(sq) &^ excl
		score = score.Plus(mat.Mobility[King].Times(king.Popcnt()))
		score = score.Plus(mat.PieceSquareTable[King][sq^mask])
	}

	return score
}

// evaluate returns position's score from White's POV.
func (e *Evaluation) evaluate() Score {
	pos := e.position // shortcut

	// Evaluate pawn structure, possible using a cached score.
	white := pos.ByPiece(White, Pawn)
	black := pos.ByPiece(Black, Pawn)
	score, ok := e.pawnTable.get(white, black)
	if !ok {
		score = e.pawnStructure(White).Minus(e.pawnStructure(Black))
		e.pawnTable.put(white, black, score)
	}

	// Evaluate the remaining pieces.
	score = score.Plus(e.evaluateSide(White)).Minus(e.evaluateSide(Black))
	// Include the static evaluation, too
	return score.Plus(e.Static)
}

// phase returns the score phase between mid game and end game.
//
// phase is determined by the number of pieces left in the game where
// pawn has score 0, knight and bishop 1, rook 2, queen 2.
// See tapered eval: // https://chessprogramming.wikispaces.com/Tapered+Eval
func phase(pos *Position, score Score) int32 {
	total := int32(16*0 + 4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	// curr -= int32(pos.NumPieces[NoColor][Pawn]) * 0
	curr -= int32(pos.NumPieces[NoColor][Knight]) * 1
	curr -= int32(pos.NumPieces[NoColor][Bishop]) * 1
	curr -= int32(pos.NumPieces[NoColor][Rook]) * 2
	curr -= int32(pos.NumPieces[NoColor][Queen]) * 4
	curr = (curr*256 + total/2) / total
	return (score.M*(256-curr) + score.E*curr) / 256
}

// Evaluate evaluates position from White's POV.
//
// Returns a score phased between mid and end game.
// The returned is always between KnowLossScore and KnownWinScore, excluding.
func (e *Evaluation) Evaluate() int16 {
	score := e.evaluate()
	eval := phase(e.position, score)
	if int32(KnownLossScore) >= eval || eval >= int32(KnownWinScore) {
		panic(fmt.Sprintf("score %d (%v) should be between %d and %d",
			eval, score, KnownLossScore, KnownWinScore))
	}
	return int16(eval)
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
