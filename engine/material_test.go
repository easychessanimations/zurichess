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
	e.position.GenerateMoves(Violent, &moves)
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
	good, bad := 0, 0
	for i, fen := range testFENs {
		var moves []Move
		pos, _ := PositionFromFEN(fen)
		e := MakeEvaluation(pos, &GlobalMaterial)
		e.position.GenerateMoves(All, &moves)
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
	e.position.GenerateMoves(All, &moves)
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
	e.position.GenerateMoves(All, &moves)
	for i := 0; i < b.N; i++ {
		for _, m := range moves {
			e.SEE(m)
		}
	}
}
