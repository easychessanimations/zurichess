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
	Weights = [162]Score{
		{M: 150, E: 193}, {M: 15392, E: 12084}, {M: 54509, E: 47222}, {M: 57423, E: 42713}, {M: 92437, E: 95697}, {M: 220316, E: 158067}, {M: -512, E: 137}, {M: 251, E: 61},
		{M: 1185, E: 2711}, {M: 1742, E: 107}, {M: 1581, E: 512}, {M: 1185, E: 466}, {M: 321, E: 1586}, {M: 0, E: -1407}, {M: -204, E: -154}, {M: -27, E: 42},
		{M: 151, E: -25}, {M: 93, E: -41}, {M: -142, E: 91}, {M: 7, E: 129}, {M: 109, E: -70}, {M: -8, E: 142}, {M: -3295, E: 2735}, {M: -2080, E: 1116},
		{M: -2222, E: 1241}, {M: -1983, E: -620}, {M: -3751, E: 705}, {M: 5068, E: -1906}, {M: 3714, E: -2321}, {M: -2468, E: -1900}, {M: -2870, E: 1718}, {M: -3006, E: 1460},
		{M: -822, E: -966}, {M: -2701, E: -331}, {M: -1826, E: -1121}, {M: 644, E: -1486}, {M: 2349, E: -1963}, {M: -2081, E: -801}, {M: -2669, E: 3596}, {M: -3305, E: 2572},
		{M: 294, E: -509}, {M: 1387, E: -1987}, {M: 834, E: -1623}, {M: 1563, E: -1840}, {M: -1771, E: 622}, {M: -3374, E: 869}, {M: -2669, E: 5705}, {M: -584, E: 3413},
		{M: -13, E: 1072}, {M: 1244, E: -1840}, {M: 394, E: -1010}, {M: 2397, E: -486}, {M: 19, E: 2393}, {M: -3166, E: 3011}, {M: 1382, E: 9463}, {M: 268, E: 8328},
		{M: 4128, E: 3985}, {M: 4106, E: -1468}, {M: 11140, E: -4362}, {M: 16847, E: -1310}, {M: 6690, E: 4325}, {M: 643, E: 6831}, {M: 0, E: 7166}, {M: 2542, E: 3674},
		{M: -50, E: 25}, {M: 3350, E: -4604}, {M: 101, E: -1013}, {M: 9949, E: -5299}, {M: -14066, E: 4080}, {M: -21190, E: 10847}, {M: -454, E: -132}, {M: 175, E: -161},
		{M: -82, E: -158}, {M: 144, E: 185}, {M: -203, E: 144}, {M: -218, E: 30}, {M: 27, E: -72}, {M: 65, E: 77}, {M: -63, E: -160}, {M: -836, E: 1027},
		{M: -285, E: 1516}, {M: -371, E: 5826}, {M: 4471, E: 10484}, {M: 4897, E: 22246}, {M: 17389, E: 37156}, {M: -83, E: 236}, {M: 22, E: 107}, {M: 2539, E: 3811},
		{M: 955, E: 2008}, {M: -255, E: -1306}, {M: -1246, E: -1962}, {M: -208, E: -893}, {M: 3348, E: -810}, {M: 215, E: 1205}, {M: 6655, E: -2858}, {M: 7518, E: 4},
		{M: 9271, E: 2431}, {M: 11320, E: 4000}, {M: 11157, E: 4711}, {M: 13113, E: 1618}, {M: 6238, E: 280}, {M: -12370, E: -1488}, {M: 589, E: -1533}, {M: 4096, E: 1155},
		{M: 3964, E: 3872}, {M: 5513, E: 4961}, {M: 6108, E: 4421}, {M: 6367, E: 3238}, {M: 5601, E: 707}, {M: 3495, E: -2195}, {M: 6592, E: 3587}, {M: 8633, E: 3956},
		{M: 9869, E: 5240}, {M: 7354, E: 5600}, {M: 6578, E: 5715}, {M: 7402, E: 4489}, {M: -831, E: 5857}, {M: -3243, E: 4295}, {M: 2033, E: 3935}, {M: 6145, E: 3524},
		{M: 5507, E: 4542}, {M: 3448, E: 5571}, {M: 4284, E: 5499}, {M: 4686, E: 4958}, {M: 8468, E: 2939}, {M: 3548, E: 3587}, {M: -3, E: -7625}, {M: -678, E: -53},
		{M: -1654, E: -74}, {M: -5607, E: 9}, {M: 95, E: 829}, {M: 11722, E: 575}, {M: 6934, E: 630}, {M: 8585, E: -10815}, {M: -36, E: -7282}, {M: 6704, E: -1837},
		{M: 2712, E: 1716}, {M: -9153, E: 4893}, {M: -1351, E: 3755}, {M: -4184, E: 4452}, {M: 6946, E: -37}, {M: 2895, E: -5740}, {M: -77, E: 134}, {M: 3221, E: -742},
		{M: 11669, E: -3627}, {M: 16132, E: -5870}, {M: -2720, E: -2078}, {M: 1567, E: -180}, {M: 1822, E: 1130}, {M: 1935, E: 1883}, {M: 2958, E: 1590}, {M: 3429, E: 3211},
		{M: 1320, E: 1473}, {M: 2176, E: 1127}, {M: 2177, E: -166}, {M: -1376, E: 552}, {M: -1252, E: -1326}, {M: 9769, E: 1184}, {M: -3447, E: 1962}, {M: 6260, E: 9272},
		{M: 8750, E: -1591}, {M: 2091, E: 1374},
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
	w = slice(w, wConnectedPawn[:])
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
