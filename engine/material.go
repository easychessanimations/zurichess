// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	// KnownWinScore is strictly greater than all evaluation scores (mate not included).
	KnownWinScore = 25000
	// KnownLossScore is strictly smaller than all evaluation scores (mated not included).
	KnownLossScore = -KnownWinScore
	// MateScore - N is mate in N plies.
	MateScore = 30000
	// MatedScore + N is mated in N plies.
	MatedScore = -MateScore
	// InfinityScore is possible score. -InfinityScore is the minimum possible score.
	InfinityScore = 32000
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
	// extracted from the position. These features are symmetrical wrt colors.
	// The network is trained using the Texel's Tuning Method
	// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.
	Weights = [155]Score{
		{M: 263, E: -33}, {M: 14237, E: 12106}, {M: 63685, E: 50962}, {M: 63916, E: 53470}, {M: 92721, E: 94906}, {M: 209286, E: 167904}, {M: -10, E: 115}, {M: 12, E: 38},
		{M: 1017, E: 2697}, {M: 2119, E: -217}, {M: 1724, E: 400}, {M: 1186, E: 515}, {M: 417, E: 1324}, {M: -678, E: -1570}, {M: -313, E: -221}, {M: -20, E: 267},
		{M: 111, E: -146}, {M: -6, E: -2}, {M: -76, E: 46}, {M: -66, E: 216}, {M: 83, E: -6}, {M: 84, E: -68}, {M: -3044, E: 2709}, {M: -2468, E: 2134},
		{M: -2719, E: 1669}, {M: -352, E: -658}, {M: -3354, E: -309}, {M: 4612, E: -1234}, {M: 3569, E: -2208}, {M: -2937, E: -2514}, {M: -2523, E: 1774}, {M: -4021, E: 2008},
		{M: -705, E: -42}, {M: -1552, E: -1321}, {M: -27, E: 181}, {M: 198, E: -122}, {M: 1312, E: -1560}, {M: -2272, E: -701}, {M: -2094, E: 3822}, {M: -4074, E: 3279},
		{M: 730, E: -330}, {M: 2469, E: -1339}, {M: 1516, E: -836}, {M: 1159, E: -762}, {M: -1573, E: 830}, {M: -3343, E: 1000}, {M: -2764, E: 6101}, {M: -2621, E: 4980},
		{M: 608, E: 1301}, {M: 1910, E: -2252}, {M: 260, E: -688}, {M: 2036, E: 124}, {M: -656, E: 2280}, {M: -4213, E: 3352}, {M: 268, E: 9527}, {M: 1829, E: 6525},
		{M: 10492, E: -272}, {M: 6718, E: -4275}, {M: 13005, E: -5837}, {M: 13502, E: -469}, {M: 6790, E: 3388}, {M: 5313, E: 5107}, {M: 7502, E: 1072}, {M: 9380, E: 36},
		{M: -75, E: 1916}, {M: 451, E: -5258}, {M: -259, E: -994}, {M: 11849, E: -3044}, {M: -17385, E: 9972}, {M: -11490, E: 1217}, {M: 3, E: -145}, {M: -74, E: -280},
		{M: 191, E: 12}, {M: 517, E: -109}, {M: -175, E: 37}, {M: 456, E: -128}, {M: -87, E: -41}, {M: 50, E: -134}, {M: 94, E: -100}, {M: 1264, E: 670},
		{M: 1134, E: 1065}, {M: 1179, E: 5335}, {M: 6177, E: 10697}, {M: 2822, E: 24799}, {M: 14080, E: 38448}, {M: 57, E: 3}, {M: -68, E: 23}, {M: 64, E: 6039},
		{M: -2096, E: 4631}, {M: -1232, E: 192}, {M: -1323, E: -1347}, {M: 559, E: -1883}, {M: 2857, E: -529}, {M: 928, E: 550}, {M: -382, E: -3543}, {M: 10, E: -337},
		{M: 1332, E: 2602}, {M: 3564, E: 3096}, {M: 3546, E: 4189}, {M: 5867, E: 1205}, {M: -4045, E: -39}, {M: -23230, E: -2475}, {M: -2711, E: -2395}, {M: -10, E: -19},
		{M: -13, E: 2820}, {M: 1799, E: 3255}, {M: 1676, E: 3562}, {M: 1363, E: 2164}, {M: 1751, E: -812}, {M: -367, E: -3470}, {M: 3618, E: -1912}, {M: 5043, E: -1103},
		{M: 6314, E: 513}, {M: 3736, E: 658}, {M: 2130, E: 566}, {M: 531, E: 50}, {M: -4649, E: 6}, {M: -12688, E: -804}, {M: -737, E: -279}, {M: 1922, E: 512},
		{M: 718, E: -16}, {M: -831, E: 1430}, {M: -47, E: 1430}, {M: 257, E: 796}, {M: 4309, E: -651}, {M: 148, E: -1425}, {M: 5, E: -7610}, {M: -3, E: 597},
		{M: -1616, E: 225}, {M: -4851, E: 28}, {M: -1419, E: -29}, {M: 8711, E: 11}, {M: 8284, E: -2940}, {M: 128, E: -10962}, {M: -2933, E: -6748}, {M: 5889, E: -826},
		{M: 3596, E: 1207}, {M: -8382, E: 4394}, {M: -103, E: 3450}, {M: -2663, E: 4144}, {M: 8313, E: -123}, {M: 1865, E: -5032}, {M: -265, E: -74}, {M: 2644, E: -803},
		{M: 10663, E: -3710}, {M: 14022, E: -3944}, {M: -1933, E: -2043}, {M: 2361, E: 1106}, {M: -383, E: 157}, {M: -1190, E: -1346}, {M: 15271, E: 8671}, {M: -4336, E: 2437},
		{M: 6670, E: 9072}, {M: 8618, E: -1544}, {M: 3271, E: 1386},
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

func evaluatePawnsAndShelter(pos *Position, us Color, eval *Eval) {
	evaluatePawns(pos, us, eval)
	evaluateShelter(pos, us, eval)
}

func evaluatePawns(pos *Position, us Color, eval *Eval) {
	them := us.Opposite()
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

	if kingPawnDist != 8 {
		// Add a bonus for king protecting most advance pawn.
		eval.add(wPassedPawnKing[kingPawnDist])
	}
}

func evaluateShelter(pos *Position, us Color, eval *Eval) {
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
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	eval.merge(pawnsAndShelterCache.load(pos, us))
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	theirPawns := pos.ByPiece(them, Pawn)
	theirKing := pos.ByPiece(them, King)
	theirKingArea := bbKingArea[theirKing.AsSquare()]
	numAttackers, attackStrength := 0, int32(0) // opposite king attack strength

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

		if a := mobility & theirKingArea &^ theirPawns &^ excl; a != 0 {
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

		if a := mobility & theirKingArea &^ theirPawns &^ excl; a != 0 {
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

		if a := mobility & theirKingArea &^ theirPawns &^ excl; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.add(wFigure[Queen])
		mobility := QueenMobility(sq, all)
		eval.addN(wMobility[Queen], (mobility &^ excl).Count())

		if a := mobility & theirKingArea &^ theirPawns &^ excl; a != 0 {
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
