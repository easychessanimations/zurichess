// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// material.go implements position evaluation.
//
// Zurichess' evaluation is a simple neural network with no hidden layers,
// and one output node y = W_m * x * (1-p) + W_e * x * p where W_m are
// middle game weights, W_e are endgame weights, x is input, p is phase between
// middle game and end game, and y is the score.
// The network has |x| = len(Weights) inputs corresponding to features
// extracted from the position. These features are symmetrical wrt colors.
// The network is trained using the Texel's Tuning Method
// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.
// Tuning is done by bitbucket.org/zurichess/tuner tool which uses
// tensorflow.org machine learning framework.

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
	// Evaluation caches.
	pawnsAndShelterCache = &pawnsTable{}

	// Figure bonuses to use when computing the futility margin.
	futilityFigureBonus [FigureArraySize]int32
)

// Eval contains necessary information for evaluation.
type Eval struct {
	Accum    Accum
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
	w := evaluate(pos, White)
	b := evaluate(pos, Black)

	wps, bps := pawnsAndShelterCache.load(pos)
	w.merge(wps)
	b.merge(bps)

	e := Eval{}
	e.Accum.merge(w)
	e.Accum.deduct(b)
	return e
}

func evaluatePawnsAndShelter(pos *Position, us Color) (accum Accum) {
	evaluatePawns(pos, us, &accum)
	evaluateShelter(pos, us, &accum)
	return accum
}

func evaluatePawns(pos *Position, us Color, accum *Accum) {
	groupBySquare(fPawnSquare, us, pos.ByPiece(us, Pawn), accum)
	groupByBoard(fBackwardPawns, BackwardPawns(pos, us), accum)
	groupByBoard(fConnectedPawns, ConnectedPawns(pos, us), accum)
	groupByBoard(fDoubledPawns, DoubledPawns(pos, us), accum)
	groupByBoard(fIsolatedPawns, IsolatedPawns(pos, us), accum)
	groupByRank(fPassedPawnRank, us, PassedPawns(pos, us), accum)
}

func evaluateShelter(pos *Position, us Color, accum *Accum) {
	// King's position and mobility.
	bb := pos.ByPiece(us, King)
	kingSq := bb.AsSquare()
	mobility := KingMobility(kingSq)
	groupByFileSq(fKingFile, us, kingSq, accum)
	groupByRankSq(fKingRank, us, kingSq, accum)
	groupByBoard(fKingAttack, mobility, accum)

	// King's shelter.
	ekw := East(bb) | bb | West(bb)
	ourPawns := pos.ByPiece(us, Pawn)
	groupByBoard(fKingShelterNear, ekw|Forward(us, ekw)&ourPawns, accum)
	groupByBoard(fKingShelterFar, ForwardSpan(us, ekw)&ourPawns, accum)
	groupByBoard(fKingShelterFront, ForwardSpan(us, bb)&ourPawns, accum)

	// King passed pawn tropism.
	dist := int32(8)
	for bb := PassedPawns(pos, us); bb != BbEmpty; {
		sq := bb.Pop()
		if sq.POV(us).Rank() >= 4 {
			dist = min(dist, distance[sq][kingSq])
		}
	}
	groupByBucket(fKingPassedPawnTropism, int(dist), 8, accum)

	dist = 8
	for bb := PassedPawns(pos, us.Opposite()); bb != BbEmpty; {
		sq := bb.Pop()
		if sq.POV(us.Opposite()).Rank() >= 4 {
			dist = min(dist, distance[sq][kingSq])
		}
	}
	groupByBucket(fKingEnemyPassedPawnTropism, int(dist), 8, accum)
}

// evaluate evaluates position for a single side.
func evaluate(pos *Position, us Color) Accum {
	var accum Accum
	them := us.Opposite()
	all := pos.ByColor(White) | pos.ByColor(Black)
	danger := PawnThreats(pos, them)
	ourPawns := pos.ByPiece(us, Pawn)
	theirPawns := pos.ByPiece(them, Pawn)
	theirKingArea := KingArea(pos, them)

	groupByBoard(fNoFigure, BbEmpty, &accum)
	groupByBoard(fPawn, pos.ByPiece(us, Pawn), &accum)
	groupByBoard(fKnight, pos.ByPiece(us, Knight), &accum)
	groupByBoard(fBishop, pos.ByPiece(us, Bishop), &accum)
	groupByBoard(fRook, pos.ByPiece(us, Rook), &accum)
	groupByBoard(fQueen, pos.ByPiece(us, Queen), &accum)
	groupByBoard(fKing, BbEmpty, &accum)

	// Evaluate various pawn attacks and potential pawn attacks
	// on the enemy pieces.
	groupByBoard(fPawnMobility, ourPawns&^Backward(us, all), &accum)
	groupByBoard(fMinorsPawnsAttack, Minors(pos, us)&danger, &accum)
	groupByBoard(fMajorsPawnsAttack, Majors(pos, us)&danger, &accum)
	groupByBoard(fMinorsPawnsPotentialAttack, Minors(pos, us)&Backward(us, danger), &accum)
	groupByBoard(fMajorsPawnsPotentialAttack, Majors(pos, us)&Backward(us, danger), &accum)

	numAttackers := 0
	attacks := PawnThreats(pos, us)

	// Knight
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		mobility := KnightMobility(sq) &^ (danger | ourPawns)
		attacks |= mobility
		groupByFileSq(fKnightFile, us, sq, &accum)
		groupByRankSq(fKnightRank, us, sq, &accum)
		groupByBoard(fKnightAttack, mobility, &accum)
		if mobility&theirKingArea&^theirPawns != 0 {
			numAttackers++
		}
	}
	// Bishop
	// TODO Fix bishop's attack.
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		mobility := BishopMobility(sq, all)
		attacks |= mobility
		mobility &^= danger | ourPawns
		groupByFileSq(fBishopFile, us, sq, &accum)
		groupByRankSq(fBishopRank, us, sq, &accum)
		groupByBoard(fBishopAttack, mobility, &accum)
		if mobility&theirKingArea&^theirPawns != 0 {
			numAttackers++
		}
	}
	// Rook
	openFiles := OpenFiles(pos, us)
	semiOpenFiles := SemiOpenFiles(pos, us)
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		mobility := RookMobility(sq, all) &^ (danger | ourPawns)
		attacks |= mobility
		groupByFileSq(fRookFile, us, sq, &accum)
		groupByRankSq(fRookRank, us, sq, &accum)
		groupByBoard(fRookAttack, mobility, &accum)
		groupByBool(fRookOnOpenFile, openFiles.Has(sq), &accum)
		groupByBool(fRookOnSemiOpenFile, semiOpenFiles.Has(sq), &accum)
		if mobility&theirKingArea&^theirPawns != 0 {
			numAttackers++
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		mobility := QueenMobility(sq, all) &^ (danger | ourPawns)
		attacks |= mobility
		groupByFileSq(fQueenFile, us, sq, &accum)
		groupByRankSq(fQueenRank, us, sq, &accum)
		groupByBoard(fQueenAttack, mobility, &accum)
		if mobility&theirKingArea&^theirPawns != 0 {
			numAttackers++
		}

		dist := distance[sq][pos.ByPiece(them, King).AsSquare()]
		groupByCount(fKingQueenTropism, dist, &accum)
	}

	groupByBoard(fAttackedMinors, attacks&Minors(pos, them), &accum)
	groupByBool(fBishopPair, pos.ByPiece(us, Bishop).CountMax2() == 2, &accum)

	// Kink's safety is very primitive:
	// - king's shelter is evaluated by evaluateShelter.
	// - the following counts the number of attackers.
	// TODO: Queen tropism which was dropped during the last refactoring.
	groupByBucket(fKingAttackers, numAttackers, 4, &accum)
	return accum
}

// Phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func Phase(pos *Position) int32 {
	total := int32(4*1 + 4*1 + 4*3 + 2*6)
	curr := total
	curr -= pos.ByFigure(Knight).Count() * 1
	curr -= pos.ByFigure(Bishop).Count() * 1
	curr -= pos.ByFigure(Rook).Count() * 3
	curr -= pos.ByFigure(Queen).Count() * 6
	curr = max(curr, 0)
	return (curr*256 + total/2) / total
}

// scaleToCentipawns scales a score in the original scale to centipawns.
func scaleToCentipawns(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + 64 + score>>31) >> 7
}

func init() {
	// Initialize futility figure bonus.
	if len(Weights) != 0 {
		for i, w := range Weights[:FigureArraySize] {
			futilityFigureBonus[i] = scaleToCentipawns(max(w.M, w.E))
		}
	}
}
