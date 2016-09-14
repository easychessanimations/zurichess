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
		{M: 76, E: -62}, {M: 15248, E: 12214}, {M: 65676, E: 48584}, {M: 64819, E: 51151}, {M: 92444, E: 95697}, {M: 220966, E: 157693}, {M: -129, E: 175}, {M: -83, E: 97},
		{M: 1189, E: 2685}, {M: 1756, E: 114}, {M: 1577, E: 515}, {M: 1210, E: 488}, {M: 318, E: 1579}, {M: 1, E: -1405}, {M: 290, E: -57}, {M: 443, E: -155},
		{M: -174, E: -283}, {M: -58, E: -27}, {M: -9, E: -17}, {M: -142, E: 185}, {M: -255, E: 359}, {M: -27, E: 22}, {M: -3827, E: 2133}, {M: -2950, E: 693},
		{M: -2325, E: 1450}, {M: -1143, E: -487}, {M: -2537, E: 1462}, {M: 4599, E: -1526}, {M: 3451, E: -2884}, {M: -2424, E: -2557}, {M: -3390, E: 1038}, {M: -3992, E: 904},
		{M: -786, E: -654}, {M: -1661, E: -9}, {M: -415, E: 0}, {M: 309, E: -972}, {M: 2059, E: -2670}, {M: -2058, E: -1519}, {M: -2829, E: 3010}, {M: -3996, E: 2076},
		{M: 313, E: -295}, {M: 2193, E: -1495}, {M: 1826, E: -849}, {M: 1315, E: -1312}, {M: -1836, E: -4}, {M: -3311, E: 334}, {M: -2793, E: 5297}, {M: -1092, E: 2972},
		{M: 6, E: 1197}, {M: 1871, E: -1595}, {M: 1214, E: -410}, {M: 2347, E: -276}, {M: 29, E: 1891}, {M: -3022, E: 2665}, {M: 1270, E: 9111}, {M: -7, E: 8024},
		{M: 4173, E: 3890}, {M: 4359, E: -1492}, {M: 11242, E: -4219}, {M: 16897, E: -1317}, {M: 6653, E: 3992}, {M: 781, E: 6527}, {M: -17, E: 7293}, {M: 2422, E: 3792},
		{M: 48, E: 23}, {M: 3125, E: -4298}, {M: -94, E: -826}, {M: 9825, E: -5146}, {M: -13931, E: 4107}, {M: -20817, E: 11031}, {M: 198, E: 297}, {M: 33, E: 243},
		{M: 122, E: 57}, {M: -169, E: -425}, {M: -114, E: 135}, {M: -30, E: 297}, {M: -167, E: -103}, {M: 26, E: 139}, {M: 17, E: 3}, {M: -456, E: 1143},
		{M: -24, E: 1560}, {M: -220, E: 5924}, {M: 4553, E: 10604}, {M: 4860, E: 22458}, {M: 17311, E: 36992}, {M: 273, E: -27}, {M: 224, E: 37}, {M: 2276, E: 3723},
		{M: 597, E: 2005}, {M: -558, E: -1274}, {M: -1409, E: -1906}, {M: -136, E: -734}, {M: 3515, E: -537}, {M: 365, E: 1636}, {M: -436, E: -3115}, {M: 347, E: -219},
		{M: 2099, E: 2166}, {M: 4156, E: 3775}, {M: 3983, E: 4458}, {M: 5929, E: 1381}, {M: -705, E: 4}, {M: -19470, E: -1721}, {M: -3384, E: -2742}, {M: -2, E: 11},
		{M: -91, E: 2671}, {M: 1457, E: 3809}, {M: 2051, E: 3264}, {M: 2328, E: 2057}, {M: 1534, E: -383}, {M: -462, E: -3380}, {M: 2903, E: -860}, {M: 4921, E: -509},
		{M: 6180, E: 737}, {M: 3664, E: 1089}, {M: 2900, E: 1198}, {M: 3681, E: -16}, {M: -4543, E: 1307}, {M: -7010, E: -100}, {M: -1568, E: -18}, {M: 2487, E: -409},
		{M: 1865, E: 595}, {M: -135, E: 1598}, {M: 704, E: 1522}, {M: 1079, E: 998}, {M: 4817, E: -1000}, {M: -39, E: -365}, {M: 0, E: -7606}, {M: -705, E: -27},
		{M: -1681, E: -36}, {M: -5665, E: 31}, {M: -45, E: 865}, {M: 11540, E: 614}, {M: 6896, E: 623}, {M: 8143, E: -10703}, {M: -102, E: -7247}, {M: 6554, E: -1827},
		{M: 2569, E: 1652}, {M: -9274, E: 4794}, {M: -1474, E: 3665}, {M: -4320, E: 4373}, {M: 6765, E: -75}, {M: 2767, E: -5761}, {M: -172, E: -143}, {M: 3223, E: -745},
		{M: 11681, E: -3636}, {M: 16122, E: -5851}, {M: -2679, E: -2152}, {M: 2201, E: 1071}, {M: -1266, E: 762}, {M: -1256, E: -1468}, {M: 9724, E: 1209}, {M: -3433, E: 1984},
		{M: 6242, E: 9277}, {M: 8769, E: -1552}, {M: 2130, E: 1414},
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

// scratchpad stores various information about evaluation of a single side.
type scratchpad struct {
	us            Color
	exclude       Bitboard // squares to exclude from mobility calculation
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
	eval.pad[us] = scratchpad{
		us:            us,
		exclude:       pos.ByPiece(us, Pawn) | pos.PawnThreats(them),
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
			accum.add(wConnectedPawn)
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
	}

	// King, each side has one.
	{
		sq := pos.ByPiece(us, King).AsSquare()
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
		futilityFigureBonus[i] = scaleToCentipawns(max(w.M, w.E))
	}
}
