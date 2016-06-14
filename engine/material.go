// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	Weights = [122]Score{
		{M: 476, E: 149}, {M: 14711, E: 11943}, {M: 58914, E: 49809}, {M: 68484, E: 51664}, {M: 91684, E: 95017}, {M: 204392, E: 170320}, {M: 91, E: 344}, {M: 27, E: -125},
		{M: 837, E: 2544}, {M: 2136, E: -81}, {M: 1775, E: 678}, {M: 1262, E: 514}, {M: 457, E: 1307}, {M: -455, E: -1515}, {M: -2902, E: 2856}, {M: -2793, E: 2492},
		{M: -2910, E: 1837}, {M: -300, E: -577}, {M: -3010, E: 96}, {M: 4866, E: -993}, {M: 3295, E: -1648}, {M: -3133, E: -2301}, {M: -2532, E: 1678}, {M: -4099, E: 2068},
		{M: -1401, E: -24}, {M: -2279, E: -1275}, {M: -318, E: 302}, {M: 136, E: -146}, {M: 1562, E: -1695}, {M: -2755, E: -673}, {M: -2243, E: 3782}, {M: -4480, E: 3419},
		{M: 390, E: -470}, {M: 1881, E: -1301}, {M: 1215, E: -1000}, {M: 879, E: -1010}, {M: -1436, E: 599}, {M: -3976, E: 942}, {M: -2452, E: 6202}, {M: -2809, E: 5209},
		{M: 507, E: 1422}, {M: 1818, E: -2203}, {M: 425, E: -554}, {M: 2566, E: -67}, {M: -26, E: 2044}, {M: -4314, E: 3512}, {M: 686, E: 9215}, {M: 1171, E: 6524},
		{M: 9993, E: -591}, {M: 5979, E: -4556}, {M: 12992, E: -6087}, {M: 13809, E: -1200}, {M: 7259, E: 2700}, {M: 5244, E: 4750}, {M: 5510, E: 1235}, {M: 6498, E: -23},
		{M: -50, E: 502}, {M: 305, E: -7033}, {M: -704, E: -2259}, {M: 7911, E: -2820}, {M: -20638, E: 9709}, {M: -13457, E: 818}, {M: -119, E: 105}, {M: 1505, E: 1316},
		{M: 1547, E: 1695}, {M: 1165, E: 5880}, {M: 5429, E: 10873}, {M: 3765, E: 24841}, {M: 17891, E: 38055}, {M: 70, E: 86}, {M: -34, E: -158}, {M: -2333, E: 4767},
		{M: -2815, E: 3547}, {M: -2352, E: -187}, {M: -5126, E: -805}, {M: -7963, E: -85}, {M: -8281, E: 1916}, {M: -3295, E: 371}, {M: 4313, E: -3430}, {M: 4440, E: -207},
		{M: 5620, E: 2424}, {M: 8102, E: 3129}, {M: 8275, E: 4239}, {M: 10596, E: 1230}, {M: 966, E: 87}, {M: -16450, E: -2347}, {M: -2694, E: -1539}, {M: 7, E: 440},
		{M: 6, E: 3154}, {M: 2078, E: 3595}, {M: 1860, E: 3874}, {M: 1330, E: 2429}, {M: 1901, E: -146}, {M: -302, E: -2757}, {M: 900, E: -8913}, {M: 3, E: -39},
		{M: -3050, E: 839}, {M: -6737, E: 1383}, {M: -4860, E: 1926}, {M: 4104, E: 2041}, {M: 2552, E: -649}, {M: -142, E: -9488}, {M: -334, E: -7663}, {M: 6626, E: -420},
		{M: 3889, E: 2078}, {M: -9249, E: 5634}, {M: -154, E: 4480}, {M: -1984, E: 4734}, {M: 9355, E: -63}, {M: 4398, E: -5995}, {M: 78, E: -294}, {M: 2651, E: -878},
		{M: 10434, E: -3677}, {M: 13824, E: -3768}, {M: 2350, E: 836}, {M: -951, E: 185}, {M: -1012, E: -1160}, {M: 15364, E: 8646}, {M: -4662, E: 2182}, {M: 6817, E: 8907},
		{M: 8592, E: -1687}, {M: 3249, E: 1069},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
	wPassedPawn         [8]Score
	wPassedPawnKing     [8]Score
	wKnightRank         [8]Score
	wKnightFile         [8]Score
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
	w = slice(w, wPassedPawnKing[:])
	w = slice(w, wKnightRank[:])
	w = slice(w, wKnightFile[:])
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
	wings := East(ours) | West(ours)
	double := ours & Backward(us, ours)
	isolated := ours &^ Fill(wings)                           // no pawn on the adjacent files
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	passed := passedPawns(pos, us)                            // no pawn in front and no enemy on the adjacent files

	kingPawnDist := 8
	kingSq := pos.ByPiece(us, King).AsSquare()

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.add(wFigure[Pawn])
		eval.add(wPawn[povSq-8])

		if passed.Has(sq) {
			eval.add(wPassedPawn[rank])
			if kingPawnDist > distance[sq][kingSq] {
				kingPawnDist = distance[sq][kingSq]
			}
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

	if kingPawnDist != 8 && pos.ByPiece(us.Opposite(), Queen) == 0 {
		// Add a bonus for king protecting most advance pawn.
		eval.add(wPassedPawnKing[kingPawnDist])
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

		sq = sq.POV(us)
		eval.add(wKnightFile[sq.File()])
		eval.add(wKnightRank[sq.Rank()])

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
	var eval Eval
	evaluateSide(pos, Black, &eval)
	eval.neg()
	evaluateSide(pos, White, &eval)
	return eval
}

// Evaluate evaluates position from White's POV.
// The returned s fits into a int16.
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
// From white's POV: w - white pawn, b - black pawn, x - non-passed pawns.
// ........
// ........
// .....w..
// .....x..
// ..b..x..
// .xxx.x..
// .xxx.x..
// .xxx.x..
// .xxx.x..
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
