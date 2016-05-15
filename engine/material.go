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
	Weights = [106]Score{
		{M: -48, E: -197}, {M: 14305, E: 11977}, {M: 63129, E: 45056}, {M: 67751, E: 51779}, {M: 89497, E: 95498}, {M: 194880, E: 177272}, {M: -194, E: 148}, {M: -5, E: -23},
		{M: 1308, E: 2568}, {M: 2585, E: 1564}, {M: 1688, E: 737}, {M: 1340, E: 540}, {M: 434, E: 1292}, {M: -388, E: -1534}, {M: -2479, E: 2796}, {M: -2633, E: 2632},
		{M: -2372, E: 1753}, {M: -53, E: -318}, {M: -2680, E: -16}, {M: 5294, E: -1130}, {M: 3531, E: -1650}, {M: -2597, E: -2385}, {M: -2224, E: 1607}, {M: -4007, E: 2081},
		{M: -1232, E: -58}, {M: -2129, E: -1177}, {M: -52, E: 286}, {M: 23, E: -260}, {M: 1764, E: -1582}, {M: -2557, E: -848}, {M: -2309, E: 3980}, {M: -4601, E: 3469},
		{M: 433, E: -49}, {M: 2468, E: -1636}, {M: 991, E: -704}, {M: 1225, E: -1058}, {M: -1490, E: 445}, {M: -3576, E: 868}, {M: -1960, E: 6101}, {M: -3107, E: 5417},
		{M: 228, E: 1245}, {M: 1636, E: -2320}, {M: 57, E: -402}, {M: 2061, E: 160}, {M: 10, E: 1812}, {M: -4100, E: 3280}, {M: -64, E: 9256}, {M: 774, E: 6799},
		{M: 9631, E: -645}, {M: 6694, E: -4773}, {M: 14417, E: -6842}, {M: 11401, E: -242}, {M: 6446, E: 2744}, {M: 4732, E: 5393}, {M: 3338, E: 1512}, {M: 4808, E: -27},
		{M: 105, E: 197}, {M: 304, E: -5895}, {M: -170, E: -3773}, {M: 4517, E: -3231}, {M: -19689, E: 8711}, {M: -12345, E: 76}, {M: 112, E: 119}, {M: 1560, E: 1316},
		{M: 1147, E: 1880}, {M: 622, E: 5979}, {M: 5384, E: 10933}, {M: 3264, E: 25213}, {M: 16315, E: 39103}, {M: 16, E: 123}, {M: -24, E: -75}, {M: -2267, E: 4528},
		{M: -1895, E: 3267}, {M: -1164, E: -559}, {M: -5547, E: -758}, {M: -8037, E: 12}, {M: -9102, E: 2026}, {M: -5211, E: 871}, {M: 1351, E: -9069}, {M: -1, E: -57},
		{M: -2474, E: 848}, {M: -5125, E: 1141}, {M: -5116, E: 2042}, {M: 3214, E: 2151}, {M: 1829, E: -683}, {M: -141, E: -9639}, {M: -294, E: -7841}, {M: 6227, E: -458},
		{M: 3730, E: 2071}, {M: -9084, E: 5498}, {M: -504, E: 4429}, {M: -2178, E: 4695}, {M: 8819, E: -73}, {M: 4297, E: -6127}, {M: 3, E: 253}, {M: 3035, E: -1054},
		{M: 11315, E: -4018}, {M: 13936, E: -3868}, {M: 2265, E: 807}, {M: -632, E: 309}, {M: -973, E: -1189}, {M: 15556, E: 8738}, {M: -4565, E: 2298}, {M: 7482, E: 8668},
		{M: 8235, E: -1635}, {M: 3112, E: 1038},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
	wPassedPawn         [8]Score
	wPassedPawnKing     [8]Score
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
