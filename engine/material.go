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
	Weights = [139]Score{
		{M: -125, E: -121}, {M: 14257, E: 12139}, {M: 63689, E: 50937}, {M: 63632, E: 53680}, {M: 92569, E: 95117}, {M: 210954, E: 167754}, {M: -149, E: 152}, {M: 13, E: 128},
		{M: 1030, E: 2674}, {M: 2094, E: -174}, {M: 1732, E: 418}, {M: 1175, E: 522}, {M: 414, E: 1315}, {M: -682, E: -1564}, {M: -2610, E: 2546}, {M: -2343, E: 2106},
		{M: -2842, E: 1647}, {M: -528, E: -654}, {M: -3450, E: -304}, {M: 4691, E: -1255}, {M: 3709, E: -2201}, {M: -2901, E: -2507}, {M: -2056, E: 1588}, {M: -3844, E: 1947},
		{M: -853, E: -53}, {M: -1704, E: -1334}, {M: -34, E: 155}, {M: 274, E: -137}, {M: 1491, E: -1585}, {M: -2243, E: -735}, {M: -1560, E: 3633}, {M: -3912, E: 3207},
		{M: 646, E: -405}, {M: 2327, E: -1394}, {M: 1490, E: -877}, {M: 1254, E: -808}, {M: -1446, E: 831}, {M: -3254, E: 960}, {M: -2147, E: 5900}, {M: -2498, E: 4902},
		{M: 435, E: 1294}, {M: 1767, E: -2237}, {M: 320, E: -746}, {M: 2223, E: 73}, {M: -518, E: 2251}, {M: -4213, E: 3302}, {M: 893, E: 9274}, {M: 2313, E: 6293},
		{M: 10607, E: -350}, {M: 6332, E: -4089}, {M: 13939, E: -6127}, {M: 13495, E: -554}, {M: 6604, E: 3464}, {M: 5416, E: 4996}, {M: 7201, E: 1179}, {M: 8410, E: 36},
		{M: -8, E: 1756}, {M: 544, E: -5625}, {M: -1070, E: -1004}, {M: 9493, E: -2193}, {M: -18836, E: 9954}, {M: -11447, E: 949}, {M: 171, E: -136}, {M: 1663, E: 1146},
		{M: 1722, E: 1376}, {M: 1127, E: 5742}, {M: 5464, E: 11186}, {M: 3205, E: 25081}, {M: 16435, E: 38056}, {M: -203, E: 65}, {M: 393, E: 122}, {M: -5552, E: 6427},
		{M: -4098, E: 4420}, {M: -2925, E: 164}, {M: -4931, E: -788}, {M: -8210, E: -40}, {M: -10113, E: 2516}, {M: -7295, E: 1864}, {M: -409, E: -3538}, {M: 0, E: -330},
		{M: 1345, E: 2487}, {M: 3594, E: 2994}, {M: 3660, E: 4047}, {M: 5715, E: 1166}, {M: -4070, E: -46}, {M: -23120, E: -2348}, {M: -2656, E: -2334}, {M: 2, E: -18},
		{M: -18, E: 2776}, {M: 1793, E: 3207}, {M: 1698, E: 3479}, {M: 1388, E: 2063}, {M: 1748, E: -859}, {M: -391, E: -3428}, {M: 3846, E: -1874}, {M: 5280, E: -1118},
		{M: 6575, E: 471}, {M: 3967, E: 653}, {M: 2421, E: 562}, {M: 764, E: -4}, {M: -4444, E: -9}, {M: -12593, E: -793}, {M: -568, E: -309}, {M: 1912, E: 445},
		{M: 621, E: -7}, {M: -910, E: 1431}, {M: -103, E: 1398}, {M: 208, E: 794}, {M: 4218, E: -629}, {M: 140, E: -1447}, {M: -2, E: -7711}, {M: 41, E: 553},
		{M: -1564, E: 235}, {M: -4395, E: -7}, {M: -989, E: -1}, {M: 7891, E: 211}, {M: 7100, E: -2755}, {M: 74, E: -10881}, {M: -2779, E: -6714}, {M: 6032, E: -756},
		{M: 3719, E: 1302}, {M: -8395, E: 4560}, {M: -82, E: 3596}, {M: -2629, E: 4259}, {M: 8452, E: -69}, {M: 2081, E: -5012}, {M: 177, E: 5}, {M: 2639, E: -781},
		{M: 10665, E: -3690}, {M: 14005, E: -3863}, {M: -1977, E: -1997}, {M: 2336, E: 1099}, {M: -407, E: 153}, {M: -1150, E: -1323}, {M: 15344, E: 8479}, {M: -4385, E: 2410},
		{M: 6845, E: 8933}, {M: 8604, E: -1616}, {M: 3397, E: 1070},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
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
