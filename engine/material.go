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

	//white, black := pawnsAndShelterCache.load(pos)
	//e.pad[White].accum.merge(white)
	//e.pad[Black].accum.merge(black)

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
	pad := &e.pad[us]

	groupByBoard(fNoFigure, BbEmpty, &pad.accum)
	groupByBoard(fPawn, pos.ByPiece(us, Pawn), &pad.accum)
	groupByBoard(fKnight, pos.ByPiece(us, Knight), &pad.accum)
	groupByBoard(fBishop, pos.ByPiece(us, Bishop), &pad.accum)
	groupByBoard(fRook, pos.ByPiece(us, Rook), &pad.accum)
	groupByBoard(fQueen, pos.ByPiece(us, Queen), &pad.accum)
	groupByBoard(fKing, BbEmpty, &pad.accum)
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

func init() {
	// Initialize weights.
	initWeights()

	// Initialize futility figure bonus.
	for i, w := range wFigure {
		futilityFigureBonus[i] = scaleToCentipawns(max(w.M, w.E))
	}
}
