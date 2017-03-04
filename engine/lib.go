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
