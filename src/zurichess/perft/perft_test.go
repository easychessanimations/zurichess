package main

import (
	"testing"

	"zurichess/engine"
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

		actual := perft(pos, depth, new([]engine.Move))
		if !expected.Equals(actual) {
			t.Errorf("at depth %d expected %+v got %+v",
				depth, expected, actual)
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
	testHelper(t, duplain, data[duplain][:6])
}
