package engine

import (
	"math"
	"testing"
)

func TestStack(t *testing.T) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		st := &stack{} // no killer, no hash
		st.Reset(pos)
		st.GenerateMoves(All, NullMove)

		t.Log("fen ", fen)
		limit := int16(math.MaxInt16)
		for move := st.PopMove(); move != NullMove; move = st.PopMove() {
			t.Log(move, " ", mvvlva(move))
			if curr := mvvlva(move); curr > limit {
				t.Errorf("moves not sorted")
			} else {
				limit = curr
			}
		}
	}
}

func TestReturnsHashMove(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)

	for i, str := range []string{"f3f5", "e2b5", "a1b1"} {
		hash := pos.UCIToMove(str)
		st := &stack{}
		st.Reset(pos)
		st.GenerateMoves(All, hash)
		if move := st.PopMove(); hash != move {
			t.Errorf("#%d expected move %v, got %v", i, hash, move)
		}
	}
}

func TestReturnsMoves(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	seen := make(map[Move]int)

	var moves []Move
	pos.GenerateMoves(All, &moves)
	for _, m := range moves {
		seen[m] |= 1
	}

	st := &stack{}
	st.Reset(pos)
	st.GenerateMoves(All, NullMove)
	for m := st.PopMove(); m != NullMove; m = st.PopMove() {
		if seen[m]&2 != 0 {
			t.Errorf("move %v not expected", m)
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
