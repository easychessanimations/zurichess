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
		ConnectedPawn:   Score{4, 13},
		DoublePawn:      Score{12, 14},
		IsolatedPawn:    Score{9, -2},
		PassedPawn:      [8]Score{{0, 0}, {0, 0}, {0, 0}, {20, 64}, {27, 98}, {62, 145}, {101, 192}, {0, 0}},
		BishopPairBonus: Score{36, 38},
		Mobility:        [FigureArraySize]Score{{0, 0}, {0, 0}, {8, 7}, {3, 8}, {7, 5}, {2, 5}, {-5, -4}},
		FigureBonus:     [FigureArraySize]Score{{0, 0}, {0, 0}, {311, 285}, {332, 308}, {408, 581}, {1036, 1054}, {0, 0}},

		PieceSquareTable: [FigureArraySize][SquareArraySize]Score{
			{}, // NoFigure
			{ // Pawn
				{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
				{70, 104}, {54, 118}, {56, 109}, {67, 67}, {64, 79}, {73, 107}, {75, 97}, {69, 84},
				{59, 118}, {50, 110}, {61, 105}, {59, 96}, {66, 102}, {59, 115}, {75, 98}, {70, 102},
				{53, 126}, {57, 118}, {61, 101}, {78, 88}, {77, 102}, {66, 92}, {50, 109}, {58, 104},
				{52, 144}, {43, 122}, {66, 108}, {72, 91}, {77, 79}, {75, 105}, {46, 130}, {60, 121},
				{47, 159}, {61, 131}, {53, 122}, {84, 104}, {72, 86}, {54, 107}, {84, 129}, {70, 134},
				{64, 136}, {83, 119}, {71, 97}, {69, 77}, {105, 98}, {57, 115}, {91, 103}, {95, 120},
				{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
			},
			{}, // Knight
			{}, // Bishop
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
	ConnectedPawn   Score
	DoublePawn      Score
	IsolatedPawn    Score
	PassedPawn      [8]Score               // score of each passed pawn, indexed by rank
	BishopPairBonus Score                  // how much a pair of bishop is worth
	Mobility        [FigureArraySize]Score // how much each piece's mobility is worth
	FigureBonus     [FigureArraySize]Score // how much each piece is worth

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
	position  *Position // position to evaluate
	material  *Material // evaluation parameters
	pawnTable pawnTable // a cache for pawn evaluation

	piece [PieceArraySize]Score // cached scores for piece
	promo [PieceArraySize]Score // cached scores for promotion
}

// MakeEvaluation returns a new Evaluation object which evaluates
// pos using parameters in mat.
func MakeEvaluation(pos *Position, mat *Material) Evaluation {
	var piece, promo [PieceArraySize]Score
	for pi := PieceMinValue; pi <= PieceMaxValue; pi++ {
		piece[pi] = pov(mat.FigureBonus[pi.Figure()], pi.Color())
		promo[pi] = pov(mat.FigureBonus[pi.Figure()].Minus(mat.FigureBonus[Pawn]), pi.Color())
	}
	return Evaluation{
		position: pos,
		material: mat,
		piece:    piece,
		promo:    promo,
	}
}

// pawns computes the pawn structure score of side.
func (e *Evaluation) pawnStructure(us Color) (score Score) {
	// TODO: Evaluate double pawns that are not next to each other.
	// TODO: Evaluate opposed pawns.
	// TODO: Evaluate larger pawn structures.

	pos, mat := e.position, e.material // shortcut
	mask := colorMask[us]

	// Award pawns based on the Hans Berliner's system.
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)

	// From white's POV (P - white pawn, p - black pawn).
	// block   wings
	// ....... .....
	// .....P. .....
	// .....x. .....
	// ..p..x. .....
	// .xxx.x. .xPx.
	// .xxx.x. .....
	// .xxx.x. .....
	// .xxx.x. .....
	block := East(theirs) | theirs | West(theirs)
	wings := East(ours) | West(ours)
	double := Bitboard(0)
	if us == White {
		block = SouthSpan(block) | SouthSpan(ours)
		double = ours & South(ours)
	} else /* if us == Black */ {
		block = NorthSpan(block) | NorthSpan(ours)
		double = ours & North(ours)
	}

	isolated := ours &^ Fill(wings)                           // no pawn on the adjacent files
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	passed := ours &^ block                                   // no pawn in front and no enemy on the adjacent files

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		rank := (sq ^ mask).Rank() // from our POV

		ps := mat.PieceSquareTable[Pawn][sq^mask]
		if passed.Has(sq) {
			ps = ps.Plus(mat.PassedPawn[rank])
		}
		if connected.Has(sq) { // bonus added to both pawns
			ps = ps.Plus(mat.ConnectedPawn)
		}
		if double.Has(sq) {
			ps = ps.Minus(mat.DoublePawn)
		}
		if isolated.Has(sq) {
			ps = ps.Minus(mat.IsolatedPawn)
		}

		score = score.Plus(ps)
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
		score = score.Plus(mat.FigureBonus[Knight])
		score = score.Plus(mat.Mobility[Knight].Times(knight.Popcnt()))
	}
	for bb := pos.ByPiece(us, Bishop); bb != 0; {
		sq := bb.Pop()
		bishop := pos.BishopMobility(sq, all) &^ excl
		score = score.Plus(mat.FigureBonus[Bishop])
		score = score.Plus(mat.Mobility[Bishop].Times(bishop.Popcnt()))
	}
	all = pos.ByFigure[Pawn] | pos.ByFigure[Knight] | pos.ByFigure[Bishop]
	for bb := pos.ByPiece(us, Rook); bb != 0; {
		sq := bb.Pop()
		rook := pos.RookMobility(sq, all) &^ excl
		score = score.Plus(mat.FigureBonus[Rook])
		score = score.Plus(mat.Mobility[Rook].Times(rook.Popcnt()))
	}
	for bb := pos.ByPiece(us, Queen); bb != 0; {
		sq := bb.Pop()
		queen := pos.QueenMobility(sq, all) &^ excl
		score = score.Plus(mat.FigureBonus[Queen])
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
	return score.Plus(e.evaluateSide(White)).Minus(e.evaluateSide(Black))
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
