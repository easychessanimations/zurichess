// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

import "sync"

type featureType string

const (
	// Value of each figure.
	fNoFigure featureType = "NoFigure"
	fPawn     featureType = "Pawn"
	fKnight   featureType = "Knight"
	fBishop   featureType = "Bishop"
	fRook     featureType = "Rook"
	fQueen    featureType = "Queen"
	fKing     featureType = "King"

	// PSqT for each figure from white's POV.
	// For pawns evaluate each square, but other figures
	// we only evaluate the coordinates.
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

	// Mobility of each figure.
	fKnightAttack featureType = "KnightAttack"
	fBishopAttack featureType = "BishopAttack"
	fRookAttack   featureType = "RookAttack"
	fQueenAttack  featureType = "QueenAttack"
	fKingAttack   featureType = "KingAttack"

	// Pawn structure.
	fBackwardPawns  featureType = "BackwardPawns"
	fConnectedPawns featureType = "ConnectedPawns"
	fDoubledPawns   featureType = "DoubledPawns"
	fIsolatedPawns  featureType = "IsolatedPawns"
	fRammedPawns    featureType = "RammedPawns"
	fPassedPawnRank featureType = "PassedPawnRank"
	fPawnMobility   featureType = "PawnMobility"
	// Minors and majors attacked by pawns.
	fMinorsPawnsAttack featureType = "MinorsPawnsAttack"
	fMajorsPawnsAttack featureType = "MajorsPawnsAttack"
	// Minors and majors attacked after a pawn push.
	fMinorsPawnsPotentialAttack featureType = "MinorsPawnsPotentialAttack"
	fMajorsPawnsPotentialAttack featureType = "MajorsPawnsPotentialAttack"
	// How close is the king from a friendly passed pawn.
	fKingPassedPawnTropism featureType = "KingPassedPawnTropism"
	// How close is the king from an enemy passed pawn.
	fKingEnemyPassedPawnTropism featureType = "KingEnemyPassedPawnTropism"

	// Attacked minors.
	fAttackedMinors featureType = "AttackedMinors"
	// Bishop pair.
	fBishopPair featureType = "BishopPair"
	// Rook on open and semi-open files.
	fRookOnOpenFile     featureType = "RookOnOpenFile"
	fRookOnSemiOpenFile featureType = "RookOnSemiOpenFile"
	fKingQueenTropism   featureType = "KingQueenTropism"

	// King safety.
	fKingAttackers featureType = "KingAttackers"
	// Pawn in king's area.
	fKingShelterNear featureType = "KingShelterNear"
	// Pawn in front of the king, on the same file.
	fKingShelterFront featureType = "KingShelterFront"
	// Pawn in front of the king, including adjacent files.
	fKingShelterFar featureType = "KingShelterFar"
)

var (
	// Placeholder for the weights array when running in coach mode.
	Weights []Score

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
