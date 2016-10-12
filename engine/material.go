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
	Weights = [170]Score{
		{M: -63, E: -38}, {M: 14667, E: 12656}, {M: 63291, E: 47165}, {M: 66049, E: 50575}, {M: 87603, E: 95305}, {M: 206829, E: 160843}, {M: 28, E: -148}, {M: 54, E: 125},
		{M: 1047, E: 2689}, {M: 1749, E: 400}, {M: 1572, E: 572}, {M: 1277, E: 457}, {M: 496, E: 1278}, {M: -23, E: -1312}, {M: -21, E: 274}, {M: 56, E: -262},
		{M: -221, E: 293}, {M: -29, E: -282}, {M: -322, E: -54}, {M: -223, E: 13}, {M: -90, E: 88}, {M: 151, E: -216}, {M: -1780, E: 1554}, {M: -768, E: 43},
		{M: -1537, E: 492}, {M: -1361, E: -592}, {M: -3056, E: 50}, {M: 4658, E: -1883}, {M: 3458, E: -2418}, {M: -1669, E: -2476}, {M: -1557, E: 680}, {M: -1769, E: 440},
		{M: -411, E: -1334}, {M: -2516, E: -323}, {M: -1709, E: -1165}, {M: 20, E: -1342}, {M: 1943, E: -1984}, {M: -1605, E: -1307}, {M: -1615, E: 2675}, {M: -2278, E: 1598},
		{M: 518, E: -805}, {M: 1331, E: -2023}, {M: 791, E: -1708}, {M: 882, E: -1758}, {M: -1744, E: 269}, {M: -2901, E: 274}, {M: -1339, E: 4683}, {M: 23, E: 2552},
		{M: 62, E: 517}, {M: 1112, E: -1837}, {M: 174, E: -1079}, {M: 1417, E: -344}, {M: -62, E: 1903}, {M: -2554, E: 2298}, {M: 348, E: 9002}, {M: 260, E: 7558},
		{M: 1612, E: 3719}, {M: 408, E: -343}, {M: 5154, E: -2402}, {M: 11770, E: 3}, {M: 2552, E: 4385}, {M: 210, E: 6156}, {M: 353, E: 6058}, {M: 157, E: 2895},
		{M: 40, E: 15}, {M: -25, E: -1385}, {M: -270, E: 21}, {M: 87, E: -504}, {M: -81, E: 30}, {M: 44, E: 4111}, {M: 10, E: 66}, {M: -73, E: -71},
		{M: -16, E: -183}, {M: -303, E: 385}, {M: 141, E: -138}, {M: -77, E: 95}, {M: 123, E: -240}, {M: 106, E: -230}, {M: -33, E: -79}, {M: -532, E: 601},
		{M: -41, E: 1086}, {M: -236, E: 5394}, {M: 3941, E: 10121}, {M: 6677, E: 21095}, {M: 13118, E: 36941}, {M: 178, E: -104}, {M: 40, E: 171}, {M: 1242, E: 4257},
		{M: 83, E: 2349}, {M: -248, E: -988}, {M: -1107, E: -1588}, {M: -121, E: -362}, {M: 2217, E: 106}, {M: -57, E: 1093}, {M: -284, E: -2579}, {M: -1, E: -95},
		{M: 1576, E: 1844}, {M: 3584, E: 3440}, {M: 3680, E: 3932}, {M: 5104, E: 929}, {M: -305, E: 26}, {M: -14974, E: -1829}, {M: -3194, E: -2002}, {M: -32, E: -89},
		{M: 1, E: 2107}, {M: 1458, E: 3272}, {M: 2120, E: 2713}, {M: 2240, E: 1552}, {M: 1263, E: -94}, {M: -365, E: -2500}, {M: -2, E: -800}, {M: 1647, E: -459},
		{M: 2774, E: 674}, {M: 390, E: 903}, {M: -63, E: 836}, {M: 215, E: -12}, {M: -6532, E: 754}, {M: -7364, E: -294}, {M: -2096, E: -46}, {M: 1548, E: -224},
		{M: 1151, E: 421}, {M: -849, E: 1434}, {M: 14, E: 1337}, {M: 116, E: 898}, {M: 3809, E: -750}, {M: -745, E: -43}, {M: 467, E: -7207}, {M: -117, E: 112},
		{M: -131, E: -45}, {M: -2733, E: -195}, {M: 65, E: 1002}, {M: 4841, E: 2011}, {M: 70, E: 1680}, {M: 356, E: -7192}, {M: -220, E: -7111}, {M: 4553, E: -1474},
		{M: 1690, E: 1545}, {M: -9110, E: 4443}, {M: -1881, E: 3414}, {M: -4657, E: 4140}, {M: 5749, E: -138}, {M: 1250, E: -5404}, {M: 33, E: -205}, {M: 2528, E: -684},
		{M: 10512, E: -3257}, {M: 13341, E: -4140}, {M: -2575, E: -1985}, {M: 1276, E: -18}, {M: 1262, E: 1305}, {M: 1719, E: 1799}, {M: 2918, E: 1138}, {M: 3296, E: 2893},
		{M: 1591, E: 1134}, {M: 2311, E: 809}, {M: 2003, E: -59}, {M: -1045, E: 113}, {M: -1324, E: -1316}, {M: 9577, E: 1205}, {M: -3457, E: 1994}, {M: 5880, E: 9055},
		{M: 8253, E: -1395}, {M: 1892, E: 1335}, {M: 200, E: -52}, {M: 11696, E: 19}, {M: 6357, E: 5300}, {M: -2344, E: 8445}, {M: -945, E: 2630}, {M: -6, E: -112},
		{M: 1375, E: -7045}, {M: 2149, E: -9050},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [SquareArraySize]Score
	wPassedPawn         [8]Score
	wPassedPawnKing     [8]Score
	wFigureFile         [FigureArraySize][8]Score
	wFigureRank         [FigureArraySize][8]Score
	wKingAttack         [4]Score
	wBackwardPawn       Score
	wConnectedPawn      [8]Score
	wDoublePawn         Score
	wIsolatedPawn       Score
	wPawnThreat         Score
	wKingShelter        Score
	wBishopPair         Score
	wRookOnOpenFile     Score
	wRookOnHalfOpenFile Score
	wQueenKingTropism   [8]Score

	// Evaluation caches.
	pawnsAndShelterCache *pawnsTable

	// Figure bonuses to use when computing the futility margin.
	futilityFigureBonus [FigureArraySize]int32
)

// scratchpad stores various information about evaluation of a single side.
type scratchpad struct {
	us            Color
	exclude       Bitboard // squares to exclude from mobility calculation
	kingSq        Square
	theirPawns    Bitboard
	theirKingArea Bitboard

	accum          Accum
	numAttackers   int32 // number of pieces attacking opposite king
	attackStrength int32 // strength of the attack
}

// Eval contains necessary information for evaluation.
type Eval struct {
	Accum    Accum
	position *Position
	pad      [ColorArraySize]scratchpad
}

// init initializes some used info.
func (eval *Eval) init(us Color) {
	pos := eval.position
	them := us.Opposite()
	kingSq := pos.ByPiece(us, King).AsSquare()
	eval.pad[us] = scratchpad{
		us:            us,
		exclude:       pos.ByPiece(us, Pawn) | pos.PawnThreats(them),
		kingSq:        kingSq,
		theirPawns:    pos.ByPiece(them, Pawn),
		theirKingArea: bbKingArea[pos.ByPiece(them, King).AsSquare()],
	}
}

// Feed return the score phased between midgame and endgame score.
func (e *Eval) Feed(phase int32) int32 {
	return (e.Accum.M*(256-phase) + e.Accum.E*phase) / 256
}

// Evaluate evaluates position from White's POV.
// The returned s fits into a int16.
func Evaluate(pos *Position) int32 {
	eval := EvaluatePosition(pos)
	score := eval.Feed(Phase(pos))
	return scaleToCentipawns(score)
}

// EvaluatePosition evaluates position exported to be used by the tuner.
func EvaluatePosition(pos *Position) Eval {
	eval := Eval{position: pos}
	eval.init(White)
	eval.init(Black)
	eval.evaluateSide(White)
	eval.evaluateSide(Black)
	eval.merge()
	return eval
}

func evaluatePawnsAndShelter(pos *Position, us Color) (accum Accum) {
	accum.merge(evaluatePawns(pos, us))
	accum.merge(evaluateShelter(pos, us))
	return accum
}

func evaluatePawns(pos *Position, us Color) (accum Accum) {
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

		accum.add(wFigure[Pawn])
		accum.add(wPawn[povSq])

		if passed.Has(sq) {
			accum.add(wPassedPawn[rank])
			if kingPawnDist > distance[sq][kingSq] {
				kingPawnDist = distance[sq][kingSq]
			}
		}
		if connected.Has(sq) {
			accum.add(wConnectedPawn[povSq.File()])
		}
		if double.Has(sq) {
			accum.add(wDoublePawn)
		}
		if isolated.Has(sq) {
			accum.add(wIsolatedPawn)
		}
		if backward.Has(sq) {
			accum.add(wBackwardPawn)
		}
	}

	if kingPawnDist != 8 {
		// Add a bonus for king protecting most advance pawn.
		accum.add(wPassedPawnKing[kingPawnDist])
	}

	return accum
}

// evaluateShelter evaluates king's shelter.
func evaluateShelter(pos *Position, us Color) (accum Accum) {
	pawns := pos.ByPiece(us, Pawn)
	king := pos.ByPiece(us, King)
	sq := king.AsSquare().POV(us)
	king = ForwardSpan(us, king)
	file := sq.File()
	if file > 0 && West(king)&pawns == 0 {
		accum.add(wKingShelter)
	}
	if king&pawns == 0 {
		accum.addN(wKingShelter, 2)
	}
	if file < 7 && East(king)&pawns == 0 {
		accum.add(wKingShelter)
	}
	return accum
}

// evaluateFigure computes the material score for figure.
func (eval *Eval) evaluateFigure(pad *scratchpad, fig Figure, sq Square, mobility Bitboard) {
	sq = sq.POV(pad.us)
	pad.accum.add(wFigure[fig])
	pad.accum.addN(wMobility[fig], (mobility &^ pad.exclude).Count())
	if fig == Knight || fig == Bishop || fig == King {
		pad.accum.add(wFigureFile[fig][sq.File()])
		pad.accum.add(wFigureRank[fig][sq.Rank()])
	}

	if a := mobility & pad.theirKingArea &^ pad.theirPawns &^ pad.exclude; fig != King && a != 0 {
		pad.numAttackers++
		pad.attackStrength += a.CountMax2()
	}
}

// evaluateSide evaluates position for a single side.
func (eval *Eval) evaluateSide(us Color) {
	pos := eval.position
	them := us.Opposite()
	pad := &eval.pad[us]
	all := pos.ByColor[White] | pos.ByColor[Black]

	pad.accum.merge(pawnsAndShelterCache.load(pos, us))

	// Pawn forward mobility.
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	pad.accum.addN(wMobility[Pawn], mobility.Count())
	mobility = pos.PawnThreats(us)
	pad.accum.addN(wPawnThreat, (mobility & pos.ByColor[them]).Count())

	// Knight
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		mobility := KnightMobility(sq)
		eval.evaluateFigure(pad, Knight, sq, mobility)
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		mobility := BishopMobility(sq, all)
		eval.evaluateFigure(pad, Bishop, sq, mobility)
		numBishops++
	}
	pad.accum.addN(wBishopPair, numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		mobility := RookMobility(sq, all)
		eval.evaluateFigure(pad, Rook, sq, mobility)

		// Evaluate rook on open and semi open files.
		// https://chessprogramming.wikispaces.com/Rook+on+Open+File
		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				pad.accum.add(wRookOnOpenFile)
			} else {
				pad.accum.add(wRookOnHalfOpenFile)
			}
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		mobility := QueenMobility(sq, all)
		eval.evaluateFigure(pad, Queen, sq, mobility)
		pad.accum.add(wQueenKingTropism[distance[sq][eval.pad[them].kingSq]])
	}

	// King, each side has one.
	{
		sq := pad.kingSq
		mobility := KingMobility(sq)
		eval.evaluateFigure(pad, King, sq, mobility)
	}

	// Evaluate attacking the king. See more at:
	// https://chessprogramming.wikispaces.com/King+Safety#Attacking%20King%20Zone
	pad.numAttackers = min(pad.numAttackers, int32(len(wKingAttack)-1))
	pad.accum.addN(wKingAttack[pad.numAttackers], pad.attackStrength)
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

// scaleToCentipawns scales a score in the original scale to centipawns.
func scaleToCentipawns(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + 64 + score>>31) >> 7
}

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
	w = slice(w, wFigureRank[Knight][:])
	w = slice(w, wFigureFile[Knight][:])
	w = slice(w, wFigureRank[Bishop][:])
	w = slice(w, wFigureFile[Bishop][:])
	w = slice(w, wFigureRank[King][:])
	w = slice(w, wFigureFile[King][:])
	w = slice(w, wKingAttack[:])
	w = entry(w, &wBackwardPawn)
	w = slice(w, wConnectedPawn[:])
	w = entry(w, &wDoublePawn)
	w = entry(w, &wIsolatedPawn)
	w = entry(w, &wPawnThreat)
	w = entry(w, &wKingShelter)
	w = entry(w, &wBishopPair)
	w = entry(w, &wRookOnOpenFile)
	w = entry(w, &wRookOnHalfOpenFile)
	w = slice(w, wQueenKingTropism[:])

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}

	// Initialize caches.
	pawnsAndShelterCache = new(pawnsTable)

	// Initializes futility figure bonus
	for i, w := range wFigure {
		futilityFigureBonus[i] = scaleToCentipawns(max(w.M, w.E))
	}
}
