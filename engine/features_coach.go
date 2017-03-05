// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

import "sync"

type featureType string

const (
	// Figure.
	fNoFigure featureType = "NoFigure"
	fPawn     featureType = "Pawn"
	fKnight   featureType = "Knight"
	fBishop   featureType = "Bishop"
	fRook     featureType = "Rook"
	fQueen    featureType = "Queen"
	fKing     featureType = "King"

	// PSqT
	fPawnSquare featureType = "PawnSquare"
	fKnightFile featureType = "KnightFile"
	fKnightRank featureType = "KnightRank"
	fBishopFile featureType = "BishopFile"
	fBishopRank featureType = "BishopRank"
	fRookFile   featureType = "RookFile"
	fRookRank   featureType = "RookRank"
	fQueenFile  featureType = "QueenFile"
	fQueenRank  featureType = "QueenRank"
	fKingFile   featureType = "KingFile"
	fKingRank   featureType = "KingRank"

	// Mobility.
	fKnightAttack featureType = "KnightAttack"
	fBishopAttack featureType = "BishopAttack"
	fRookAttack   featureType = "RookAttack"
	fQueenAttack  featureType = "QueenAttack"
	fKingAttack   featureType = "KingAttack"

	// Pawn structur.
	fBackwardPawns     featureType = "BackwardPawns"
	fConnectedPawns    featureType = "ConnectedPawns"
	fDoubledPawns      featureType = "DoubledPawns"
	fIsolatedPawns     featureType = "IsolatedPawns"
	fPassedPawnRank    featureType = "PassedPawnRank"
	fPawnMobility      featureType = "PawnMobility"
	fMinorsPawnsAttack featureType = "MinorsPawnsAttack"
	fMajorsPawnsAttack featureType = "MajorsPawnsAttack"

	// Other stuff.
	fRookOnOpenFile     featureType = "RookOnOpenFile"
	fRookOnSemiOpenFile featureType = "RookOnSemiOpenFile"

	// Shelter
	fKingAttackers    featureType = "KingAttackers"
	fKingShelterNear  featureType = "KingShelterNear"
	fKingShelterFront featureType = "KingShelterFront"
	fKingShelterFar   featureType = "KingShelterFar"
)

var (
	FeaturesMap     = make(map[featureType]*FeatureInfo)
	featuresMapLock sync.Mutex
)

type FeatureInfo struct {
	Name  featureType // Name of this feature.
	Start int         // Start position in the weights array.
	Num   int         // Number of weights used.
}

func getFeatureStart(feature featureType, num int) int {
	featuresMapLock.Lock()
	defer featuresMapLock.Unlock()

	if info, has := FeaturesMap[feature]; has {
		return info.Start
	}
	FeaturesMap[feature] = &FeatureInfo{
		Name:  feature,
		Start: len(Weights),
		Num:   num,
	}
	for i := 0; i < num; i++ {
		Weights = append(Weights, Score{M: 0, E: 0, I: len(Weights)})
	}
	return FeaturesMap[feature].Start
}
