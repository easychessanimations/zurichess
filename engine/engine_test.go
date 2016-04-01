package engine

import (
	"strings"
	"testing"
)

func TestGame(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, nil, Options{})
	for i := 0; i < 1; i++ {
		tc := NewFixedDepthTimeControl(pos, 3)
		tc.Start(false)
		move := eng.Play(tc)
		eng.DoMove(move[0])
	}
}

func TestMateIn1(t *testing.T) {
	for i, d := range mateIn1 {
		pos, _ := PositionFromFEN(d.fen)
		bm, err := pos.UCIToMove(d.bm)
		if err != nil {
			t.Errorf("#%d cannot parse move %s", i, d.bm)
			continue
		}

		tc := NewFixedDepthTimeControl(pos, 2)
		tc.Start(false)
		eng := NewEngine(pos, nil, Options{})
		pv := eng.Play(tc)

		if len(pv) != 1 {
			t.Errorf("#%d Expected at most one move, got %d", i, len(pv))
			t.Errorf("position is %v", pos)
			continue
		}

		if pv[0] != bm {
			t.Errorf("#%d expected move %v, got %v", i, bm, pv[0])
			t.Errorf("position is %v", pos)
			continue
		}
	}
}

// Test score is the same if we start with the position or move.
func TestScore(t *testing.T) {
	for _, game := range testGames {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, nil, Options{})
		static := NewEngine(pos, nil, Options{})

		moves := strings.Fields(game)
		for _, move := range moves {
			m, _ := pos.UCIToMove(move)
			if !pos.IsPseudoLegal(m) {
				// t.Fatalf("bad bad bad")
			}

			dynamic.DoMove(m)
			static.SetPosition(pos)
			if dynamic.Score() != static.Score() {
				t.Fatalf("expected static score %v, got dynamic score %v", static.Score(), dynamic.Score())
			}
		}
	}
}

func TestEndGamePosition(t *testing.T) {
	pos, _ := PositionFromFEN("6k1/5p1p/4p1p1/3p4/5P1P/8/3r2q1/6K1 w - - 2 55")
	tc := NewFixedDepthTimeControl(pos, 3)
	tc.Start(false)
	eng := NewEngine(pos, nil, Options{})
	moves := eng.Play(tc)
	if 0 != len(moves) {
		t.Errorf("expected no pv, got %d moves", len(moves))
	}
}

func TestPassed(t *testing.T) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		var moves []Move
		pos.GenerateMoves(All, &moves)
		before := passedPawns(pos, White) | passedPawns(pos, Black)

		for _, m := range moves {
			pos.DoMove(m)
			after := passedPawns(pos, White) | passedPawns(pos, Black)
			if passed(pos, m) && before == after {
				t.Errorf("expected no passed pawn, got passed pawn: move = %v, position = %v", m, pos)
			}

			pos.UndoMove()
			if passed(pos, m) && before == after {
				t.Errorf("expected no passed pawn, got passed pawn: move = %v, position = %v", m, pos)
			}
		}
	}
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		eng := NewEngine(pos, nil, Options{})
		for j := 0; j < 20; j++ {
			tc := NewFixedDepthTimeControl(pos, 4)
			tc.Start(false)
			move := eng.Play(tc)
			eng.DoMove(move[0])
		}
	}
}
