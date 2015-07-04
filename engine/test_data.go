package engine

var (
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
		// zurichess: many captures
		"6k1/Qp1r1pp1/p1rP3p/P3q3/2Bnb1P1/1P3PNP/4p1K1/R1R5 b - - 0 1",
		"3r2k1/2Q2pb1/2n1r3/1p1p4/pB1PP3/n1N2p2/B1q2P1R/6RK b - - 0 1",
		"2r3k1/5p1n/6p1/pp3n2/2BPp2P/4P2P/q1rN1PQb/R1BKR3 b - - 0 1",
		"r3r3/bpp1Nk1p/p1bq1Bp1/5p2/PPP3n1/R7/3QBPPP/5RK1 w - - 0 1",
		"4r1q1/1p4bk/2pp2np/4N2n/2bp2pP/PR3rP1/2QBNPB1/4K2R b K - 0 1",
		// crafted:
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	}
)
