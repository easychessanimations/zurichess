// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// material.go implements position evaluation.
//
// Zurichess' evaluation is a very simple neural network with no hidden layers,
// and one output node y = W_m * x * (1-p) + W_e * x * p where W_m are
// middle game weights, W_e are endgame weights, x is input, p is phase between
// middle game and end game, and y is the score.
// The network has |x| = len(Weights) inputs corresponding to features
// extracted from the position. These features are symmetrical wrt colors.
// The network is trained using the Texel's Tuning Method
// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.

package engine

import (
	"fmt"

	. "bitbucket.org/zurichess/zurichess/board"
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
	theirKing := pos.ByPiece(them, King)

	e.pad[us] = scratchpad{
		us:            us,
		exclude:       pos.ByPiece(us, Pawn) | PawnThreats(pos, them),
		kingSq:        kingSq,
		theirPawns:    pos.ByPiece(them, Pawn),
		theirKingArea: BbKingArea[theirKing.AsSquare()],
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
