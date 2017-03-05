// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was generated by bitbucket.org/zurichess/tuner/extract

// +build !coach

package engine

type featureType int

const (
	fNoFigure                   featureType = 0
	fPawn                       featureType = 1
	fKnight                     featureType = 2
	fBishop                     featureType = 3
	fRook                       featureType = 4
	fQueen                      featureType = 5
	fKing                       featureType = 6
	fPawnMobility               featureType = 7
	fMinorsPawnsAttack          featureType = 8
	fMajorsPawnsAttack          featureType = 9
	fMinorsPawnsPotentialAttack featureType = 10
	fMajorsPawnsPotentialAttack featureType = 11
	fKnightFile                 featureType = 12
	fKnightRank                 featureType = 20
	fKnightAttack               featureType = 28
	fBishopFile                 featureType = 29
	fBishopRank                 featureType = 37
	fBishopAttack               featureType = 45
	fRookFile                   featureType = 46
	fRookRank                   featureType = 54
	fRookAttack                 featureType = 62
	fRookOnOpenFile             featureType = 63
	fRookOnSemiOpenFile         featureType = 64
	fQueenFile                  featureType = 65
	fQueenRank                  featureType = 73
	fQueenAttack                featureType = 81
	fKingAttackers              featureType = 82
	fPawnSquare                 featureType = 86
	fBackwardPawns              featureType = 150
	fConnectedPawns             featureType = 151
	fDoubledPawns               featureType = 152
	fIsolatedPawns              featureType = 153
	fKingFile                   featureType = 154
	fKingRank                   featureType = 162
	fKingAttack                 featureType = 170
	fKingShelterNear            featureType = 171
	fKingShelterFar             featureType = 172
	fKingShelterFront           featureType = 173
	fKingPassedPawnTropism      featureType = 174
	fPassedPawnRank             featureType = 182
)

func getFeatureStart(feature featureType, num int) int {
	return int(feature)
}
