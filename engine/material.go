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
		{M: 4402, E: 2138}, {M: 6150, E: 5321}, {M: 50472, E: 31259}, {M: 52963, E: 35144}, {M: 71547, E: 66427},
		{M: 181692, E: 110704}, {M: 2139, E: 9130}, {M: 9001, E: 1690}, {M: 1049, E: 2824}, {M: 1288, E: 1453},
		{M: 822, E: 935}, {M: 801, E: 661}, {M: 331, E: 629}, {M: -637, E: -827}, {M: 3946, E: 8410},
		{M: 2633, E: 8431}, {M: 2686, E: 7856}, {M: 4728, E: 4804}, {M: 2632, E: 7088}, {M: 7354, E: 6912},
		{M: 5716, E: 6132}, {M: 2564, E: 4649}, {M: 4001, E: 8336}, {M: 2442, E: 7448}, {M: 3448, E: 5966},
		{M: 3693, E: 5763}, {M: 5083, E: 7105}, {M: 4964, E: 6711}, {M: 5172, E: 6065}, {M: 3319, E: 6502},
		{M: 4008, E: 9345}, {M: 2454, E: 8433}, {M: 4448, E: 6128}, {M: 5698, E: 5589}, {M: 5268, E: 6389},
		{M: 3793, E: 7004}, {M: 2451, E: 7845}, {M: 2755, E: 7173}, {M: 3633, E: 11615}, {M: 3487, E: 9996},
		{M: 4428, E: 7307}, {M: 4739, E: 5666}, {M: 4525, E: 7235}, {M: 5938, E: 5685}, {M: 3627, E: 8547},
		{M: 2355, E: 8596}, {M: 6214, E: 13402}, {M: 5118, E: 10358}, {M: 7613, E: 7582}, {M: 8915, E: 3027},
		{M: 11506, E: 495}, {M: 11892, E: 5856}, {M: 8310, E: 7754}, {M: 9703, E: 7731}, {M: 16627, E: 16125},
		{M: 21533, E: 12855}, {M: 10815, E: 12242}, {M: 16998, E: 7020}, {M: 13210, E: 7844}, {M: 13164, E: 10863},
		{M: 3434, E: 14881}, {M: 2813, E: 13106}, {M: 6051, E: 5956}, {M: -251, E: 2159}, {M: 481, E: 2610},
		{M: 321, E: 5364}, {M: 3741, E: 8413}, {M: 6305, E: 15564}, {M: 11295, E: 14100}, {M: 2734, E: 3605},
		{M: 4494, E: -2584}, {M: 2655, E: 4063}, {M: 1976, E: 4799}, {M: 1415, E: 6530}, {M: -6358, E: 9623},
		{M: -6884, E: 10768}, {M: 24222, E: 6851}, {M: 37209, E: -775}, {M: 1303, E: 448}, {M: 7577, E: 5060},
		{M: 6207, E: 7302}, {M: -1262, E: 8982}, {M: 2902, E: 8077}, {M: 124, E: 8727}, {M: 8859, E: 5480},
		{M: 6076, E: 295}, {M: 1377, E: 260}, {M: 62, E: -836}, {M: -997, E: -1211}, {M: 1771, E: 894},
		{M: -2856, E: 707}, {M: 4201, E: 6793}, {M: 5280, E: -196}, {M: 1473, E: 1080},
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
