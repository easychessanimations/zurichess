// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"testing"
)

func TestScoreRange(t *testing.T) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		score := Evaluate(pos)
		if KnownLossScore >= score || score >= KnownWinScore {
			t.Errorf("expected %d in interval (%d, %d) for %s",
				score, KnownLossScore, KnownWinScore, fen)
		}
	}
}

func BenchmarkScore(b *testing.B) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		for i := 0; i < b.N; i++ {
			Evaluate(pos)
		}
	}
}
