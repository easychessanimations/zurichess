// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

import "sync"

const (
	// Figure.
	fNoFigure string = "NoFigure"
	fPawn     string = "Pawn"
	fKnight   string = "Knight"
	fBishop   string = "Bishop"
	fRook     string = "Rook"
	fQueen    string = "Queen"
	fKing     string = "King"

	// PSqT
	fPawnSquare string = "PawnSquare"
	fKnightFile string = "KnightFile"
	fKnightRank string = "KnightRank"
	fBishopFile string = "BishopFile"
	fBishopRank string = "BishopRank"
	fRookFile   string = "RookFile"
	fRookRank   string = "RookRank"
	fQueenFile  string = "QueenFile"
	fQueenRank  string = "QueenRank"
	fKingFile   string = "KingFile"
	fKingRank   string = "KingRank"

	// Mobility.
	fKnightAttack string = "KnightAttack"
	fBishopAttack string = "BishopAttack"
	fRookAttack   string = "RookAttack"
	fQueenAttack  string = "QueenAttack"
	fKingAttack   string = "KingAttack"

	// Pawn structur.
	fBackwardPawns     string = "BackwardPawns"
	fConnectedPawns    string = "ConnectedPawns"
	fDoubledPawns      string = "DoubledPawns"
	fIsolatedPawns     string = "IsolatedPawns"
	fPassedPawnRank    string = "PassedPawnRank"
	fPawnMobility      string = "PawnMobility"
	fMinorsPawnsAttack string = "MinorsPawnsAttack"
	fMajorsPawnsAttack string = "MajorsPawnsAttack"

	// Other stuff.
	fRookOnOpenFile     string = "RookOnOpenFile"
	fRookOnSemiOpenFile string = "RookOnSemiOpenFile"

	// Shelter
	fKingShelter string = "KingShelter"
)

var (
	FeaturesMap     = make(map[string]*FeatureInfo)
	featuresMapLock sync.Mutex
)

type FeatureInfo struct {
	Name  string // Name of this feature.
	Start int    // Start position in the weights array.
	Num   int    // Number of weights used.
}

func getFeatureStart(feature string, num int) int {
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
