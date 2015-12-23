package engine

var (
	// Some games played by zurichess against itself or other engines.
	testGames = []string{
		"g2g3 b8c6 c2c3 d7d5 f1h3 e7e6 d1b3 f8d6 f2f4 g8f6 b3d1 e8g8 d1c2 c8d7 b2b3 f6e4 h3g4 d6f4 g3f4 d8h4 e1d1 e4f2 d1e1 f2d3 e1f1 h4f2",
		"h2h3 e7e5 g1f3 b8c6 f3g1 d7d5 h1h2 f8b4 c2c3 b4f8 h3h4 f8d6 g2g3 g8f6 f1h3 c8e6 h3e6 f7e6 d1a4 h8f8 h4h5 a7a6 h5h6 g7g6 d2d3 b7b5 a4h4 a6a5 c1g5 d5d4 g5f6 d8f6 h4f6 f8f6 b1d2 d4c3 b2c3 g6g5 d2e4 f6f5 g3g4 f5f7 e4g5 e5e4 h2g2 d6e5 d3d4 e5f6 g5f7 e8f7 a2a4 b5a4 a1a4 c6e7 g4g5 f6h8 g2g4 c7c5 g4e4 e7c6 d4c5 h8e5 e4e3 e5c7 a4a2 c7f4 e3f3 e6e5 e2e3 c6e7 e3f4 e7d5 f3d3 d5f4 d3d7 f7g8 d7g7 g8h8 c5c6 a8d8 c6c7 d8c8 g7d7 c8g8 a2a5 f4g2 e1d2 g2f4 d7d8 f4g6 c7c8q g6f8 d8f8 e5e4 f8g8",
		"e2e4 d7d5 e4e5 f7f6 d2d4 e7e6 f1b5 b8c6 e5f6 d8f6 b5c6 b7c6 h2h3 c6c5 d4c5 f6g6 d1d3 g6g2 d3f3 g2f3 g1f3 f8c5 c1e3 c5e7 e3d4 g8f6 f3e5 a8b8 e5c6 b8a8 c6e7 e8e7 d4f6 e7f6 b1c3 c8b7 c3b5 c7c6 b5c3 c6c5 c3a4 h8c8 h3h4 d5d4 h1h3 b7e4 c2c4 f6e5 h3g3 g7g6 e1c1 a8b8 g3g4 e4f5 f2f4 e5e4 g4g5 b8b7 d1h1 b7b4 h1e1 e4d3 b2b3 f5g4 g5g4 e6e5 g4g3",
	}

	// Few test positions from past bugs.
	testFENs = []string{
		// Initial position
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		// Kiwipete: https://chessprogramming.wikispaces.com/Perft+Results
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		// Duplain: https://chessprogramming.wikispaces.com/Perft+Results
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		// Underpromotion: http://www.stmintz.com/ccc/index.php?id=366606
		"8/p1P5/P7/3p4/5p1p/3p1P1P/K2p2pp/3R2nk w - - 0 1",
		// Enpassant: http://www.10x8.net/chess/PerfT.html
		"8/7p/p5pb/4k3/P1pPn3/8/P5PP/1rB2RK1 b - d3 0 28",
		// http://www.talkchess.com/forum/viewtopic.php?t=48609
		"1K1k4/8/5n2/3p4/8/1BN2B2/6b1/7b w - - 0 1",
		// http://www.talkchess.com/forum/viewtopic.php?t=51272
		"6k1/5ppp/3r4/8/3R2b1/8/5PPP/R3qB1K b - - 0 1",
		// http://www.stmintz.com/ccc/index.php?id=206056
		"2rqkb1r/p1pnpppp/3p3n/3B4/2BPP3/1QP5/PP3PPP/RN2K1NR w KQk - 0 1",
		// http://www.stmintz.com/ccc/index.php?id=60880
		"1rr3k1/4ppb1/2q1bnp1/1p2B1Q1/6P1/2p2P2/2P1B2R/2K4R w - - 0 1",
		// https://chessprogramming.wikispaces.com/SEE+-+The+Swap+Algorithm
		"1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - - 0 1",
		"1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - - 0 1",
		// http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=419315&t=40054
		"8/8/3p4/4r3/2RKP3/5k2/8/8 b - - 0 1",
		// zurichess: various
		"8/K5p1/1P1k1p1p/5P1P/2R3P1/8/8/8 b - - 0 78",
		"8/1P6/5ppp/3k1P1P/6P1/8/1K6/8 w - - 0 78",
		"1K6/1P6/5ppp/3k1P1P/6P1/8/8/8 w - - 0 1",
		"r1bqkb1r/ppp1pp2/2n3P1/3p4/3Pn3/5N1P/PPP1PPB1/RNBQK2R b KQkq - 0 1",
		"r1bqkb1r/ppp2p2/2n1p1pP/3p4/3Pn3/2N2N1P/PPP1PPB1/R1BQK2R b KQkq - 0 1",
		"r3kb2/ppp2pp1/6n1/7Q/8/2P1BN1b/1q2PPB1/3R1K1R b q - 0 1",
		"r7/1p4p1/2p2kb1/3r4/3N3n/4P2P/1p2BP2/3RK1R1 w - - 0 1",
		"r7/1p4p1/5k2/8/6P1/3Nn3/1p3P2/3BK3 w - - 0 1",
		"8/1p2k1p1/4P3/8/1p2N3/4P3/5P2/3BK3 b - - 0 1",
		// zurichess: many captures
		"6k1/Qp1r1pp1/p1rP3p/P3q3/2Bnb1P1/1P3PNP/4p1K1/R1R5 b - - 0 1",
		"3r2k1/2Q2pb1/2n1r3/1p1p4/pB1PP3/n1N2p2/B1q2P1R/6RK b - - 0 1",
		"2r3k1/5p1n/6p1/pp3n2/2BPp2P/4P2P/q1rN1PQb/R1BKR3 b - - 0 1",
		"r3r3/bpp1Nk1p/p1bq1Bp1/5p2/PPP3n1/R7/3QBPPP/5RK1 w - - 0 1",
		"4r1q1/1p4bk/2pp2np/4N2n/2bp2pP/PR3rP1/2QBNPB1/4K2R b K - 0 1",
		// crafted:
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		// ECM
		"7r/1p2k3/2bpp3/p3np2/P1PR4/2N2PP1/1P4K1/3B4 b - - 0 1",
		"4k3/p1P3p1/2q1np1p/3N4/8/1Q3PP1/6KP/8 w - - 0 1",
		"3q4/pp3pkp/5npN/2bpr1B1/4r3/2P2Q2/PP3PPP/R4RK1 w - - 0 1",
		"4k3/p1P3p1/2q1np1p/3N4/8/1Q3PP1/7P/5K2 b - - 1 1",
	}

	fenKiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	fenDuplain  = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
)
