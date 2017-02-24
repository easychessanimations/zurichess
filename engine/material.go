// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// material.go implements position evaluation.

package engine

import "fmt"

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
	Weights = [187]Score{
		// 3009 ; train error = 0.0575463 ; validation error = 0.0576783
		{M: -19, E: -118}, {M: 15197, E: 11914}, {M: 62523, E: 49031}, {M: 64882, E: 52558}, {M: 88417, E: 92924}, {M: 212975, E: 154343}, {M: 31, E: 146}, {M: 59, E: 67},
		{M: 976, E: 2745}, {M: 1754, E: 402}, {M: 1551, E: 591}, {M: 1273, E: 429}, {M: 507, E: 1309}, {M: 208, E: -1574}, {M: 19, E: 140}, {M: -130, E: 94},
		{M: -80, E: 206}, {M: 60, E: -54}, {M: 134, E: -21}, {M: -10, E: -29}, {M: -43, E: 1}, {M: 7, E: 18}, {M: -2584, E: 1917}, {M: -1011, E: 231},
		{M: -1550, E: 350}, {M: -1487, E: -915}, {M: -2841, E: -9}, {M: 4785, E: -2016}, {M: 3630, E: -2329}, {M: -1745, E: -2203}, {M: -2098, E: 981}, {M: -2234, E: 810},
		{M: -467, E: -1300}, {M: -2657, E: -313}, {M: -1781, E: -1140}, {M: -13, E: -1164}, {M: 2076, E: -1950}, {M: -1602, E: -1059}, {M: -1913, E: 3273}, {M: -2494, E: 1874},
		{M: 449, E: -509}, {M: 1072, E: -1632}, {M: 583, E: -1335}, {M: 825, E: -1472}, {M: -1596, E: 245}, {M: -2810, E: 727}, {M: -1672, E: 4976}, {M: 22, E: 2868},
		{M: 37, E: 597}, {M: 1271, E: -1894}, {M: 241, E: -1099}, {M: 1106, E: -242}, {M: 63, E: 2255}, {M: -2490, E: 2483}, {M: 797, E: 8975}, {M: -70, E: 8027},
		{M: 1956, E: 4186}, {M: 227, E: -243}, {M: 7049, E: -3224}, {M: 11251, E: 4}, {M: 2442, E: 4073}, {M: -14, E: 6062}, {M: 133, E: 5094}, {M: 513, E: 2584},
		{M: -118, E: -56}, {M: -168, E: -3443}, {M: 3, E: -64}, {M: 200, E: -1750}, {M: -125, E: 0}, {M: -112, E: 3083}, {M: -53, E: 46}, {M: 8, E: 43},
		{M: -6, E: -24}, {M: -63, E: 49}, {M: -115, E: -59}, {M: 0, E: -25}, {M: 102, E: -4}, {M: -57, E: 43}, {M: -16, E: 2478}, {M: -64, E: 7},
		{M: -645, E: 185}, {M: -10, E: 534}, {M: 51, E: 4721}, {M: 3820, E: 10205}, {M: 5633, E: 21731}, {M: 10649, E: 39032}, {M: -58, E: 25}, {M: -43, E: 245},
		{M: 2059, E: 4847}, {M: 171, E: 2981}, {M: -1097, E: 9}, {M: -1639, E: -634}, {M: -189, E: 89}, {M: 2118, E: 1019}, {M: -48, E: 2650}, {M: -302, E: -2891},
		{M: -43, E: -70}, {M: 1250, E: 1927}, {M: 3207, E: 3675}, {M: 3491, E: 4139}, {M: 4486, E: 1532}, {M: 26, E: -1}, {M: -15624, E: -1404}, {M: -3309, E: -2196},
		{M: -141, E: -13}, {M: -15, E: 1965}, {M: 1255, E: 3509}, {M: 2037, E: 2783}, {M: 2331, E: 1500}, {M: 1121, E: -56}, {M: -274, E: -2425}, {M: -44, E: -1032},
		{M: 1426, E: -662}, {M: 2556, E: 595}, {M: 336, E: 779}, {M: 73, E: 728}, {M: 72, E: 10}, {M: -6696, E: 763}, {M: -8409, E: -237}, {M: -2209, E: -106},
		{M: 1420, E: -29}, {M: 1302, E: 519}, {M: -728, E: 1447}, {M: 16, E: 1502}, {M: 314, E: 1052}, {M: 3768, E: -724}, {M: -525, E: -32}, {M: 2456, E: -2913},
		{M: -373, E: -1554}, {M: 25, E: -1558}, {M: -502, E: 16}, {M: 68, E: 653}, {M: 374, E: 953}, {M: 2511, E: 1234}, {M: 18, E: 1794}, {M: -939, E: 321},
		{M: -21, E: 636}, {M: 573, E: 876}, {M: 2258, E: -28}, {M: 2364, E: -527}, {M: 3814, E: -1002}, {M: -1130, E: 302}, {M: 0, E: -1168}, {M: 17, E: -7547},
		{M: -996, E: 197}, {M: -669, E: -104}, {M: -3959, E: -92}, {M: 122, E: 603}, {M: 10468, E: 281}, {M: 173, E: 1128}, {M: 79, E: -7625}, {M: -183, E: -8152},
		{M: 3573, E: -1692}, {M: 811, E: 1486}, {M: -8945, E: 4047}, {M: -597, E: 2717}, {M: -3919, E: 3561}, {M: 4373, E: -131}, {M: 572, E: -6034}, {M: 22, E: 55},
		{M: 2507, E: -1000}, {M: 10579, E: -3582}, {M: 12232, E: -3842}, {M: -2576, E: -2254}, {M: 1572, E: 6}, {M: 1314, E: 1022}, {M: 1851, E: 1940}, {M: 3079, E: 1112},
		{M: 3081, E: 2922}, {M: 1715, E: 1233}, {M: 1933, E: 863}, {M: 2015, E: -53}, {M: -1059, E: 32}, {M: -1161, E: -1450}, {M: 9404, E: 1358}, {M: -3497, E: 1998},
		{M: 6059, E: 9472}, {M: 8528, E: -1778}, {M: 2206, E: 1269}, {M: -15, E: 219}, {M: 13847, E: -12}, {M: 5886, E: 6941}, {M: -2506, E: 9564}, {M: -1005, E: 3269},
		{M: -103, E: -68}, {M: 1388, E: -6362}, {M: 2309, E: -7659},
	}

	// The following variables are named chunks of Weights

	// wFigure stores how much each figure is valued.
	wFigure [FigureArraySize]Score
	// wMobility stores bonus for each figure's reachable square.
	wMobility [FigureArraySize]Score
	// wPawn is a piece square table dedicated to pawns.
	wPawn        [SquareArraySize]Score
	wEndgamePawn Score
	// wPassedPawn contains bonuses for passed pawns based on how advanced they are.
	wPassedPawn [8]Score
	// wPassedPawnKing is a bonus between king and closest passed pawn.
	wPassedPawnKing [8]Score
	// wFigureFile gives bonus to each figure depending on its file.
	wFigureFile [FigureArraySize][8]Score
	// wFigureRank gives bonus to each figure depending on its Rank.
	wFigureRank [FigureArraySize][8]Score
	wKingAttack [4]Score
	// wBackwardPawn is the bonus of a backward pawn.
	wBackwardPawn Score
	// wConnectedPawn is the bonus of a connected pawn.
	wConnectedPawn [8]Score
	// wDoublePawn is the bonus of a double pawn, a pawn with another
	// friendly in right in front of it.
	wDoublePawn Score
	// wIsolatedPawn is the bonus of an isolated pawn, a pawn with no
	// other friendlyy pawns on adjacent files.
	wIsolatedPawn Score
	// wPassedThreat is a small bonus for each enemy piece attacked by a pawn.
	wPawnThreat Score
	// wKingShelter rewards pawns in front of the king.
	wKingShelter Score
	// wBishopPair rewards the bishop pair, useful in endgames.
	wBishopPair Score
	// wBishopPair rewards a rook on a open file, a file with no pawns.
	wRookOnOpenFile Score
	// wBishopPair rewards a rook on a open file, a file with no enemy pawns.
	wRookOnHalfOpenFile Score
	// wQueenKingTropism rewards queen being closer to the enemy king.
	wQueenKingTropism [8]Score

	// FeatureNames is an array with the names of features.
	// When compiled with -tags coach, Score.I is the index in this array.
	FeatureNames [len(Weights)]string

	// Evaluation caches.
	pawnsAndShelterCache = &pawnsTable{}

	// Figure bonuses to use when computing the futility margin.
	futilityFigureBonus [FigureArraySize]int32
)

// scratchpad stores various information about evaluation of a single side.
type scratchpad struct {
	us            Color
	exclude       Bitboard // squares to exclude from mobility calculation
	kingSq        Square   // position of the king
	theirPawns    Bitboard // opponent's pawns
	theirKingArea Bitboard // opponent's king area

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
func (e *Eval) init(us Color) {
	pos := e.position
	them := us.Opposite()
	kingSq := pos.ByPiece(us, King).AsSquare()
	e.pad[us] = scratchpad{
		us:            us,
		exclude:       pos.ByPiece(us, Pawn) | PawnThreats(pos, them),
		kingSq:        kingSq,
		theirPawns:    pos.ByPiece(them, Pawn),
		theirKingArea: bbKingArea[pos.ByPiece(them, King).AsSquare()],
	}
}

// Feed returns the score phased between midgame and endgame score.
func (e *Eval) Feed(phase int32) int32 {
	return (e.Accum.M*(256-phase) + e.Accum.E*phase) / 256
}

// Evaluate evaluates position from White's POV.
// The returned s fits into a int16.
func Evaluate(pos *Position) int32 {
	e := EvaluatePosition(pos)
	score := e.Feed(Phase(pos))
	return scaleToCentipawns(score)
}

// EvaluatePosition evaluates position exported to be used by the tuner.
func EvaluatePosition(pos *Position) Eval {
	e := Eval{position: pos}
	e.init(White)
	e.init(Black)

	white, black := pawnsAndShelterCache.load(pos)
	e.pad[White].accum.merge(white)
	e.pad[Black].accum.merge(black)

	e.evaluateSide(White)
	e.evaluateSide(Black)

	e.Accum.merge(e.pad[White].accum)
	e.Accum.deduct(e.pad[Black].accum)
	return e
}

func evaluatePawnsAndShelter(pos *Position, us Color) (accum Accum) {
	accum.merge(evaluatePawns(pos, us))
	accum.merge(evaluateShelter(pos, us))
	return accum
}

func evaluatePawns(pos *Position, us Color) (accum Accum) {
	connected := ConnectedPawns(pos, us)
	double := DoubledPawns(pos, us)
	isolated := IsolatedPawns(pos, us)
	passed := PassedPawns(pos, us)
	backward := BackwardPawns(pos, us)

	kingPawnDist := int32(8)
	kingSq := pos.ByPiece(us, King).AsSquare()

	for bb := pos.ByPiece(us, Pawn); bb != 0; {
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

// evaluateFigure computes the material score for a figure fig at sq reaching mobility squares.
func (e *Eval) evaluateFigure(pad *scratchpad, fig Figure, sq Square, mobility Bitboard) {
	sq = sq.POV(pad.us)
	pad.accum.add(wFigure[fig])
	pad.accum.addN(wMobility[fig], (mobility &^ pad.exclude).Count())
	if fig != Queen {
		pad.accum.add(wFigureFile[fig][sq.File()])
		pad.accum.add(wFigureRank[fig][sq.Rank()])
	}

	if a := mobility & pad.theirKingArea &^ pad.theirPawns &^ pad.exclude; fig != King && a != 0 {
		pad.numAttackers++
		pad.attackStrength += a.CountMax2()
	}
}

// evaluateSide evaluates position for a single side.
func (e *Eval) evaluateSide(us Color) {
	pos := e.position
	them := us.Opposite()
	pad := &e.pad[us]
	all := pos.ByColor[White] | pos.ByColor[Black]

	// Pawn forward mobility.
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	pad.accum.addN(wMobility[Pawn], mobility.Count())
	mobility = PawnThreats(pos, us)
	pad.accum.addN(wPawnThreat, (mobility & pos.ByColor[them]).Count())

	if Majors(pos, them) == 0 {
		pad.accum.addN(wEndgamePawn, pos.ByPiece(us, Pawn).Count())
	}

	// Knight
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		mobility := KnightMobility(sq)
		e.evaluateFigure(pad, Knight, sq, mobility)
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		mobility := BishopMobility(sq, all)
		e.evaluateFigure(pad, Bishop, sq, mobility)
		numBishops++
	}
	pad.accum.addN(wBishopPair, numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		mobility := RookMobility(sq, all)
		e.evaluateFigure(pad, Rook, sq, mobility)

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
		e.evaluateFigure(pad, Queen, sq, mobility)
		pad.accum.add(wQueenKingTropism[distance[sq][e.pad[them].kingSq]])
	}

	// King, each side has one.
	{
		sq := pad.kingSq
		mobility := KingMobility(sq)
		e.evaluateFigure(pad, King, sq, mobility)
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
	curr -= (pos.NumPieces[WhiteKnight] + pos.NumPieces[BlackKnight]) * 1
	curr -= (pos.NumPieces[WhiteBishop] + pos.NumPieces[BlackBishop]) * 1
	curr -= (pos.NumPieces[WhiteRook] + pos.NumPieces[BlackRook]) * 3
	curr -= (pos.NumPieces[WhiteQueen] + pos.NumPieces[BlackQueen]) * 6
	return (curr*256 + total/2) / total
}

// scaleToCentipawns scales a score in the original scale to centipawns.
func scaleToCentipawns(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + 64 + score>>31) >> 7
}

// registerMany registers a slice of weights, setting the correct name.
func registerMany(n int, name string, out []Score) int {
	for i := range out {
		out[i] = Weights[n+i]
		FeatureNames[n+i] = fmt.Sprint(name, ".", i)
	}
	return n + len(out)
}

// registerone registers one weight, setting the correct name.
func registerOne(n int, name string, out *Score) int {
	*out = Weights[n]
	FeatureNames[n] = name
	return n + 1
}

func init() {
	// Initialize weights.
	initWeights()

	n := 0
	n = registerMany(n, "Figure", wFigure[:])
	n = registerMany(n, "Mobility", wMobility[:])
	n = registerMany(n, "Pawn", wPawn[:])
	n = registerOne(n, "EndgamePawn", &wEndgamePawn)
	n = registerMany(n, "PassedPawn", wPassedPawn[:])
	n = registerMany(n, "PassedPawnKing", wPassedPawnKing[:])
	n = registerMany(n, "FigureRank[Knight]", wFigureRank[Knight][:])
	n = registerMany(n, "FigureFile[Knight]", wFigureFile[Knight][:])
	n = registerMany(n, "FigureRank[Bishop]", wFigureRank[Bishop][:])
	n = registerMany(n, "FigureFile[Bishop]", wFigureFile[Bishop][:])
	n = registerMany(n, "FigureRank[Rook]", wFigureRank[Rook][:])
	n = registerMany(n, "FigureFile[Rook]", wFigureFile[Rook][:])
	n = registerMany(n, "FigureRank[King]", wFigureRank[King][:])
	n = registerMany(n, "FigureFile[King]", wFigureFile[King][:])
	n = registerMany(n, "KingAttack", wKingAttack[:])
	n = registerOne(n, "BackwardPawn", &wBackwardPawn)
	n = registerMany(n, "ConnectedPawn", wConnectedPawn[:])
	n = registerOne(n, "DoublePawn", &wDoublePawn)
	n = registerOne(n, "IsolatedPawn", &wIsolatedPawn)
	n = registerOne(n, "PawnThreat", &wPawnThreat)
	n = registerOne(n, "KingShelter", &wKingShelter)
	n = registerOne(n, "BishopPair", &wBishopPair)
	n = registerOne(n, "RookOnOpenFile", &wRookOnOpenFile)
	n = registerOne(n, "RookOnHalfOpenFile", &wRookOnHalfOpenFile)
	n = registerMany(n, "QueenKingTropism", wQueenKingTropism[:])

	if n != len(Weights) {
		panic(fmt.Sprintf("not all weights used, used %d out of %d", n, len(Weights)))
	}

	// Initialize futility figure bonus.
	for i, w := range wFigure {
		futilityFigureBonus[i] = scaleToCentipawns(max(w.M, w.E))
	}
}
