package engine

import (
	"strings"
	"testing"
)

var (
	// Some games played by zurichess against itself or other engines.
	games = []string{
		"g2g3 b8c6 c2c3 d7d5 f1h3 e7e6 d1b3 f8d6 f2f4 g8f6 b3d1 e8g8 d1c2 c8d7 b2b3 f6e4 h3g4 d6f4 g3f4 d8h4 e1d1 e4f2 d1e1 f2d3 e1f1 h4f2",
		"h2h3 e7e5 g1f3 b8c6 f3g1 d7d5 h1h2 f8b4 c2c3 b4f8 h3h4 f8d6 g2g3 g8f6 f1h3 c8e6 h3e6 f7e6 d1a4 h8f8 h4h5 a7a6 h5h6 g7g6 d2d3 b7b5 a4h4 a6a5 c1g5 d5d4 g5f6 d8f6 h4f6 f8f6 b1d2 d4c3 b2c3 g6g5 d2e4 f6f5 g3g4 f5f7 e4g5 e5e4 h2g2 d6e5 d3d4 e5f6 g5f7 e8f7 a2a4 b5a4 a1a4 c6e7 g4g5 f6h8 g2g4 c7c5 g4e4 e7c6 d4c5 h8e5 e4e3 e5c7 a4a2 c7f4 e3f3 e6e5 e2e3 c6e7 e3f4 e7d5 f3d3 d5f4 d3d7 f7g8 d7g7 g8h8 c5c6 a8d8 c6c7 d8c8 g7d7 c8g8 a2a5 f4g2 e1d2 g2f4 d7d8 f4g6 c7c8q g6f8 d8f8 e5e4 f8g8",
		"e2e4 d7d5 e4e5 f7f6 d2d4 e7e6 f1b5 b8c6 e5f6 d8f6 b5c6 b7c6 h2h3 c6c5 d4c5 f6g6 d1d3 g6g2 d3f3 g2f3 g1f3 f8c5 c1e3 c5e7 e3d4 g8f6 f3e5 a8b8 e5c6 b8a8 c6e7 e8e7 d4f6 e7f6 b1c3 c8b7 c3b5 c7c6 b5c3 c6c5 c3a4 h8c8 h3h4 d5d4 h1h3 b7e4 c2c4 f6e5 h3g3 g7g6 e1c1 a8b8 g3g4 e4f5 f2f4 e5e4 g4g5 b8b7 d1h1 b7b4 h1e1 e4d3 b2b3 f5g4 g5g4 e6e5 g4g3",
		"d2d4 d7d5 c1f4 c8f5 g1f3 g8f6 e2e3 d8b6 b2b3 e7e6 f1d3 f8b4 c2c3 f5d3 d1d3 b4e7 b1d2 e8f8 h4h5 h7h6 e1f1 f8g8 f1g1 b8d7 f3e5 d7e5 f4e5 f6d7 e5g3 d7f6 c3c4 e7b4 c4c5 b6a7 g3c7 a8c8 c7e5 f6d7 d2f3 b7b6 c5b6 a7b6 a1f1 c6c5 f1c1 b4a3 c1c2 d7e5 f3e5 b6d6 f2f4 d6e7 e3e4 c8c7 f4f5 d5e4 d3e4 e6f5 e4a8 g8h7 a8a5 c7b7 a5c3 c5d4 c3d4 h8e8 e5f3 b7b3 f3d2 b3b1 c2c1 b1c1 d2f1 c1f1 g1h2 a3d6 g2g3 e7e2 d4f2 e2f2 h2h3 f1h1",
		"f1b5 c8d7 b1c3 d7b5 c3b5 f8g7 g1h3 c7c6 b5a3 g8f6 e1g1 e8g8 d2d4 c6c5 c1d2 c5d4 e3d4 b8c6 c2c3 d8b6 b2b3 b6a6 a3c2 e7e6 h3f2 a8c8 d1f3 a6a5 a2a4 a7a6 f3d3 a5f5 a4a5 f5d3 f2d3 f6e4 d2e1 c6d4 c2d4 e4c3 d4f3 c3e2 g1f2 g7a1 f2e2 a1c3 e1c3 c8c3 f3d2 f8c8 g2g4 c3c2 f1a1 h7h5 g4h5 g6h5 a1g1 g8h7 e2e3 c8g8 g1d1 h7g7 d3b4 c2c3 b4d3 h5h4 e3d4 c3c2 d3b4 c2c5 d2c4 g7f6 c4d6 c5a5 h2h3 g8d8 d4c3 a5c5 c3b2 c5c7 d6e4 f6e7 d1d8 e7d8 b4d3 d8e7 d3e5 a6a5 e4c3 c7c8 c3e4 c8d8 e4c5 d8b8 e5c4 b7b6 c5d3 e7f6 b2c3 f6g6 c3d4 g6f5 c4e3 f5f6 d3e5 f6g7 e3c4 b8d8 d4c3 d8d1 c4b6 d1h1 f4f5 h1h3 c3c2 e6f5 b6c4 h3g3 c4d6 g7f6 e5c4 h4h3 c4d2 g3g2 d6e8 f6g6 e8c7 h3h2 c7d5 g2f2 d5b6 h2h1Q b6c4 g6f6 c2d3",
		"e2e4 d7d5 f1e2 e7e5 e1f1 g4h6 e4d5 e5f4 e2f3 f8d6 d1e2 e8f8 b1c3 f8g8 d2d4 c8f5 g2g3 f4g3 h2g3 b8d7 g3g4 f5g6 c1h6 g7h6 h1h3 d8f6 e2f2 a8e8 c3b5 d7b6 f1g2 f6g5 b5d6 c7d6 g2h2 b6d5 h3g3 g5e3 g1h3 e3f2 h3f2 d5b6 f3b7 e8e2 h2g1 e2c2 b2b3 c2d2 a1d1 d2b2 g3e3 g6c2 e3e8 g8g7 e8e2 b2a2 b3b4 h8b8 b7c6 c2d1 e2a2 b8c8 b4b5 d1b3 a2a1 c8g8 g1g2 b6c4 g2f3 b3a4 f2e4 a7a6 e4c3 a6b5 c6b5 a4b5 c3b5 g7g6 f3f4 g8e8 a3a4 e8e2 b5c3 e2f2 f4e4 f2g2 e4f3 g2h2 c3d5 h2h3 f3e2 h3g3 a4a5 c4a5 a1a5 g3g4 e2d3 g4g5 d3e4 g6g7 a5a7 h6h5 d5e3 g7f6 e3d5 f6g7 d5e3 g7f6 a7a8 h5h4 e4f4 h7h6 a8a6 h4h3 a6a1 g5h5 f4g3 f6g6 a1h1 g6g5 h1f1 g5g6 g3h2 f7f6 f1f3 f6f5 e3c4 g6f6 d4d5 f5f4 f3f4 h5f5 f4f5 f6f5 h2h3 f5e4 c4b6 e4d4 h3g4 d4c5 b6d7 c5d5 g4f5 h6h5 f5g5 d5c4 d7b6 c4c5 b6a4 c5d4 g5h5 d6d5 h5g5 d4e3 a4c3 d5d4 c3b5 d4d3 b5a3 d3d2 a3c4 e3e4 c4d2",
	}
)

func TestGame(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, Options{})
	for i := 0; i < 1; i++ {
		tc := NewFixedDepthTimeControl(pos, 3)
		tc.Start(false)
		move := eng.Play(tc)
		eng.DoMove(move[0])
	}
}

// Test score is the same if we start with the position or move.
func TestScore(t *testing.T) {
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, Options{})
		static := NewEngine(pos, Options{})

		moves := strings.Fields(game)
		for _, move := range moves {
			m := pos.UCIToMove(move)
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
	eng := NewEngine(pos, Options{})
	moves := eng.Play(tc)
	if 0 != len(moves) {
		t.Errorf("expected no pv, got %d moves", len(moves))
	}
}

func BenchmarkStallingFENs(b *testing.B) {
	fens := []string{
		// Causes quiscence search to explode.
		"rnb1kbnr/pppp1ppp/8/8/3PPp1q/6P1/PPP4P/RNBQKBNR b KQkq -1 0 4",
		"r2qr1k1/2pn1ppp/pp2pn2/3b4/3P4/B2BPN2/P1P1QPPP/R4RK1 w - -1 4 13",
		"r1bq2k1/ppp4p/2n5/2bpPr2/5pQ1/2P5/PP4PP/RNB1NR1K b - -1 4 15",
	}

	for i := 0; i < b.N; i++ {
		for _, fen := range fens {
			pos, _ := PositionFromFEN(fen)
			eng := NewEngine(pos, Options{})
			tc := NewFixedDepthTimeControl(pos, 5)
			tc.Start(false)
			eng.Play(tc)
		}
	}
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		eng := NewEngine(pos, Options{})
		for j := 0; j < 20; j++ {
			tc := NewFixedDepthTimeControl(pos, 4)
			tc.Start(false)
			move := eng.Play(tc)
			eng.DoMove(move[0])
		}
	}
}

func BenchmarkScore(b *testing.B) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, Options{})

	for i := 0; i < b.N; i++ {
		for _, g := range games {
			var done []Move
			todo := strings.Fields(g)

			for j := range todo {
				move := eng.Position.UCIToMove(todo[j])
				done = append(done, move)
				eng.DoMove(move)
				_ = eng.Score()
			}

			for j := range done {
				move := done[len(done)-1-j]
				eng.UndoMove(move)
				_ = eng.Score()
			}
		}
	}
}
