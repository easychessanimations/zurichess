package main

import (
	"testing"

	"bitbucket.org/brtzsnr/zurichess/engine"
)

func testHelper(t *testing.T, fen string, testData []counters) {
	for depth, expected := range testData {
		if testing.Short() && expected.nodes > 200000 {
			return
		}

		pos, err := engine.PositionFromFEN(fen)
		if err != nil {
			t.Errorf("invalid FEN: %s", fen)
		}

		actual := perft(pos, depth, hashTable, new([]engine.Move))
		if expected != actual {
			t.Errorf("at depth %d expected %+v got %+v", depth, expected, actual)
		}
	}
}

func TestPerftInitial(t *testing.T) {
	testHelper(t, startpos, data[startpos][:6])
}

func TestPerftKiwipete(t *testing.T) {
	testHelper(t, kiwipete, data[kiwipete][:5])
}

func TestPerftDuplain(t *testing.T) {
	testHelper(t, duplain, data[duplain][:7])
}

func benchHelper(b *testing.B, fen string, depth int) {
	pos, _ := engine.PositionFromFEN(fen)
	for i := 0; i < b.N; i++ {
		perft(pos, depth, nil, new([]engine.Move))
	}
}

func BenchmarkPerftInitial(b *testing.B) {
	benchHelper(b, startpos, 4)
}

func BenchmarkPerftKiwipete(b *testing.B) {
	benchHelper(b, kiwipete, 3)
}

func BenchmarkPerftDuplain(b *testing.B) {
	benchHelper(b, duplain, 4)
}
