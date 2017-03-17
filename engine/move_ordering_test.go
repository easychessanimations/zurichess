// Copyright 2014-2016 The Zurichess Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"testing"

	. "bitbucket.org/zurichess/zurichess/board"
	. "bitbucket.org/zurichess/zurichess/internal/testdata"
)

func TestOrdersViolentMovesByMVVLVA(t *testing.T) {
	for _, fen := range TestFENs {
		pos, _ := PositionFromFEN(fen)
		st := &stack{
			history: new(historyTable),
			counter: new([1 << 11]Move),
		}
		st.Reset(pos)
		st.GenerateMoves(Violent, NullMove)

		limit := int16(0x7fff)
		for move := st.PopMove(); move != NullMove; move = st.PopMove() {
			if curr := mvvlva(move); curr > limit {
				t.Errorf("moves not sorted: %v", move)
			} else {
				limit = curr
			}
		}
	}
}

func TestReturnsHashMove(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)

	for i, str := range []string{"f3f5", "e2b5", "a1b1"} {
		hash, _ := pos.UCIToMove(str)
		st := &stack{}
		st.Reset(pos)
		st.GenerateMoves(Violent|Quiet, hash)
		if move := st.PopMove(); hash != move {
			t.Errorf("#%d expected move %v, got %v", i, hash, move)
		}
	}
}

func TestReturnsMoves(t *testing.T) {
	for _, fen := range TestFENs {
		pos, _ := PositionFromFEN(fen)
		seen := make(map[Move]int)

		var moves []Move
		pos.GenerateMoves(Violent|Quiet, &moves)
		for _, m := range moves {
			seen[m] |= 1
		}

		st := &stack{
			history: new(historyTable),
			counter: new([1 << 11]Move),
		}
		st.Reset(pos)
		st.GenerateMoves(Violent|Quiet, moves[1234567891%len(moves)])
		for m := st.PopMove(); m != NullMove; m = st.PopMove() {
			if seen[m]&2 != 0 {
				t.Errorf("move %v is duplicate", m)
			}
			seen[m] |= 2
		}

		for m, v := range seen {
			if v == 1 {
				t.Errorf("move %v not generated", m)
			}
			if v == 2 {
				t.Errorf("move %v not expected", m)
			}
		}
	}
}
