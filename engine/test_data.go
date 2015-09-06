package engine

var (
	// Some games played by zurichess against itself or other engines.
	testGames = []string{
		"g2g3 b8c6 c2c3 d7d5 f1h3 e7e6 d1b3 f8d6 f2f4 g8f6 b3d1 e8g8 d1c2 c8d7 b2b3 f6e4 h3g4 d6f4 g3f4 d8h4 e1d1 e4f2 d1e1 f2d3 e1f1 h4f2",
		"h2h3 e7e5 g1f3 b8c6 f3g1 d7d5 h1h2 f8b4 c2c3 b4f8 h3h4 f8d6 g2g3 g8f6 f1h3 c8e6 h3e6 f7e6 d1a4 h8f8 h4h5 a7a6 h5h6 g7g6 d2d3 b7b5 a4h4 a6a5 c1g5 d5d4 g5f6 d8f6 h4f6 f8f6 b1d2 d4c3 b2c3 g6g5 d2e4 f6f5 g3g4 f5f7 e4g5 e5e4 h2g2 d6e5 d3d4 e5f6 g5f7 e8f7 a2a4 b5a4 a1a4 c6e7 g4g5 f6h8 g2g4 c7c5 g4e4 e7c6 d4c5 h8e5 e4e3 e5c7 a4a2 c7f4 e3f3 e6e5 e2e3 c6e7 e3f4 e7d5 f3d3 d5f4 d3d7 f7g8 d7g7 g8h8 c5c6 a8d8 c6c7 d8c8 g7d7 c8g8 a2a5 f4g2 e1d2 g2f4 d7d8 f4g6 c7c8q g6f8 d8f8 e5e4 f8g8",
		"e2e4 d7d5 e4e5 f7f6 d2d4 e7e6 f1b5 b8c6 e5f6 d8f6 b5c6 b7c6 h2h3 c6c5 d4c5 f6g6 d1d3 g6g2 d3f3 g2f3 g1f3 f8c5 c1e3 c5e7 e3d4 g8f6 f3e5 a8b8 e5c6 b8a8 c6e7 e8e7 d4f6 e7f6 b1c3 c8b7 c3b5 c7c6 b5c3 c6c5 c3a4 h8c8 h3h4 d5d4 h1h3 b7e4 c2c4 f6e5 h3g3 g7g6 e1c1 a8b8 g3g4 e4f5 f2f4 e5e4 g4g5 b8b7 d1h1 b7b4 h1e1 e4d3 b2b3 f5g4 g5g4 e6e5 g4g3",
		"d2d4 d7d5 c1f4 c8f5 g1f3 g8f6 e2e3 d8b6 b2b3 e7e6 f1d3 f8b4 c2c3 f5d3 d1d3 b4e7 b1d2 e8f8 h4h5 h7h6 e1f1 f8g8 f1g1 b8d7 f3e5 d7e5 f4e5 f6d7 e5g3 d7f6 c3c4 e7b4 c4c5 b6a7 g3c7 a8c8 c7e5 f6d7 d2f3 b7b6 c5b6 a7b6 a1f1 c6c5 f1c1 b4a3 c1c2 d7e5 f3e5 b6d6 f2f4 d6e7 e3e4 c8c7 f4f5 d5e4 d3e4 e6f5 e4a8 g8h7 a8a5 c7b7 a5c3 c5d4 c3d4 h8e8 e5f3 b7b3 f3d2 b3b1 c2c1 b1c1 d2f1 c1f1 g1h2 a3d6 g2g3 e7e2 d4f2 e2f2 h2h3 f1h1",
		"f1b5 c8d7 b1c3 d7b5 c3b5 f8g7 g1h3 c7c6 b5a3 g8f6 e1g1 e8g8 d2d4 c6c5 c1d2 c5d4 e3d4 b8c6 c2c3 d8b6 b2b3 b6a6 a3c2 e7e6 h3f2 a8c8 d1f3 a6a5 a2a4 a7a6 f3d3 a5f5 a4a5 f5d3 f2d3 f6e4 d2e1 c6d4 c2d4 e4c3 d4f3 c3e2 g1f2 g7a1 f2e2 a1c3 e1c3 c8c3 f3d2 f8c8 g2g4 c3c2 f1a1 h7h5 g4h5 g6h5 a1g1 g8h7 e2e3 c8g8 g1d1 h7g7 d3b4 c2c3 b4d3 h5h4 e3d4 c3c2 d3b4 c2c5 d2c4 g7f6 c4d6 c5a5 h2h3 g8d8 d4c3 a5c5 c3b2 c5c7 d6e4 f6e7 d1d8 e7d8 b4d3 d8e7 d3e5 a6a5 e4c3 c7c8 c3e4 c8d8 e4c5 d8b8 e5c4 b7b6 c5d3 e7f6 b2c3 f6g6 c3d4 g6f5 c4e3 f5f6 d3e5 f6g7 e3c4 b8d8 d4c3 d8d1 c4b6 d1h1 f4f5 h1h3 c3c2 e6f5 b6c4 h3g3 c4d6 g7f6 e5c4 h4h3 c4d2 g3g2 d6e8 f6g6 e8c7 h3h2 c7d5 g2f2 d5b6 h2h1Q b6c4 g6f6 c2d3",
		"e2e4 d7d5 f1e2 e7e5 e1f1 g4h6 e4d5 e5f4 e2f3 f8d6 d1e2 e8f8 b1c3 f8g8 d2d4 c8f5 g2g3 f4g3 h2g3 b8d7 g3g4 f5g6 c1h6 g7h6 h1h3 d8f6 e2f2 a8e8 c3b5 d7b6 f1g2 f6g5 b5d6 c7d6 g2h2 b6d5 h3g3 g5e3 g1h3 e3f2 h3f2 d5b6 f3b7 e8e2 h2g1 e2c2 b2b3 c2d2 a1d1 d2b2 g3e3 g6c2 e3e8 g8g7 e8e2 b2a2 b3b4 h8b8 b7c6 c2d1 e2a2 b8c8 b4b5 d1b3 a2a1 c8g8 g1g2 b6c4 g2f3 b3a4 f2e4 a7a6 e4c3 a6b5 c6b5 a4b5 c3b5 g7g6 f3f4 g8e8 a3a4 e8e2 b5c3 e2f2 f4e4 f2g2 e4f3 g2h2 c3d5 h2h3 f3e2 h3g3 a4a5 c4a5 a1a5 g3g4 e2d3 g4g5 d3e4 g6g7 a5a7 h6h5 d5e3 g7f6 e3d5 f6g7 d5e3 g7f6 a7a8 h5h4 e4f4 h7h6 a8a6 h4h3 a6a1 g5h5 f4g3 f6g6 a1h1 g6g5 h1f1 g5g6 g3h2 f7f6 f1f3 f6f5 e3c4 g6f6 d4d5 f5f4 f3f4 h5f5 f4f5 f6f5 h2h3 f5e4 c4b6 e4d4 h3g4 d4c5 b6d7 c5d5 g4f5 h6h5 f5g5 d5c4 d7b6 c4c5 b6a4 c5d4 g5h5 d6d5 h5g5 d4e3 a4c3 d5d4 c3b5 d4d3 b5a3 d3d2 a3c4 e3e4 c4d2",
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
	}

	fenKiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	fenDuplain  = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
)
