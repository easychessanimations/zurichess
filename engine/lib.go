// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !coach

package engine

import (
	. "bitbucket.org/zurichess/zurichess/board"
)

func groupByBoard(feature int, bb Bitboard, accum *Accum) {
	accum.addN(Weights[feature], bb.Count())
}

func groupBySquare(feature int, bb Bitboard, accum *Accum) {
	for bb != BbEmpty {
		sq := bb.Pop()
		accum.addN(Weights[feature+int(sq)], bb.Count())
	}
}

func groupByBool(feature int, b bool, accum *Accum) {
	if b {
		accum.addN(Weights[feature], 1)
	} else {
		accum.addN(Weights[feature], 0)
	}
}

func groupByFileSq(feature int, sq Square, accum *Accum) {
	accum.add(Weights[feature+sq.File()])
}

func groupByRankSq(feature int, sq Square, accum *Accum) {
	accum.add(Weights[feature+sq.Rank()])
}

func groupByRank(feature int, bb Bitboard, accum *Accum) {
	for bb != BbEmpty {
		sq := bb.Pop()
		groupByRankSq(feature, sq, accum)
	}
}
