// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import . "bitbucket.org/zurichess/zurichess/board"

// distance stores the number of king steps required
// to reach from one square to another on an empty board.
var distance [SquareArraySize][SquareArraySize]int32

var murmurSeed = [ColorArraySize]uint64{
	0x77a166129ab66e91,
	0x4f4863d5038ea3a3,
	0xe14ec7e648a4068b,
}

// max returns maximum of a and b.
func max(a, b int32) int32 {
	if a >= b {
		return a
	}
	return b
}

// min returns minimum of a and b.
func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
}

// murmuxMix function mixes two integers k&h.
//
// murmurMix is based on MurmurHash2 https://sites.google.com/site/murmurhash/ which is on public domain.
func murmurMix(k, h uint64) uint64 {
	h ^= k
	h *= uint64(0xc6a4a7935bd1e995)
	return h ^ (h >> uint(51))
}

func init() {
	for i := SquareMinValue; i <= SquareMaxValue; i++ {
		for j := SquareMinValue; j <= SquareMaxValue; j++ {
			f, r := int32(i.File()-j.File()), int32(i.Rank()-j.Rank())
			f, r = max(f, -f), max(r, -r) // absolute value
			distance[i][j] = max(f, r)
		}
	}
}
