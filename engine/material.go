// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  int32 = 25000000       // KnownWinScore is strictly greater than all evaluation scores (mate not included).
	KnownLossScore int32 = -KnownWinScore // KnownLossScore is strictly smaller than all evaluation scores (mated not included).
	MateScore      int32 = 30000000       // MateScore - N is mate in N plies.
	MatedScore     int32 = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  int32 = 32000000       // InfinityScore is possible score. -InfinityScore is the minimum possible score.
)

var (
	// Weights stores all evaluation parameters under one array for easy handling.
	Weights = [94]Score{
		{M: 3481, E: 5706}, {M: 4520, E: 7387}, {M: 45831, E: 31782}, {M: 49213, E: 35926}, {M: 67581, E: 66240},
		{M: 159356, E: 116233}, {M: 2206, E: 8019}, {M: 7814, E: 7480}, {M: 775, E: 2896}, {M: 1189, E: 1449},
		{M: 741, E: 944}, {M: 635, E: 687}, {M: 340, E: 578}, {M: -907, E: -707}, {M: 3645, E: 6747},
		{M: 2993, E: 6233}, {M: 3720, E: 5385}, {M: 5707, E: 2879}, {M: 4189, E: 4525}, {M: 7354, E: 5123},
		{M: 5491, E: 4670}, {M: 2656, E: 3874}, {M: 3768, E: 6669}, {M: 2851, E: 5690}, {M: 4149, E: 4637},
		{M: 4841, E: 3341}, {M: 5621, E: 4238}, {M: 5542, E: 4852}, {M: 5205, E: 4297}, {M: 2946, E: 5201},
		{M: 3998, E: 8201}, {M: 2787, E: 6364}, {M: 5141, E: 4047}, {M: 6333, E: 3795}, {M: 6308, E: 3475},
		{M: 4197, E: 4700}, {M: 3354, E: 5547}, {M: 2568, E: 6030}, {M: 3380, E: 9937}, {M: 3600, E: 8452},
		{M: 4641, E: 5304}, {M: 5387, E: 4064}, {M: 4266, E: 4735}, {M: 5189, E: 4351}, {M: 2280, E: 7394},
		{M: 2043, E: 7685}, {M: 6277, E: 10855}, {M: 6389, E: 8912}, {M: 6557, E: 6298}, {M: 8077, E: 1469},
		{M: 7070, E: 324}, {M: 9906, E: 4133}, {M: 3625, E: 7927}, {M: 6976, E: 8760}, {M: 16094, E: 13825},
		{M: 20142, E: 9776}, {M: 9315, E: 13083}, {M: 12448, E: 7361}, {M: 4236, E: 8096}, {M: 8727, E: 10441},
		{M: 2948, E: 13426}, {M: 3997, E: 13618}, {M: 5212, E: 1237}, {M: -119, E: 1996}, {M: 254, E: 2253},
		{M: -156, E: 5744}, {M: 2735, E: 8861}, {M: 5822, E: 15670}, {M: 9062, E: 14143}, {M: 9744, E: 1520},
		{M: 6048, E: 950}, {M: 4576, E: 7138}, {M: 3231, E: 8426}, {M: 1396, E: 10090}, {M: -3867, E: 13085},
		{M: -7652, E: 14337}, {M: 3290, E: 10614}, {M: 17417, E: 4356}, {M: 5069, E: 1224}, {M: 9734, E: 5819},
		{M: 8447, E: 7117}, {M: 1444, E: 9043}, {M: 5685, E: 8395}, {M: 3599, E: 8867}, {M: 10955, E: 5823},
		{M: 7365, E: 1618}, {M: 1165, E: 308}, {M: -718, E: 153}, {M: -696, E: -1279}, {M: 1784, E: 1163},
		{M: -2835, E: 989}, {M: 2927, E: 8446}, {M: 4997, E: -187}, {M: 1433, E: 938},
	}

	// Named chunks of Weights
	wFigure     [FigureArraySize]Score
	wMobility   [FigureArraySize]Score
	wPawn       [48]Score
	wPassedPawn [8]Score
	wKingRank   [8]Score
	wKingFile   [8]Score
	wFlags      [8]Score // see flags defined below
)

const (
	fConnectedPawn = iota
	fDoublePawn
	fIsolatedPawn
	fPawnThreat
	fKingShelter
	fBishopPair
	fRookOnOpenFile
	fRookOnHalfOpenFile
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

	fixPSqT(wPawn[:], &wFigure[Pawn])

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}
}

// fixPSqT fixes psqt to have the average 0.
// adjusts figure with the corresponding amount.
func fixPSqT(psqt []Score, figure *Score) {
	var avg Score
	for i := range psqt {
		avg.M += psqt[i].M
		avg.E += psqt[i].E
	}
	n := int32(len(psqt))
	avg.M = (avg.M + n/2) / n
	avg.E = (avg.E + n/2) / n

	figure.M += avg.M
	figure.E += avg.E
	for i := range psqt {
		psqt[i].M -= avg.M
		psqt[i].E -= avg.E
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
		mobility := KnightMobility(sq) &^ excl
		eval.AddN(wMobility[Knight], mobility.Popcnt())
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Bishop])
		mobility := BishopMobility(sq, all) &^ excl
		eval.AddN(wMobility[Bishop], mobility.Popcnt())
		numBishops++
	}
	eval.AddN(wFlags[fBishopPair], numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Rook])
		mobility := RookMobility(sq, all) &^ excl
		eval.AddN(wMobility[Rook], mobility.Popcnt())

		// Evaluate rook on open and semi open files.
		// https://chessprogramming.wikispaces.com/Rook+on+Open+File
		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				eval.Add(wFlags[fRookOnOpenFile])
			} else {
				eval.Add(wFlags[fRookOnHalfOpenFile])
			}
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Queen])
		mobility := QueenMobility(sq, all) &^ excl
		eval.AddN(wMobility[Queen], mobility.Popcnt())
	}

	// King
	for bb := pos.ByPiece(us, King); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[King])
		mobility := KingMobility(sq) &^ excl
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

// Evaluate evaluates position from White's POV.
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

// ScaleToCentiPawn scales the score returned by Evaluate
// such that one pawn ~= 100.
func ScaleToCentiPawn(score int32) int32 {
	return (score + 64) / 128
}

// phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func phase(pos *Position) int32 {
	total := int32(16*0 + 4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	curr -= pos.ByFigure[Pawn].Popcnt() * 0
	curr -= pos.ByFigure[Knight].Popcnt() * 1
	curr -= pos.ByFigure[Bishop].Popcnt() * 1
	curr -= pos.ByFigure[Rook].Popcnt() * 2
	curr -= pos.ByFigure[Queen].Popcnt() * 4
	return (curr*256 + total/2) / total
}
