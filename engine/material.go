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
	Weights = [155]Score{
		{M: -202, E: -14}, {M: 14350, E: 12102}, {M: 63876, E: 50891}, {M: 66417, E: 53568}, {M: 92905, E: 95034}, {M: 213194, E: 166348}, {M: 203, E: -39}, {M: -100, E: 49},
		{M: 1032, E: 2676}, {M: 2100, E: -183}, {M: 1728, E: 423}, {M: 1176, E: 520}, {M: 408, E: 1326}, {M: -699, E: -1563}, {M: -154, E: -76}, {M: -407, E: 50},
		{M: -84, E: -56}, {M: 256, E: 88}, {M: -61, E: 41}, {M: 93, E: -4}, {M: 0, E: 138}, {M: 105, E: 16}, {M: -2680, E: 2556}, {M: -2423, E: 2140},
		{M: -2908, E: 1668}, {M: -612, E: -616}, {M: -3534, E: -282}, {M: 4631, E: -1245}, {M: 3665, E: -2203}, {M: -2982, E: -2481}, {M: -2124, E: 1592}, {M: -3910, E: 1958},
		{M: -909, E: -51}, {M: -1769, E: -1318}, {M: -104, E: 158}, {M: 222, E: -129}, {M: 1457, E: -1587}, {M: -2310, E: -724}, {M: -1617, E: 3637}, {M: -3977, E: 3226},
		{M: 577, E: -402}, {M: 2295, E: -1404}, {M: 1430, E: -855}, {M: 1203, E: -795}, {M: -1486, E: 842}, {M: -3307, E: 968}, {M: -2200, E: 5904}, {M: -2540, E: 4915},
		{M: 420, E: 1283}, {M: 1707, E: -2235}, {M: 257, E: -719}, {M: 2096, E: 98}, {M: -592, E: 2265}, {M: -4270, E: 3308}, {M: 793, E: 9295}, {M: 2278, E: 6294},
		{M: 10580, E: -374}, {M: 6413, E: -4135}, {M: 13889, E: -6161}, {M: 13405, E: -538}, {M: 6653, E: 3408}, {M: 5430, E: 4967}, {M: 7624, E: 941}, {M: 7848, E: 12},
		{M: -36, E: 1613}, {M: 2377, E: -6336}, {M: -1269, E: -1163}, {M: 10267, E: -2608}, {M: -18347, E: 9763}, {M: -11847, E: 941}, {M: -76, E: 201}, {M: 71, E: 15},
		{M: 31, E: 55}, {M: -158, E: -122}, {M: -153, E: -7}, {M: -44, E: -63}, {M: -35, E: 14}, {M: -127, E: 7}, {M: -60, E: 13}, {M: 1683, E: 1144},
		{M: 1752, E: 1380}, {M: 1157, E: 5733}, {M: 5501, E: 11174}, {M: 3171, E: 25097}, {M: 16335, E: 38214}, {M: -24, E: 68}, {M: -113, E: 109}, {M: -5689, E: 6464},
		{M: -4257, E: 4468}, {M: -3040, E: 222}, {M: -5106, E: -729}, {M: -8371, E: 4}, {M: -10294, E: 2595}, {M: -8017, E: 2094}, {M: -405, E: -3521}, {M: 5, E: -308},
		{M: 1339, E: 2520}, {M: 3599, E: 3014}, {M: 3673, E: 4066}, {M: 5658, E: 1198}, {M: -3976, E: 38}, {M: -23273, E: -2314}, {M: -2648, E: -2354}, {M: -1, E: 19},
		{M: -24, E: 2772}, {M: 1785, E: 3208}, {M: 1678, E: 3483}, {M: 1391, E: 2052}, {M: 1741, E: -888}, {M: -405, E: -3426}, {M: 1408, E: -1841}, {M: 2877, E: -1106},
		{M: 4176, E: 477}, {M: 1564, E: 652}, {M: -7, E: 562}, {M: -1547, E: 12}, {M: -6911, E: 2}, {M: -15118, E: -757}, {M: -683, E: -325}, {M: 1776, E: 463},
		{M: 519, E: 11}, {M: -1033, E: 1443}, {M: -223, E: 1397}, {M: 105, E: 798}, {M: 4112, E: -650}, {M: -19, E: -1460}, {M: -4, E: -7714}, {M: 65, E: 554},
		{M: -1553, E: 270}, {M: -4415, E: -8}, {M: -1063, E: -11}, {M: 8202, E: 159}, {M: 7691, E: -2856}, {M: 268, E: -10916}, {M: -2856, E: -6699}, {M: 6014, E: -751},
		{M: 3707, E: 1314}, {M: -8448, E: 4576}, {M: -120, E: 3605}, {M: -2695, E: 4270}, {M: 8420, E: -50}, {M: 2036, E: -5010}, {M: 92, E: -24}, {M: 2657, E: -783},
		{M: 10689, E: -3702}, {M: 14036, E: -3895}, {M: -1972, E: -2011}, {M: 2343, E: 1103}, {M: -379, E: 147}, {M: -1156, E: -1323}, {M: 15348, E: 8548}, {M: -4384, E: 2404},
		{M: 6828, E: 8954}, {M: 8626, E: -1632}, {M: 3386, E: 1076},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [SquareArraySize]Score
	wPassedPawn         [8]Score
	wPassedPawnKing     [8]Score
	wKnightRank         [8]Score
	wKnightFile         [8]Score
	wBishopRank         [8]Score
	wBishopFile         [8]Score
	wKingRank           [8]Score
	wKingFile           [8]Score
	wKingAttack         [4]Score
	wBackwardPawn       Score
	wConnectedPawn      Score
	wDoublePawn         Score
	wIsolatedPawn       Score
	wPawnThreat         Score
	wKingShelter        Score
	wBishopPair         Score
	wRookOnOpenFile     Score
	wRookOnHalfOpenFile Score

	// Evaluation caches.
	pawnsAndShelterCache *pawnsTable

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
	w = slice(w, wBishopRank[:])
	w = slice(w, wBishopFile[:])
	w = slice(w, wKingRank[:])
	w = slice(w, wKingFile[:])
	w = slice(w, wKingAttack[:])
	w = entry(w, &wBackwardPawn)
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
	pawnsAndShelterCache = new(pawnsTable)

	// Initializes futility figure bonus
	for i, w := range wFigure {
		futilityFigureBonus[i] = scaleToCentipawn(max(w.M, w.E))
	}
}

func evaluatePawnsAndShelter(pos *Position, us Color) Eval {
	var eval Eval
	eval.merge(evaluatePawns(pos, us))
	eval.merge(evaluateShelter(pos, us))
	return eval
}

func evaluatePawns(pos *Position, us Color) Eval {
	them := us.Opposite()

	var eval Eval
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(them, Pawn)

	wings := East(ours) | West(ours)
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	double := DoubledPawns(us, ours)
	isolated := IsolatedPawns(ours)
	passed := PassedPawns(us, ours, theirs)
	backward := BackwardPawns(us, ours, theirs)

	kingPawnDist := 8
	kingSq := pos.ByPiece(us, King).AsSquare()

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.add(wFigure[Pawn])
		eval.add(wPawn[povSq])

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
		if backward.Has(sq) {
			eval.add(wBackwardPawn)
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

		sq = sq.POV(us)
		eval.add(wBishopFile[sq.File()])
		eval.add(wBishopRank[sq.Rank()])

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

// scaleToCentipawn scales a score in the original scale to centipawns.
func scaleToCentipawn(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + 64 + score>>31) >> 7
}
