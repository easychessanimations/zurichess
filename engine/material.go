// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  int32 = 25000000       // All scores strictly greater than KnownWinScore are sure wins.
	KnownLossScore int32 = -KnownWinScore // All scores strictly lower than KnownLossScore are sure losses.
	MateScore      int32 = 30000000       // MateScore - N is mate env N plies.
	MatedScore     int32 = -MateScore     // MatedScore + N is mated env N plies.
	InfinityScore  int32 = 32000000       // Maximum possible score. -InfinityScore is the minimum possible score.
)

var (
	// All weights under one array for easy handling.
	Weights = [92]Score{
		{M: 9528, E: 5302}, {M: 2287, E: 6992}, {M: 36156, E: 34821}, {M: 39458, E: 38848}, {M: 53014, E: 69179}, {M: 123028, E: 130670}, {M: 8188, E: 928}, {M: 4677, E: 8736}, {M: 660, E: 2663}, {M: 963, E: 1085}, {M: 636, E: 810}, {M: 916, E: 683}, {M: 240, E: 811}, {M: -307, E: -820}, {M: 4415, E: 5679}, {M: 3527, E: 5838}, {M: 4358, E: 4210}, {M: 5882, E: 2587}, {M: 4646, E: 3148}, {M: 6405, E: 3815}, {M: 5201, E: 3322}, {M: 3441, E: 3301}, {M: 4257, E: 5403}, {M: 3409, E: 5351}, {M: 4274, E: 4153}, {M: 4666, E: 3076}, {M: 5548, E: 4199}, {M: 4496, E: 4554}, {M: 5624, E: 3890}, {M: 3160, E: 4938}, {M: 3889, E: 7450}, {M: 3392, E: 5893}, {M: 4735, E: 4090}, {M: 6233, E: 3384}, {M: 5325, E: 3982}, {M: 3688, E: 3775}, {M: 3027, E: 4682}, {M: 2349, E: 5520}, {M: 3513, E: 9900}, {M: 2986, E: 8197}, {M: 4643, E: 5050}, {M: 5326, E: 3096}, {M: 4229, E: 4907}, {M: 3894, E: 4308}, {M: 2662, E: 6412}, {M: 1881, E: 7145}, {M: 5260, E: 11547}, {M: 4099, E: 10166}, {M: 7463, E: 4718}, {M: 7356, E: 1850}, {M: 6092, E: 1304}, {M: 7808, E: 4555}, {M: 4999, E: 5782}, {M: 5641, E: 9024}, {M: 12327, E: 12794}, {M: 5310, E: 15980}, {M: -1942, E: 13408}, {M: 5674, E: 9415}, {M: 4769, E: 10355}, {M: 9523, E: 8076}, {M: 499, E: 13036}, {M: 1184, E: 12382}, {M: 6050, E: 2850}, {M: 368, E: 2243}, {M: 230, E: 2994}, {M: -544, E: 6696}, {M: 2581, E: 9085}, {M: 3861, E: 16186}, {M: 8709, E: 13238}, {M: 8299, E: 7906}, {M: 5648, E: -1489}, {M: 3640, E: 4995}, {M: 2744, E: 6340}, {M: -1438, E: 7847}, {M: -8671, E: 10525}, {M: -2637, E: 10978}, {M: 16173, E: 7409}, {M: 27546, E: -204}, {M: 8162, E: 507}, {M: 9360, E: 6191}, {M: 7620, E: 7566}, {M: 2334, E: 9388}, {M: 4908, E: 8552}, {M: 2026, E: 8981}, {M: 8836, E: 6227}, {M: 7052, E: 2295}, {M: 1010, E: 287}, {M: -467, E: 756}, {M: -872, E: -860}, {M: 1456, E: 1022}, {M: -1991, E: -560}, {M: 2902, E: 6268},
	}

	// Named chunks of Weights
	wFigure     [FigureArraySize]Score
	wMobility   [FigureArraySize]Score
	wPawn       [48]Score
	wPassedPawn [8]Score
	wKingRank   [8]Score
	wKingFile   [8]Score
	wFlags      [6]Score // see flags defined below
)

const (
	fConnectedPawn = iota
	fDoublePawn
	fIsolatedPawn
	fPawnThreat
	fKingShelter
	fBishopPair
)

func init() {
	initWeights()

	chunk := func(w []Score, out []Score) []Score {
		copy(out, w)
		return w[len(out):]
	}

	w := Weights[:]
	w = chunk(w, wFigure[:])
	w = chunk(w, wMobility[:])
	w = chunk(w, wPawn[:])
	w = chunk(w, wPassedPawn[:])
	w = chunk(w, wKingRank[:])
	w = chunk(w, wKingFile[:])
	w = chunk(w, wFlags[:])

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}
}

func evaluatePawns(pos *Position, us Color, eval *Eval) {
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
	passed := ours &^ block                                   // no pawn env front and no enemy on the adjacent files

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.Add(wFigure[Pawn])
		eval.Add(wPawn[povSq-8])

		if passed.Has(sq) {
			eval.Add(wPassedPawn[rank])
		}
		if connected.Has(sq) {
			eval.Add(wFlags[fConnectedPawn])
		}
		if double.Has(sq) {
			eval.Add(wFlags[fDoublePawn])
		}
		if isolated.Has(sq) {
			eval.Add(wFlags[fIsolatedPawn])
		}
	}
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	evaluatePawnsCached(pos, us, eval)
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	// Pawn
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	eval.AddN(wMobility[Pawn], mobility.Popcnt())
	mobility = pos.PawnThreats(us) & pos.ByColor[us.Opposite()]
	eval.AddN(wFlags[fPawnThreat], mobility.Popcnt())

	// Knight
	excl := pos.ByPiece(us, Pawn) | pos.PawnThreats(them)
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Knight])
		mobility := pos.KnightMobility(sq) &^ excl
		eval.AddN(wMobility[Knight], mobility.Popcnt())
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Bishop])
		mobility := pos.BishopMobility(sq, all) &^ excl
		eval.AddN(wMobility[Bishop], mobility.Popcnt())
		numBishops++
	}
	eval.AddN(wFlags[fBishopPair], numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Rook])
		mobility := pos.RookMobility(sq, all) &^ excl
		eval.AddN(wMobility[Rook], mobility.Popcnt())
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Queen])
		mobility := pos.QueenMobility(sq, all) &^ excl
		eval.AddN(wMobility[Queen], mobility.Popcnt())
	}

	// King
	for bb := pos.ByPiece(us, King); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[King])
		mobility := pos.KingMobility(sq) &^ excl
		eval.AddN(wMobility[King], mobility.Popcnt())

		sq = sq.POV(us)
		eval.Add(wKingFile[sq.File()])
		eval.Add(wKingRank[sq.Rank()])
	}

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
			eval.Add(wFlags[fKingShelter])
		}
		if king&pawns == 0 {
			eval.AddN(wFlags[fKingShelter], 2)
		}
		if file < 7 && East(king)&pawns == 0 {
			eval.Add(wFlags[fKingShelter])
		}
	}
}

// evaluatePosition evalues position.
func EvaluatePosition(pos *Position, eval *Eval) {
	eval.Make(pos)
	evaluateSide(pos, Black, eval)
	eval.Neg()
	evaluateSide(pos, White, eval)
}

// Evaluate evaluates position from White'env POV.
func Evaluate(pos *Position) int32 {
	var env Eval
	EvaluatePosition(pos, &env)
	score := env.Feed()
	if KnownLossScore >= score || score >= KnownWinScore {
		panic(fmt.Sprintf("score %d should be between %d and %d",
			score, KnownLossScore, KnownWinScore))
	}
	return score
}

func Phase(pos *Position) int32 {
	total := int32(16*0 + 4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	curr -= pos.ByFigure[Pawn].Popcnt() * 0
	curr -= pos.ByFigure[Knight].Popcnt() * 1
	curr -= pos.ByFigure[Bishop].Popcnt() * 1
	curr -= pos.ByFigure[Rook].Popcnt() * 2
	curr -= pos.ByFigure[Queen].Popcnt() * 4
	return (curr*256 + total/2) / total
}
