// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import . "bitbucket.org/zurichess/zurichess/board"

// distance stores the number of king steps required
// to reach from one square to another on an empty board.
var distance [SquareArraySize][SquareArraySize]int32

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

func init() {
	for i := SquareMinValue; i <= SquareMaxValue; i++ {
		for j := SquareMinValue; j <= SquareMaxValue; j++ {
			f, r := int32(i.File()-j.File()), int32(i.Rank()-j.Rank())
			f, r = max(f, -f), max(r, -r) // absolute value
			distance[i][j] = max(f, r)
		}
	}
}
