// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

import (
	. "bitbucket.org/zurichess/zurichess/board"
)

func groupByBoard(feature string, bb Bitboard, accum *Accum) {
	start := getFeatureStart(feature, 1)
	accum.addN(Weights[start], bb.Count())
}

func groupBySquare(feature string, us Color, bb Bitboard, accum *Accum) {
	start := getFeatureStart(feature, 64)
	for bb != BbEmpty {
		sq := bb.Pop().POV(us)
		accum.add(Weights[start+int(sq)])
	}
}

func groupByBool(feature string, b bool, accum *Accum) {
	start := getFeatureStart(feature, 1)
	if b {
		accum.addN(Weights[start], 1)
	} else {
		accum.addN(Weights[start], 0)
	}
}

func groupByFileSq(feature string, us Color, sq Square, accum *Accum) {
	start := getFeatureStart(feature, 8)
	accum.add(Weights[start+sq.POV(us).File()])
}

func groupByRankSq(feature string, us Color, sq Square, accum *Accum) {
	start := getFeatureStart(feature, 8)
	accum.add(Weights[start+sq.POV(us).Rank()])
}

func groupByRank(feature string, us Color, bb Bitboard, accum *Accum) {
	for bb != BbEmpty {
		sq := bb.Pop()
		groupByRankSq(feature, us, sq, accum)
	}
}
