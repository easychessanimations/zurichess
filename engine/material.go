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
		{M: 124, E: -77}, {M: 14430, E: 12323}, {M: 63306, E: 50327}, {M: 68276, E: 51765}, {M: 91323, E: 95148}, {M: 207961, E: 169472}, {M: 161, E: 499}, {M: -115, E: 107},
		{M: 796, E: 2682}, {M: 2119, E: -87}, {M: 1778, E: 677}, {M: 1273, E: 506}, {M: 457, E: 1297}, {M: -423, E: -1618}, {M: -2653, E: 2443}, {M: -2534, E: 2061},
		{M: -2765, E: 1648}, {M: -63, E: -976}, {M: -2783, E: -344}, {M: 5033, E: -1246}, {M: 3507, E: -1980}, {M: -2863, E: -2657}, {M: -2338, E: 1401}, {M: -3877, E: 1676},
		{M: -1264, E: -186}, {M: -2060, E: -1546}, {M: -129, E: 38}, {M: 237, E: -236}, {M: 1732, E: -1777}, {M: -2520, E: -887}, {M: -2058, E: 3500}, {M: -4285, E: 3141},
		{M: 434, E: -393}, {M: 2021, E: -1427}, {M: 1284, E: -981}, {M: 945, E: -832}, {M: -1409, E: 816}, {M: -3815, E: 870}, {M: -2352, E: 5991}, {M: -2681, E: 5031},
		{M: 532, E: 1514}, {M: 1934, E: -2183}, {M: 480, E: -445}, {M: 2487, E: 261}, {M: -85, E: 2342}, {M: -4232, E: 3514}, {M: 822, E: 8956}, {M: 1249, E: 6378},
		{M: 10396, E: -528}, {M: 5823, E: -4251}, {M: 12741, E: -5822}, {M: 13434, E: -768}, {M: 6530, E: 3213}, {M: 5348, E: 4741}, {M: 6871, E: 857}, {M: 7191, E: -5},
		{M: -37, E: 1247}, {M: 159, E: -5799}, {M: -1161, E: -1344}, {M: 8345, E: -2267}, {M: -19726, E: 9825}, {M: -12253, E: 1007}, {M: -362, E: -88}, {M: 1632, E: 1153},
		{M: 1560, E: 1631}, {M: 1094, E: 5876}, {M: 5268, E: 11025}, {M: 3531, E: 25156}, {M: 17182, E: 38087}, {M: -218, E: 628}, {M: 279, E: 76}, {M: -5351, E: 6039},
		{M: -3667, E: 4037}, {M: -2743, E: -122}, {M: -5586, E: -953}, {M: -8772, E: -63}, {M: -9417, E: 2231}, {M: -5986, E: 1250}, {M: -59, E: -3615}, {M: -9, E: -306},
		{M: 1125, E: 2445}, {M: 3640, E: 3047}, {M: 3894, E: 4106}, {M: 6200, E: 1088}, {M: -3149, E: 42}, {M: -20656, E: -2434}, {M: -2752, E: -2003}, {M: -35, E: -5},
		{M: -8, E: 2734}, {M: 2040, E: 3179}, {M: 1810, E: 3417}, {M: 1343, E: 2015}, {M: 1752, E: -460}, {M: -357, E: -3142}, {M: 590, E: -7979}, {M: 2, E: 470},
		{M: -1524, E: 173}, {M: -4098, E: -80}, {M: -1224, E: 32}, {M: 8164, E: 60}, {M: 5978, E: -2531}, {M: -185, E: -10996}, {M: -1672, E: -6795}, {M: 6291, E: -636},
		{M: 3720, E: 1572}, {M: -8631, E: 4785}, {M: -338, E: 3899}, {M: -2220, E: 4302}, {M: 8899, E: -70}, {M: 3454, E: -5243}, {M: 64, E: -85}, {M: 2574, E: -771},
		{M: 10309, E: -3582}, {M: 13537, E: -3750}, {M: 2305, E: 897}, {M: -943, E: 158}, {M: -990, E: -1224}, {M: 15328, E: 8706}, {M: -4391, E: 2371}, {M: 6744, E: 9004},
		{M: 8512, E: -1644}, {M: 3156, E: 1244},
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
	them := us.Opposite()
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(them, Pawn)

	wings := East(ours) | West(ours)
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	double := DoubledPawns(us, ours)
	isolated := IsolatedPawns(ours)
	passed := PassedPawns(us, ours, theirs)

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
