// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// pawns.go contains various utilities for handling pawns.
// TODO: Add connected pawns.

package engine

var (
	murmurSeed = [ColorArraySize]uint64{
		0x77a166129ab66e91,
		0x4f4863d5038ea3a3,
		0xe14ec7e648a4068b,
	}
)

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

// pawnsTable is a cache entry.
type pawnsEntry struct {
	lock uint64
	eval Eval
}

// pawnsTable implements a fixed size cache.
type pawnsTable [1 << 9]pawnsEntry

// put puts a new entry in the cache.
func (c *pawnsTable) put(lock uint64, eval Eval) {
	indx := lock & uint64(len(*c)-1)
	(*c)[indx] = pawnsEntry{lock: lock, eval: eval}
}

// get gets an entry from the cache.
func (c *pawnsTable) get(lock uint64) (Eval, bool) {
	indx := lock & uint64(len(*c)-1)
	return (*c)[indx].eval, (*c)[indx].lock == lock
}

// load evaluates position, using the cache if possible.
func (c *pawnsTable) load(pos *Position, us Color) Eval {
	if disableCache {
		return evaluatePawnsAndShelter(pos, us)
	}
	h := pawnsHash(pos, us)
	if e, ok := c.get(h); ok {
		return e
	}
	e := evaluatePawnsAndShelter(pos, us)
	c.put(h, e)
	return e
}

// pawnsHash returns a hash of the pawns and king in position.
func pawnsHash(pos *Position, us Color) uint64 {
	h := murmurSeed[us]
	h = murmurMix(h, uint64(pos.ByPiece(us, Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us.Opposite(), Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us, King)))
	return h
}

// murmuxMix function mixes two integers k&h.
//
// murmurMix is based on MurmurHash2 https://sites.google.com/site/murmurhash/ which is on public domain.
func murmurMix(k, h uint64) uint64 {
	h ^= k
	h *= uint64(0xc6a4a7935bd1e995)
	return h ^ (h >> uint(51))
}
