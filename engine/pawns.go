// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// pawns.go contains various utilities for handling pawns.
// TODO: Add connected pawns.

package engine

// PassedPawns returns all passed pawns of us given our and their pawns.
func PassedPawns(us Color, ours, theirs Bitboard) Bitboard {
	// From white's POV: w - white pawn, b - black pawn, x - non-passed pawns.
	// ........
	// ........
	// .....w..
	// .....x..
	// ..b..x..
	// .xxx.x..
	// .xxx.x..
	// .xxx.x..
	// .xxx.x..
	theirs |= East(theirs) | West(theirs)
	block := BackwardSpan(us, theirs|ours)
	return ours &^ block
}

// IsolatedPawns returns the isolated pawns on bb, i.e. pawns that do
// not have other pawns on adjacent files.
func IsolatedPawns(bb Bitboard) Bitboard {
	wings := East(bb) | West(bb)
	return bb &^ Fill(wings)
}

// DoubledPawns returns the doubled pawns on bb, i.e. pawns that
// have a friendly pawn directly in front of them.
func DoubledPawns(us Color, ours Bitboard) Bitboard {
	return ours & Backward(us, ours)
}

// PawnThreats returns the squares threatened by our pawns.
func PawnThreats(us Color, ours Bitboard) Bitboard {
	return Forward(us, East(ours)|West(ours))
}

// BackwardPawns returns the our backward pawns.
// A backward pawn is a pawn that has no pawns behind them on its file or
// adjacent file, it's not isolated and cannot advance safely.
func BackwardPawns(us Color, ours Bitboard, theirs Bitboard) Bitboard {
	behind := ForwardFill(us, East(ours)|West(ours))
	doubled := BackwardSpan(us, ours)
	isolated := IsolatedPawns(ours)
	return ours & Backward(us, PawnThreats(us.Opposite(), theirs)) &^ behind &^ doubled &^ isolated
}
