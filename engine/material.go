// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  = 25000          // KnownWinScore is strictly greater than all evaluation scores (mate not included).
	KnownLossScore = -KnownWinScore // KnownLossScore is strictly smaller than all evaluation scores (mated not included).
	MateScore      = 30000          // MateScore - N is mate in N plies.
	MatedScore     = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  = 32000          // InfinityScore is possible score. -InfinityScore is the minimum possible score.
)

var (
	// Weights stores all evaluation parameters under one array for easy handling.
	// All numbers are multiplied by 128.
	//
	// Zurichess' evaluation is a very simple neural network with no hidden layers,
	// and one output node y = W_m * x * (1-p) + W_e * x * p where W_m are
	// middle game weights, W_e are endgame weights, x is input, p is phase between
	// middle game and end game, and y is the score.
	// The network has |x| = len(Weights) inputs corresponding to features
	// extracted from the position. These features are symmetrical wrt to colors.
	// The network is trained using the Texel's Tuning Method
	// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.
	Weights = [98]Score{
		{M: -48, E: -240}, {M: 10807, E: 10843}, {M: 43288, E: 27245}, {M: 45684, E: 31056}, {M: 63731, E: 60993}, {M: 181992, E: 83297}, {M: -56, E: 12}, {M: -78, E: -25},
		{M: 929, E: 2804}, {M: 1285, E: 1388}, {M: 906, E: 925}, {M: 874, E: 479}, {M: 281, E: 449}, {M: -159, E: -867}, {M: -288, E: 235}, {M: -1764, E: 812},
		{M: -1967, E: 599}, {M: -255, E: -1939}, {M: -1134, E: -32}, {M: 3166, E: 26}, {M: 2062, E: -995}, {M: -1359, E: -2497}, {M: -101, E: 57}, {M: -1676, E: -12},
		{M: -1405, E: -1000}, {M: -1299, E: -506}, {M: 167, E: -42}, {M: 287, E: 47}, {M: 1112, E: -1426}, {M: -918, E: -1028}, {M: -123, E: 957}, {M: -1316, E: 956},
		{M: 397, E: -1076}, {M: 1340, E: -1923}, {M: 748, E: -390}, {M: -753, E: -545}, {M: -1208, E: -100}, {M: -1552, E: -152}, {M: -84, E: 3537}, {M: 33, E: 1946},
		{M: -189, E: 98}, {M: 847, E: -961}, {M: 697, E: 81}, {M: 433, E: -245}, {M: 8, E: 625}, {M: -1397, E: 693}, {M: 844, E: 5027}, {M: 831, E: 3292},
		{M: 478, E: 2040}, {M: 1348, E: -832}, {M: -63, E: -1174}, {M: 6036, E: -119}, {M: 1861, E: 899}, {M: 2830, E: 83}, {M: 468, E: 2380}, {M: 348, E: 2179},
		{M: 45, E: 248}, {M: 374, E: -416}, {M: -95, E: -733}, {M: 138, E: -440}, {M: 154, E: 181}, {M: -423, E: 150}, {M: 10, E: -128}, {M: 47, E: 1691},
		{M: -949, E: 2327}, {M: 60, E: 4967}, {M: 4223, E: 7314}, {M: 11073, E: 12313}, {M: 23024, E: 16587}, {M: -93, E: 61}, {M: 17, E: -7178}, {M: -1017, E: -1331},
		{M: -1471, E: -181}, {M: -27, E: 214}, {M: 81, E: 992}, {M: 204, E: 2267}, {M: 637, E: 1150}, {M: -337, E: -5214}, {M: -203, E: -3956}, {M: 3070, E: -177},
		{M: 3313, E: 558}, {M: -2856, E: 1692}, {M: -372, E: 1661}, {M: -1920, E: 2101}, {M: 5799, E: -749}, {M: 3161, E: -5010}, {M: -127, E: 11}, {M: 1731, E: -281},
		{M: 6030, E: -2034}, {M: 5750, E: -193}, {M: 1537, E: 2}, {M: -534, E: 68}, {M: -670, E: -1376}, {M: 2317, E: 212}, {M: -2575, E: 1406}, {M: 5165, E: 6489},
		{M: 4570, E: -15}, {M: 1507, E: 684},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
	wPassedPawn         [8]Score
	wKingRank           [8]Score
	wKingFile           [8]Score
	wKingAttack         [4]Score
	wConnectedPawn      Score
	wDoublePawn         Score
	wIsolatedPawn       Score
	wPawnThreat         Score
	wKingShelter        Score
	wBishopPair         Score
	wRookOnOpenFile     Score
	wRookOnHalfOpenFile Score

	// Evaluation caches.
	pawnsAndShelterCache *cache

	// Figure bonuses to use when computing the futility margin.
	futilityFigureBonus [FigureArraySize]int32
)

func init() {
	// Initializes weights.
	initWeights()
	slice := func(w []Score, out []Score) []Score {
		copy(out, w)
		return w[len(out):]
	}
	entry := func(w []Score, out *Score) []Score {
		*out = w[0]
		return w[1:]
	}

	w := Weights[:]
	w = slice(w, wFigure[:])
	w = slice(w, wMobility[:])
	w = slice(w, wPawn[:])
	w = slice(w, wPassedPawn[:])
	w = slice(w, wKingRank[:])
	w = slice(w, wKingFile[:])
	w = slice(w, wKingAttack[:])
	w = entry(w, &wConnectedPawn)
	w = entry(w, &wDoublePawn)
	w = entry(w, &wIsolatedPawn)
	w = entry(w, &wPawnThreat)
	w = entry(w, &wKingShelter)
	w = entry(w, &wBishopPair)
	w = entry(w, &wRookOnOpenFile)
	w = entry(w, &wRookOnHalfOpenFile)

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}

	// Initialize caches.
	pawnsAndShelterCache = newCache(9, hashPawnsAndShelter, evaluatePawnsAndShelter)

	// Initializes futility figure bonus
	for i, w := range wFigure {
		futilityFigureBonus[i] = scaleToCentipawn(max(w.M, w.E))
	}
}

func hashPawnsAndShelter(pos *Position, us Color) uint64 {
	h := murmurSeed[us]
	h = murmurMix(h, uint64(pos.ByPiece(us, Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us.Opposite(), Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us, King)))
	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		// Mixes in something to signal queen's presence.
		h = murmurMix(h, murmurSeed[NoColor])
	}
	return h
}

func evaluatePawnsAndShelter(pos *Position, us Color) Eval {
	var eval Eval
	eval.merge(evaluatePawns(pos, us))
	eval.merge(evaluateShelter(pos, us))
	return eval
}

func evaluatePawns(pos *Position, us Color) Eval {
	var eval Eval
	ours := pos.ByPiece(us, Pawn)

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
	wings := East(ours) | West(ours)
	double := Bitboard(0)
	if us == White {
		double = ours & South(ours)
	} else /* if us == Black */ {
		double = ours & North(ours)
	}

	isolated := ours &^ Fill(wings)                           // no pawn on the adjacent files
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	passed := passedPawns(pos, us)                            // no pawn in front and no enemy on the adjacent files

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.add(wFigure[Pawn])
		eval.add(wPawn[povSq-8])

		if passed.Has(sq) {
			eval.add(wPassedPawn[rank])
		}
		if connected.Has(sq) {
			eval.add(wConnectedPawn)
		}
		if double.Has(sq) {
			eval.add(wDoublePawn)
		}
		if isolated.Has(sq) {
			eval.add(wIsolatedPawn)
		}
	}

	return eval
}

func evaluateShelter(pos *Position, us Color) Eval {
	var eval Eval
	pawns := pos.ByPiece(us, Pawn)
	king := pos.ByPiece(us, King)

	sq := king.AsSquare().POV(us)
	eval.add(wKingFile[sq.File()])
	eval.add(wKingRank[sq.Rank()])

	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		king = ForwardSpan(us, king)
		file := sq.File()
		if file > 0 && West(king)&pawns == 0 {
			eval.add(wKingShelter)
		}
		if king&pawns == 0 {
			eval.addN(wKingShelter, 2)
		}
		if file < 7 && East(king)&pawns == 0 {
			eval.add(wKingShelter)
		}
	}
	return eval
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	eval.merge(pawnsAndShelterCache.load(pos, us))
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	theirPawns := pos.ByPiece(them, Pawn)
	theirKing := pos.ByPiece(them, King)
	theirKingArea := bbKingArea[theirKing.AsSquare()]
	numAttackers := 0
	attackStrength := int32(0)

	// Pawn forward mobility.
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	eval.addN(wMobility[Pawn], mobility.Count())
	mobility = pos.PawnThreats(us)
	eval.addN(wPawnThreat, (mobility & pos.ByColor[them]).Count())

	// Knight
	excl := pos.ByPiece(us, Pawn) | pos.PawnThreats(them)
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		eval.add(wFigure[Knight])
		mobility := KnightMobility(sq)
		eval.addN(wMobility[Knight], (mobility &^ excl).Count())

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.add(wFigure[Bishop])
		mobility := BishopMobility(sq, all)
		eval.addN(wMobility[Bishop], (mobility &^ excl).Count())
		numBishops++

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	eval.addN(wBishopPair, numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.add(wFigure[Rook])
		mobility := RookMobility(sq, all)
		eval.addN(wMobility[Rook], (mobility &^ excl).Count())

		// Evaluate rook on open and semi open files.
		// https://chessprogramming.wikispaces.com/Rook+on+Open+File
		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				eval.add(wRookOnOpenFile)
			} else {
				eval.add(wRookOnHalfOpenFile)
			}
		}

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.add(wFigure[Queen])
		mobility := QueenMobility(sq, all) &^ excl
		eval.addN(wMobility[Queen], (mobility &^ excl).Count())

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}

	// King, each side has one.
	{
		sq := pos.ByPiece(us, King).AsSquare()
		mobility := KingMobility(sq) &^ excl
		eval.addN(wMobility[King], mobility.Count())
	}

	// Evaluate attacking the king. See more at:
	// https://chessprogramming.wikispaces.com/King+Safety#Attacking%20King%20Zone
	if numAttackers >= len(wKingAttack) {
		numAttackers = len(wKingAttack) - 1
	}
	eval.addN(wKingAttack[numAttackers], attackStrength)
}

// EvaluatePosition evalues position exported to be used by the tuner.
func EvaluatePosition(pos *Position) Eval {
	// TODO: export from to score_coach.go.
	var eval Eval
	evaluateSide(pos, Black, &eval)
	eval.neg()
	evaluateSide(pos, White, &eval)
	return eval
}

// Evaluate evaluates position from White's POV.
// Scores fits into a int16.
func Evaluate(pos *Position) int32 {
	eval := EvaluatePosition(pos)
	score := eval.Feed(Phase(pos))
	return scaleToCentipawn(score)
}

// Phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func Phase(pos *Position) int32 {
	total := int32(4*1 + 4*1 + 4*3 + 2*6)
	curr := total
	curr -= pos.ByFigure[Knight].Count() * 1
	curr -= pos.ByFigure[Bishop].Count() * 1
	curr -= pos.ByFigure[Rook].Count() * 3
	curr -= pos.ByFigure[Queen].Count() * 6
	return (curr*256 + total/2) / total
}

// passedPawns returns all passed pawns of us in pos.
func passedPawns(pos *Position, us Color) Bitboard {
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)
	theirs |= East(theirs) | West(theirs)
	block := BackwardSpan(us, theirs|ours)
	return ours &^ block
}

// scaleToCentipawn scales a score in the original scale to centipawns.
func scaleToCentipawn(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + 64 + score>>31) >> 7
}
