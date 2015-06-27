package engine

import (
	"testing"
)

func (e *Evaluation) seeSlow(m Move, root bool) int32 {
	if m == NullMove {
		return 0
	}

	e.position.DoMove(m)

	// Find the smallest attacker.
	var moves []Move
	e.position.GenerateViolentMoves(&moves)
	next := NullMove
	for _, n := range moves {
		if n.To() != m.To() {
			continue
		}
		pi, sq := n.Piece().Figure(), n.From()
		if next == NullMove || pi < next.Piece().Figure() || (pi == next.Piece().Figure() && sq < next.From()) {
			next = n
		}
	}

	// Recursively compute the see.
	see := e.material.FigureBonus[m.Capture().Figure()].M - e.seeSlow(next, false)
	e.position.UndoMove(m)

	if root || see > 0 {
		return see
	}
	return 0
}

func TestSEE(t *testing.T) {
	fens := []string{
		FENKiwipete,
		FENDuplain,
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
	}

	good, bad := 0, 0
	for i, fen := range fens {
		var moves []Move
		pos, _ := PositionFromFEN(fen)
		e := MakeEvaluation(pos, &GlobalMaterial)
		e.position.GenerateMoves(&moves)
		for _, m := range moves {
			actual := e.SEE(m)
			expected := e.seeSlow(m, true)
			if expected != actual {
				t.Errorf("#%d expected %d, got %d\nfor %v on %v", i, expected, actual, m, fen)
				bad++
			} else {
				good++
			}
		}
	}

	if bad != 0 {
		t.Errorf("Failed %d out of %d", bad, good+bad)
	}
}

// A benchmark position from http://www.stmintz.com/ccc/index.php?id=60880
var seeBench = "1rr3k1/4ppb1/2q1bnp1/1p2B1Q1/6P1/2p2P2/2P1B2R/2K4R w - - 0 1"

func BenchmarkSEESlow(b *testing.B) {
	var moves []Move
	pos, _ := PositionFromFEN(seeBench)
	e := MakeEvaluation(pos, &GlobalMaterial)
	e.position.GenerateMoves(&moves)
	for i := 0; i < b.N; i++ {
		for _, m := range moves {
			e.seeSlow(m, true)
		}
	}
}

func BenchmarkSEEFast(b *testing.B) {
	var moves []Move
	pos, _ := PositionFromFEN(seeBench)
	e := MakeEvaluation(pos, &GlobalMaterial)
	e.position.GenerateMoves(&moves)
	for i := 0; i < b.N; i++ {
		for _, m := range moves {
			e.SEE(m)
		}
	}
}
