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
	Weights = [186]Score{
		{M: 32, E: -19}, {M: 14789, E: 12497}, {M: 63250, E: 47211}, {M: 66052, E: 50717}, {M: 85787, E: 96681}, {M: 202111, E: 164872}, {M: 102, E: 95}, {M: -137, E: -321},
		{M: 1029, E: 2587}, {M: 1735, E: 408}, {M: 1555, E: 586}, {M: 1227, E: 428}, {M: 512, E: 1219}, {M: 3, E: -1358}, {M: 55, E: 94}, {M: 10, E: 298},
		{M: -135, E: 319}, {M: 335, E: 142}, {M: -128, E: 141}, {M: -142, E: -7}, {M: -255, E: -44}, {M: 479, E: -127}, {M: -2419, E: 1706}, {M: -941, E: 68},
		{M: -1665, E: 383}, {M: -1493, E: -699}, {M: -3125, E: 25}, {M: 4655, E: -1922}, {M: 3344, E: -2316}, {M: -1978, E: -2292}, {M: -2070, E: 871}, {M: -2045, E: 546},
		{M: -632, E: -1261}, {M: -2677, E: -261}, {M: -1862, E: -1136}, {M: 29, E: -1264}, {M: 1831, E: -1818}, {M: -1799, E: -1083}, {M: -1988, E: 2836}, {M: -2412, E: 1724},
		{M: 398, E: -695}, {M: 1242, E: -1940}, {M: 672, E: -1604}, {M: 890, E: -1687}, {M: -1837, E: 449}, {M: -3064, E: 509}, {M: -1408, E: 4791}, {M: 47, E: 2576},
		{M: 63, E: 699}, {M: 999, E: -1734}, {M: 101, E: -967}, {M: 1465, E: -199}, {M: -71, E: 2170}, {M: -2557, E: 2541}, {M: 711, E: 9007}, {M: -124, E: 7714},
		{M: 1682, E: 3901}, {M: 215, E: -172}, {M: 5467, E: -2361}, {M: 12118, E: 36}, {M: 3184, E: 4471}, {M: 136, E: 6476}, {M: -20, E: 5914}, {M: 419, E: 2849},
		{M: 228, E: 114}, {M: -158, E: -1560}, {M: -197, E: 116}, {M: 27, E: -523}, {M: 41, E: -11}, {M: 15, E: 4239}, {M: -204, E: -210}, {M: -27, E: -131},
		{M: -272, E: 71}, {M: -261, E: -38}, {M: -381, E: -152}, {M: 139, E: 155}, {M: -289, E: -101}, {M: -289, E: -176}, {M: 58, E: -188}, {M: -592, E: 684},
		{M: -77, E: 1148}, {M: -301, E: 5468}, {M: 3921, E: 10246}, {M: 6413, E: 21179}, {M: 13163, E: 36968}, {M: 120, E: -4}, {M: -7, E: -147}, {M: 1308, E: 4229},
		{M: 39, E: 2291}, {M: -319, E: -1003}, {M: -1020, E: -1618}, {M: -142, E: -458}, {M: 2382, E: 123}, {M: -9, E: 1157}, {M: -249, E: -2680}, {M: -88, E: -48},
		{M: 1589, E: 1746}, {M: 3562, E: 3417}, {M: 3701, E: 4001}, {M: 5058, E: 1110}, {M: -194, E: -55}, {M: -15482, E: -1577}, {M: -3235, E: -1997}, {M: -52, E: -34},
		{M: 4, E: 2098}, {M: 1463, E: 3306}, {M: 2121, E: 2790}, {M: 2290, E: 1615}, {M: 1295, E: -58}, {M: -290, E: -2478}, {M: 11, E: -971}, {M: 1538, E: -703},
		{M: 2704, E: 466}, {M: 362, E: 731}, {M: -86, E: 780}, {M: 131, E: -21}, {M: -6389, E: 791}, {M: -7544, E: -229}, {M: -2090, E: 87}, {M: 1494, E: -305},
		{M: 1113, E: 384}, {M: -814, E: 1394}, {M: 14, E: 1334}, {M: 198, E: 872}, {M: 3751, E: -790}, {M: -801, E: -20}, {M: 2404, E: -2749}, {M: -500, E: -1565},
		{M: 23, E: -1543}, {M: -383, E: -15}, {M: 65, E: 812}, {M: 441, E: 818}, {M: 2884, E: 1198}, {M: -114, E: 1847}, {M: -786, E: 175}, {M: -48, E: 467},
		{M: 808, E: 619}, {M: 2210, E: -1}, {M: 2750, E: -781}, {M: 3901, E: -1228}, {M: -1011, E: 39}, {M: -12, E: -1480}, {M: 30, E: -7215}, {M: -416, E: 71},
		{M: -373, E: -59}, {M: -3074, E: -190}, {M: 131, E: 1041}, {M: 5717, E: 2050}, {M: -6, E: 1895}, {M: -39, E: -7264}, {M: -873, E: -7324}, {M: 4026, E: -1655},
		{M: 1235, E: 1334}, {M: -8104, E: 3889}, {M: -180, E: 2665}, {M: -3971, E: 3588}, {M: 4424, E: -64}, {M: 231, E: -5392}, {M: -200, E: -4}, {M: 2571, E: -971},
		{M: 10500, E: -3499}, {M: 13570, E: -4550}, {M: -2591, E: -2009}, {M: 1318, E: -8}, {M: 1380, E: 1213}, {M: 1768, E: 1848}, {M: 2932, E: 1158}, {M: 3305, E: 2938},
		{M: 1591, E: 1173}, {M: 2315, E: 830}, {M: 1879, E: -77}, {M: -1055, E: 25}, {M: -1276, E: -1342}, {M: 9433, E: 1271}, {M: -3457, E: 2004}, {M: 6019, E: 8934},
		{M: 8205, E: -1688}, {M: 2052, E: 1342}, {M: -87, E: 30}, {M: 14758, E: 146}, {M: 6438, E: 5806}, {M: -2371, E: 8846}, {M: -913, E: 2804}, {M: -13, E: -102},
		{M: 1470, E: -7270}, {M: 2356, E: -9148},
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
	if fig == Knight || fig == Bishop || fig == Rook || fig == King {
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
	w = slice(w, wFigureRank[Rook][:])
	w = slice(w, wFigureFile[Rook][:])
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
