package engine

import (
	"strings"
	"testing"
)

var (
	// Mate in one tests.
	// Position taken from http://www.hoflink.com/~npollock/chess.html,
	// with all positions that have more than one solution removed.
	mateIn1 = []struct {
		fen string
		bm  string
	}{
		{"1k1r4/2p2ppp/8/8/Qb6/2R1Pn2/PP2KPPP/3r4 b - - 0 1", "Ng1+"},
		{"1kqr4/2n2r2/1Np3pp/2p1pp2/4P3/Q2PP3/P5PP/1R4K1 w - - 0 1", "Nd7+"},
		{"1n4rk/1bp2Q1p/p2p4/1p2p3/5N1N/1P1P3P/1PP2p1K/8 b - - 0 1", "f1N+"},
		{"1q2r3/3kPpp1/1p1P1b1p/3Q1P2/1p6/P6P/1P4P1/1KR2b2 w - - 0 1", "Qc6+"},
		{"1r3r1k/6pp/6b1/pBp3B1/Pn1N2P1/4p2P/1P6/2KR3R b - - 0 1", "Na2+"},
		{"1rqkn2r/p2n1R2/2p4p/2N3p1/6P1/7P/PPP5/2KR4 w - - 0 1", "Ne6+"},
		{"2B1nrk1/p5bp/1p1p4/4p3/8/1NPKnq1P/PP1Q4/R6R b - - 0 1", "e4+"},
		{"2R5/2p1rkpB/2b2p2/2P4P/1b3PP1/4B3/5K2/8 w - - 0 1", "Bg8+"},
		{"2k1r3/Qpnq3p/5pp1/3p4/8/BP4P1/P4P1P/2R3K1 w - - 0 1", "Qa8+"},
		{"2k1r3/p3P3/1p1q4/6p1/3PQ3/2P3pP/P5B1/6K1 w - - 0 1", "Qb7+"},
		{"2k2r2/1pp4P/p2n4/2Nn2R1/1P1P4/P1RK2Q1/1r4b1/8 b - - 0 1", "Bf1+"},
		{"2k5/pp4pp/1b6/2nP4/5pb1/P7/1P2QKPP/5R2 b - - 0 1", "Nd3+"},
		{"2k5/ppp2p2/7q/6p1/2Nb1p2/1B3Kn1/PP2Q1P1/8 b - - 0 1", "Qh5+"},
		{"3k3B/7p/p1Q1p3/2n5/6P1/K3b3/PP5q/R7 w - - 0 1", "Bf6+"},
		{"3r2k1/ppp2ppp/6Q1/b7/3n1B2/2p3n1/P4PPP/RN3RK1 b - - 0 1", "Nde2+"},
		{"3rkr2/5p2/b1p2p2/4pP1P/p3P1Q1/b1P5/B1K2RP1/2RNq3 b - - 0 1", "Bd3+"},
		{"4bk2/ppp3p1/2np3p/2b5/2B2Bnq/2N5/PP4PP/4RR1K w - - 0 1", "Bxd6+"},
		{"4r1k1/pp3ppp/6q1/3p4/2bP1n2/P1Q2B2/1P3PPP/6KR b - - 0 1", "Nh3+"},
		{"4rkr1/1p1Rn1pp/p3p2B/4Qp2/8/8/PPq2PPP/3R2K1 w - - 0 1", "Qf6+"},
		{"5r2/p1n3k1/1p3qr1/7R/8/1BP1Q3/P5R1/6K1 w - - 0 1", "Qh6+"},
		{"5rk1/5ppp/p7/1pb1P3/7R/7P/PP2b2P/R1B4K b - - 0 1", "Bf3+"},
		{"5rkr/ppp2p1p/8/3qp3/2pN4/8/PPPQ1PPP/4R1K1 w - - 0 1", "Qg5+"},
		{"6k1/5qpp/pn1p2N1/B1p2p1P/Q3p3/2K1P2R/1r2BPP1/1r5R b - - 0 1", "Nxa4+"},
		{"6n1/5P1k/7p/np4b1/3B4/1pP4P/5PP1/1b4K1 w - - 0 1", "f8N+"},
		{"6q1/R2Q3p/1p1p1ppk/1P1N4/1P2rP2/6P1/7P/6K1 w - - 0 1", "Qh3+"},
		{"8/3b2p1/5P1k/1P2P3/1nP4K/p1N3PP/3P4/8 b - - 0 1", "g5+"},
		// {"8/4r2k/Q5b1/3P2P1/1qP2b1p/5PnB/P2P3R/2KR4 b - - 0 1", "Qc3+ Qb1+"},
		{"8/6P1/5K1k/6N1/5N2/8/8/8 w - - 0 1", "g8N+"},
		{"8/8/pp3Q2/7k/5Pp1/P1P3K1/3r3p/8 b - - 0 1", "h1N+"},
		{"8/p2k4/1p5R/2pp2R1/4n3/P2K4/1PP1N3/5r2 b - - 0 1", "Rf3+"},
		{"8/p4pkp/8/3B1b2/3b1ppP/P1N1r1n1/1PP3PR/R4QK1 b - - 0 1", "Re1+"},
		{"r1b1k2r/ppp1qppp/5B2/3Pn3/8/8/PPP2PPP/RN1QKB1R b KQkq - 0 1", "Nf3+"},
		{"r1b1kbnr/pppp1Npp/8/8/3nq3/8/PPPPBP1P/RNBQKR2 b Qkq - 0 1", "Nf3+"},
		{"r1b1q1kr/ppNnb1pp/5n2/8/3P4/8/PPP2PPP/R1BQKB1R b KQ - 0 1", "Bb4+"},
		{"r1b2rk1/pppp2p1/8/3qPN1Q/8/8/P5PP/b1B2R1K w - - 0 1", "Ne7+"},
		// {"r1b3k1/pppn3p/3p2rb/3P1K2/2P1P3/2N2P2/PP1QB3/R4R2 b - - 0 1", "Nf8+ Nb8+ Nb6+ Nc5+ Ne5+ Nf6+"},
		{"r1b3r1/5k2/1nn1p1p1/3pPp1P/p4P2/Kp3BQN/P1PBN1P1/3R3R b - - 0 1", "Nc4+"},
		{"r1bk3r/p1q1b1p1/7p/nB1pp1N1/8/3PB3/PPP2PPP/R3K2R w KQ - 0 1", "Nf7+"},
		{"r1bknb1r/pppnp1p1/3Np3/3p4/3P1B2/2P5/P3KPPP/7q w - - 0 1", "Nf7+"},
		{"r1bq2kr/pnpp3p/2pP1ppB/8/3Q4/8/PPP2PPP/RN2R1K1 w - - 0 1", "Qc4+"},
		{"r1bqk1nr/pppp1ppp/8/2b1P3/3nP3/6P1/PPP1N2P/RNBQKB1R b KQkq - 0 1", "Nf3+"},
		{"r1bqkb1r/pp1npppp/2p2n2/8/3PN3/8/PPP1QPPP/R1B1KBNR w KQkq - 0 1", "Nd6+"},
		{"r1bqr3/pp1nbk1p/2p2ppB/8/3P4/5Q2/PPP1NPPP/R3K2R w KQ - 0 1", "Qb3+"},
		{"r1q1r3/ppp1bpp1/2np4/5b1P/2k1NQP1/2P1B3/PPP2P2/2KR3R w - - 0 1", "Nxd6+"},
		{"r2Bk2r/ppp2p2/3b3p/8/1n1PK1b1/4P3/PPP2pPP/RN1Q1B1R b kq - 0 1", "f5+"},
		{"r2q1bnr/pp1bk1pp/4p3/3pPp1B/3n4/6Q1/PPP2PPP/R1B1K2R w KQ - 0 1", "Qa3+"},
		{"r2q1nr1/1b5k/p5p1/2pP1BPp/8/1P3N1Q/PB5P/4R1K1 w - - 0 1", "Qxh5+"},
		{"r2qk2r/pp1n2p1/2p1pn1p/3p4/3P4/B1PB1N2/P1P2PPP/R2Q2K1 w kq - 0 1", "Bg6+"},
		{"r2qk2r/pp3ppp/2p1p3/5P2/2Qn4/2n5/P2N1PPP/R1B1KB1R b KQkq - 0 1", "Nc2+"},
		{"r2qkb1r/1bp2ppp/p4n2/3p4/8/5p2/PPP1BPPP/RNBQR1K1 w kq - 0 1", "Bb5+"},
		{"r2r2k1/ppp2pp1/5q1p/4p3/4bn2/2PB2N1/P1PQ1P1P/R4RK1 b - - 0 1", "Nh3+"},
		{"r3k1nr/p1p2p1p/2pP4/8/7q/7b/PPPP3P/RNBQ2KR b kq - 0 1", "Qd4+"},
		{"r3k3/bppbq2r/p2p3p/3Pp2n/P1N1Pp2/2P2P1P/1PB3PN/R2QR2K b q - 0 1", "Ng3+"},
		{"r3kb1r/1p3ppp/8/3np1B1/1p6/8/PP3PPP/R3KB1R w KQkq - 0 1", "Bb5+"},
		{"r3rqkb/pp1b1pnp/2p1p1p1/4P1B1/2B1N1P1/5N1P/PPP2P2/2KR3R w - - 0 1", "Nf6+"},
		{"r4k1N/2p3pp/p7/1pbPn3/6b1/1P1P3P/1PP2qPK/RNB4Q b - - 0 1", "Nf3+"},
		{"r5r1/pQ5p/1qp2R2/2k1p3/P3P3/2PP4/2P3PP/6K1 w - - 0 1", "Qe7+"},
		{"r5r1/pppb1p2/3npkNp/8/3P2P1/2PB4/P1P1Q2P/6K1 w - - 0 1", "Qe5+"},
		{"r6r/pppk1ppp/8/2b5/2P5/2Nb1N2/PPnK1nPP/1RB2B1R b - - 0 1", "Be3+"},
		{"r7/1p4b1/p3Bp2/6pp/1PNN4/1P1k4/KB4P1/6q1 w - - 0 1", "Bf5+"},
		{"rk5r/p1q2ppp/Qp1B1n2/2p5/2P5/6P1/PP3PBP/4R1K1 w - - 0 1", "Qb7+"},
		{"rn1qkbnr/ppp2ppp/8/8/4Np2/5b2/PPPPQ1PP/R1B1KB1R w KQkq - 0 1", "Nf6+"},
		{"rn6/pQ5p/6r1/k1N1P3/3P4/4b1p1/1PP1K1P1/8 w - - 0 1", "b2-b4"},
		{"rnb3r1/pp1pb2p/2pk1nq1/6BQ/8/8/PPP3PP/4RRK1 w - - 0 1", "Bf4+"},
		{"rnbq3r/pppp2pp/1b6/8/1P2k3/8/PBPP1PPP/R2QK2R w KQ - 0 1", "Qf3+"},
		{"rnbqkb1r/ppp2ppp/8/3p4/8/2n2N2/PP2BPPP/R1B1R1K1 w kq - 0 1", "Bb5+"},
		{"rnbqkr2/pp1pbN1p/8/3p4/2B5/2p5/P4PPP/R3R1K1 w q - 0 1", "Nd6+"},
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

func TestMateIn1(t *testing.T) {
	for i, d := range mateIn1 {
		pos, _ := PositionFromFEN(d.fen)
		bm, err := pos.SANToMove(d.bm)
		if err != nil {
			t.Errorf("#%d cannot parse move %s", i, d.bm)
			continue
		}

		tc := NewFixedDepthTimeControl(pos, 2)
		tc.Start(false)
		eng := NewEngine(pos, Options{})
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
		dynamic := NewEngine(pos, Options{})
		static := NewEngine(pos, Options{})

		moves := strings.Fields(game)
		for _, move := range moves {
			m := pos.UCIToMove(move)
			if !pos.IsValid(m) {
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
		for _, g := range testGames {
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
