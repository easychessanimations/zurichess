// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

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

	// Pawn structre.
	fBackwardPawns  string = "BackwardPawns"
	fConnectedPawns string = "ConnectedPawns"
	fDoubledPawns   string = "DoubledPawns"
	fIsolatedPawns  string = "IsolatedPawns"
	fPassedPawnRank string = "PassedPawnRank"

	// Other stuff.
	fRookOnOpenFile     string = "RookOnOpenFile"
	fRookOnSemiOpenFile string = "RookOnSemiOpenFile"

	// Shelter
	fKingShelter string = "KingShelter"
)
