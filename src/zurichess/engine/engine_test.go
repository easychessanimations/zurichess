package engine

import (
	"strings"
	"testing"
)

var (
	// Some games between zurichess and stockfish.
	games = []string{
		"g2g3 b8c6 c2c3 d7d5 f1h3 e7e6 d1b3 f8d6 f2f4 g8f6 b3d1 e8g8 d1c2 c8d7 b2b3 f6e4 h3g4 d6f4 g3f4 d8h4 e1d1 e4f2 d1e1 f2d3 e1f1 h4f2",
		"h2h3 e7e5 g1f3 b8c6 f3g1 d7d5 h1h2 f8b4 c2c3 b4f8 h3h4 f8d6 g2g3 g8f6 f1h3 c8e6 h3e6 f7e6 d1a4 h8f8 h4h5 a7a6 h5h6 g7g6 d2d3 b7b5 a4h4 a6a5 c1g5 d5d4 g5f6 d8f6 h4f6 f8f6 b1d2 d4c3 b2c3 g6g5 d2e4 f6f5 g3g4 f5f7 e4g5 e5e4 h2g2 d6e5 d3d4 e5f6 g5f7 e8f7 a2a4 b5a4 a1a4 c6e7 g4g5 f6h8 g2g4 c7c5 g4e4 e7c6 d4c5 h8e5 e4e3 e5c7 a4a2 c7f4 e3f3 e6e5 e2e3 c6e7 e3f4 e7d5 f3d3 d5f4 d3d7 f7g8 d7g7 g8h8 c5c6 a8d8 c6c7 d8c8 g7d7 c8g8 a2a5 f4g2 e1d2 g2f4 d7d8 f4g6 c7c8q g6f8 d8f8 e5e4 f8g8",
		"e2e4 d7d5 e4e5 f7f6 d2d4 e7e6 f1b5 b8c6 e5f6 d8f6 b5c6 b7c6 h2h3 c6c5 d4c5 f6g6 d1d3 g6g2 d3f3 g2f3 g1f3 f8c5 c1e3 c5e7 e3d4 g8f6 f3e5 a8b8 e5c6 b8a8 c6e7 e8e7 d4f6 e7f6 b1c3 c8b7 c3b5 c7c6 b5c3 c6c5 c3a4 h8c8 h3h4 d5d4 h1h3 b7e4 c2c4 f6e5 h3g3 g7g6 e1c1 a8b8 g3g4 e4f5 f2f4 e5e4 g4g5 b8b7 d1h1 b7b4 h1e1 e4d3 b2b3 f5g4 g5g4 e6e5 g4g3",
	}
)

func TestGame(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, EngineOptions{})
	for i := 0; i < 1; i++ {
		tc := &FixedDepthTimeControl{MinDepth: 3, MaxDepth: 3}
		tc.Start()
		move, _ := eng.Play(tc)
		eng.DoMove(move)
	}
}

// Test score is the same if we start with the position or move.
func TestScore(t *testing.T) {
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, EngineOptions{})
		moves := strings.Fields(game)
		for _, move := range moves {
			m := pos.ParseMove(move)
			t.Log("move", m, "piece", m.Target, "capture", m.Capture)
			dynamic.DoMove(m)
			static := NewEngine(pos, EngineOptions{})
			if dynamic.Score() != static.Score() {
				t.Logf("expected static score %v, got dynamic score %v",
					static.Score(), dynamic.Score())
				t.Logf(" static pieces %v; pieceScore %v; positionScore %v",
					static.pieces, static.pieceScore, static.positionScore)
				t.Logf("dynamic pieces %v; pieceScore %v; positionScore %v",
					dynamic.pieces, dynamic.pieceScore, dynamic.positionScore)
				t.FailNow()
			}
		}
	}
}

// Test zobrist is the same if we start with the position or move.
func TestZobrist(t *testing.T) {
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, EngineOptions{})
		moves := strings.Fields(game)
		for _, move := range moves {
			m := pos.ParseMove(move)
			t.Log("move", m, "piece", m.Target, "capture", m.Capture)
			dynamic.DoMove(m)
			static := NewEngine(pos, EngineOptions{})
			if dynamic.Position.Zobrist != static.Position.Zobrist {
				t.Logf("expected static zobrist hash %v, got dynamic zobrist hash %v",
					static.Position.Zobrist, dynamic.Position.Zobrist)
				t.FailNow()
			}
		}
	}
}

func TestMateIn1(t *testing.T) {
	game := []struct {
		move []string
		fen  string
	}{
		{[]string{"a3a8"}, "2kn3r/5p2/2p5/1pN1B3/4P3/R5P1/7P/R4K2 w - - 6 46"},
		{[]string{"d8b8"}, "3Q4/3b4/2Nk4/3P3p/1KP4P/8/5p2/8 w - - 0 87"},
		{[]string{"g4f2"}, "r3k3/pR4B1/2p1p1p1/2N5/P4nn1/2P5/6r1/7K b - - 5 40"},
		{[]string{"f6f4", "f6f2"}, "1r6/4k3/5q2/p2pp3/2P5/1P1QK3/2P3rP/1R6 b - - 1 39"},
		{[]string{"e8e1"}, "4r2k/ppp2pp1/2n4p/1r6/2PP4/5R2/5PPP/2N3KR b - - 0 30"},
		{[]string{"g1g8", "h5h6"}, "6rk/2pr1p2/p3pN1p/1p5R/5P2/3P4/PP2K2P/6R1 w - - 4 29"},
	}

	for _, g := range game {
		pos, _ := PositionFromFEN(g.fen)
		eng := NewEngine(pos, EngineOptions{})

		for d := 3; d < 5; d++ {
			tc := &FixedDepthTimeControl{MinDepth: d, MaxDepth: d}
			tc.Start()

			move, _ := eng.Play(tc)
			found := false
			for _, m := range g.move {
				if move.String() == m {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("at depth %d expected one of %v, got %v for position %s",
					d, g.move, move, g.fen)
			}
		}
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
			eng := NewEngine(pos, EngineOptions{})
			tc := &FixedDepthTimeControl{MinDepth: 3, MaxDepth: 5}
			tc.Start()
			eng.Play(tc)
		}
	}
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		eng := NewEngine(pos, EngineOptions{})
		for i := 0; i < 20; i++ {
			tc := &FixedDepthTimeControl{MinDepth: 2, MaxDepth: 4}
			tc.Start()
			move, _ := eng.Play(tc)
			eng.DoMove(move)
		}
	}
}

func BenchmarkScore(b *testing.B) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, EngineOptions{})

	for i := 0; i < b.N; i++ {
		for _, g := range games {
			todo := strings.Fields(g)
			done := make([]Move, 0)

			for i := range todo {
				move := eng.ParseMove(todo[i])
				done = append(done, move)
				eng.DoMove(move)
				_ = eng.Score()
			}

			for i := range done {
				move := done[len(done)-1-i]
				eng.UndoMove(move)
				_ = eng.Score()
			}
		}
	}
}
