// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  int32 = 25000          // All scores strictly greater than KnownWinScore are sure wins.
	KnownLossScore int32 = -KnownWinScore // All scores strictly lower than KnownLossScore are sure losses.
	MateScore      int32 = 30000          // MateScore - N is mate in N plies.
	MatedScore     int32 = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  int32 = 32000          // Maximum possible score. -InfinityScore is the minimum possible score.
)

var (
	// All evaluation parameters.

	ConnectedPawn = Score{11, 2}
	DoublePawn    = Score{3, 19}
	IsolatedPawn  = Score{5, 3}
	// score of each passed pawn, indexed by rank from color's pov.
	PassedPawn = [8]Score{{0, 0}, {0, 0}, {0, 0}, {0, 0}, {23, 65}, {38, 113}, {58, 153}, {0, 0}}
	BishopPair = Score{28, 45}
	// award pawn shelter in front of the king
	KingShelter = Score{20, -10}
	// how much each piece's mobility is worth
	Mobility = [FigureArraySize]Score{{0, 0}, {2, 20}, {8, 8}, {6, 7}, {7, 7}, {2, 5}, {-11, 0}}
	// how much each piece is worth
	FigureBonus = [FigureArraySize]Score{{0, 0}, {55, 120}, {325, 316}, {341, 346}, {454, 589}, {1110, 1085}, {20000, 20000}}
	// Piece Square Table from White POV.
	// The tables are indexed from SquareA1 to SquareH8,
	// but should be accessed as PieceSquareTable[fig][sq.POV(us)].
	PieceSquareTable = [FigureArraySize][SquareArraySize]Score{
		{}, // NoFigure
		{ // Pawn
			{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
			{1, 10}, {4, 7}, {2, -1}, {3, 1}, {-5, 5}, {16, 18}, {14, 11}, {17, -2},
			{2, 11}, {0, 7}, {0, -1}, {-2, 7}, {7, 11}, {3, 9}, {12, 8}, {5, 7},
			{0, 33}, {1, 13}, {11, 19}, {17, 7}, {15, 4}, {-2, 10}, {-2, 20}, {4, 19},
			{4, 47}, {5, 33}, {4, 11}, {21, 9}, {14, 4}, {19, 5}, {3, 28}, {-1, 29},
			{17, 71}, {4, 45}, {52, 24}, {17, 48}, {27, 29}, {15, 37}, {37, 33}, {15, 40},
			{30, 69}, {47, 66}, {26, 30}, {3, 20}, {40, 45}, {22, 53}, {25, 67}, {12, 70},
			{0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0}, {0, 0},
		},
		{}, // Knight
		{}, // Bishop
		{}, // Rook
		{}, // Queen
		{ // King
			// The values for King were suggested by Tomasz Michniewski.
			// Numbers were copied from= https://chessprogramming.wikispaces.com/Simplified+evaluation+function
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

// Score represents a pair of mid game and end game scores.
type Score struct {
	M, E int32
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

// Evaluation evaluates a position.
type Evaluation struct {
	position  *Position                 // position to evaluate
	pawnTable [ColorArraySize]pawnTable // a cache for pawn evaluation
}

// MakeEvaluation returns a new Evaluation object which evaluates
// pos using parameters in
func MakeEvaluation(pos *Position) Evaluation {
	return Evaluation{position: pos}
}

// pawns computes the pawn structure score of side.
func (e *Evaluation) pawnStructure(us Color) (score Score) {
	pos := e.position // shortcut
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)

	if score, ok := e.pawnTable[us].get(ours, theirs); ok {
		// Use a cached value if available.
		return score
	}

	// TODO: Evaluate double pawns that are not next to each other.
	// TODO: Evaluate opposed pawns.
	// TODO: Evaluate larger pawn structures.
	// TODO: Penalize backward pawns to encourage pawn advancement.

	score = FigureBonus[Pawn].Times(ours.Popcnt())

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
		povSq := sq.POV(us)

		ps := PieceSquareTable[Pawn][povSq]
		if passed.Has(sq) {
			ps = ps.Plus(PassedPawn[povSq.Rank()])
		}
		if connected.Has(sq) {
			// The bonus is added to both pawns.
			// TODO: Add to a single pawn to encourage longer chains.
			ps = ps.Plus(ConnectedPawn)
		}
		if double.Has(sq) {
			ps = ps.Minus(DoublePawn)
		}
		if isolated.Has(sq) {
			ps = ps.Minus(IsolatedPawn)
		}

		score = score.Plus(ps)
	}

	e.pawnTable[us].put(ours, theirs, score)
	return score
}

// evaluate position for a single side.
//
// The returned score is from our POV.
// Pawn features are evaluated part of pawnStructure.
func (e *Evaluation) evaluateSide(us Color) Score {
	// FigureBonus is included in the static score, and thus not added here.
	pos := e.position // shortcut
	// Exclude squares attacked by enemy pawns from calculating mobility.
	excl := pos.ByColor[us] | pos.PawnThreats(us.Opposite())

	// Award pawn structure.
	score := e.pawnStructure(us)

	// Award connect bishop pair.
	if bishops := pos.ByPiece(us, Bishop); bishops.HasMoreThanOne() {
		score = score.Plus(BishopPair)
	}

	// Award pawn forward mobility.
	// Forward mobility is important especially in the end game to
	// allow pawns to promote.
	// Pawn psqt and figure bonus are computed by pawnStructure.
	mobility := Bitboard(0)
	ours := pos.ByPiece(us, Pawn)
	mobility = Forward(us, ours) &^ (pos.ByColor[White] | pos.ByColor[Black])
	score = score.Plus(Mobility[Pawn].Times(mobility.Popcnt()))

	// Knight and bishop mobility considers only pawns.
	// We exclude minors and majors because they enable tactics.
	all := pos.ByFigure[Pawn]
	for bb := pos.ByPiece(us, Knight); bb != 0; {
		sq := bb.Pop()
		knight := pos.KnightMobility(sq) &^ excl
		score = score.Plus(FigureBonus[Knight])
		score = score.Plus(Mobility[Knight].Times(knight.Popcnt()))
	}
	for bb := pos.ByPiece(us, Bishop); bb != 0; {
		sq := bb.Pop()
		bishop := pos.BishopMobility(sq, all) &^ excl
		score = score.Plus(FigureBonus[Bishop])
		score = score.Plus(Mobility[Bishop].Times(bishop.Popcnt()))
	}

	// Rook and Queen mobility considers only pawns and minor pieces.
	// We exclude majors because they enable tactics.
	all = pos.ByFigure[Pawn] | pos.ByFigure[Knight] | pos.ByFigure[Bishop]
	for bb := pos.ByPiece(us, Rook); bb != 0; {
		sq := bb.Pop()
		rook := pos.RookMobility(sq, all) &^ excl
		score = score.Plus(FigureBonus[Rook])
		score = score.Plus(Mobility[Rook].Times(rook.Popcnt()))
	}
	for bb := pos.ByPiece(us, Queen); bb != 0; {
		sq := bb.Pop()
		queen := pos.QueenMobility(sq, all) &^ excl
		score = score.Plus(FigureBonus[Queen])
		score = score.Plus(Mobility[Queen].Times(queen.Popcnt()))
	}
	for bb := pos.ByPiece(us, King); bb != 0; {
		sq := bb.Pop()
		king := pos.KingMobility(sq) &^ excl
		score = score.Plus(Mobility[King].Times(king.Popcnt()))
		score = score.Plus(PieceSquareTable[King][sq.POV(us)])
	}

	// Penalize broken shield in front of the king.
	// Ignore shelter if we entered late game.
	them := us.Opposite()
	if pos.ByPiece(them, Queen) != 0 {
		pawns := pos.ByPiece(us, Pawn)
		king := pos.ByPiece(us, King)
		file := king.AsSquare().File()

		// TODO: Should we include adjacent pawns in the computation?
		if us == White {
			king = NorthSpan(king)
		} else /* if us == Black */ {
			king = SouthSpan(king)
		}

		if file > 0 && West(king)&pawns == 0 {
			score = score.Minus(KingShelter)
		}
		if king&pawns == 0 {
			score = score.Minus(KingShelter.Times(2))
		}
		if file < 7 && East(king)&pawns == 0 {
			score = score.Minus(KingShelter)
		}
	}

	return score
}

// phase returns the score phase between mid game and end game.
//
// phase is determined by the number of pieces left in the game where
// pawn has score 0, knight and bishop 1, rook 2, queen 2.
// See tapered eval for explanation and original code:
// https://chessprogramming.wikispaces.com/Tapered+Eval
func phase(pos *Position, score Score) int32 {
	total := int32(16*0 + 4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	// curr -= pos.ByFigure[Pawn].Popcnt() * 0
	curr -= pos.ByFigure[Knight].Popcnt() * 1
	curr -= pos.ByFigure[Bishop].Popcnt() * 1
	curr -= pos.ByFigure[Rook].Popcnt() * 2
	curr -= pos.ByFigure[Queen].Popcnt() * 4
	curr = (curr*256 + total/2) / total
	return (score.M*(256-curr) + score.E*curr) / 256
}

// Evaluate evaluates position from White's POV.
//
// Returns a score phased between mid and end game.
// The returned is always between KnowLossScore and KnownWinScore, excluding.
func (e *Evaluation) Evaluate() int32 {
	score := e.evaluateSide(White).Minus(e.evaluateSide(Black))
	eval := phase(e.position, score)
	if KnownLossScore >= eval || eval >= KnownWinScore {
		panic(fmt.Sprintf("score %d (%v) should be between %d and %d",
			eval, score, KnownLossScore, KnownWinScore))
	}
	return eval
}
