// Perft test.
// https://chessprogramming.wikispaces.com/Perft
package main

import (
	"testing"
)

type counters struct {
	nodes     uint64
	captures  uint64
	enpassant uint64
}

func (co *counters) Add(ot counters) {
	co.nodes += ot.nodes
	co.captures += ot.captures
	co.enpassant += ot.enpassant
}

func (co *counters) Equal(ot counters) bool {
	return co.nodes == ot.nodes &&
		co.captures == ot.captures &&
		co.enpassant == ot.enpassant
}

func perft(pos *Position, depth int) counters {
	if depth == 0 {
		return counters{1, 0, 0}
	}

	co := counters{}
	moves := pos.GenerateMoves()
	for _, mo := range moves {
		if mo.Capture != NoPiece {
			co.captures++
		}
		if mo.MoveType == Enpassant {
			co.enpassant++
		}

		pos.DoMove(mo)
		co.Add(perft(pos, depth-1))
		pos.UndoMove(mo)
	}
	return co
}

func testHelper(t *testing.T, fen string, testData []counters) {
	for depth, expected := range testData {
		pos, err := PositionFromFEN(fen)
		if err != nil {
			t.Errorf("invalid FEN: %s", fen)
		}
		pos.PrettyPrint()

		actual := perft(pos, depth)
		if !expected.Equal(actual) {
			t.Errorf("at depth %d expected %+v got %+v",
				depth, expected, actual)
		}
	}
}

func TestPerftInitial(t *testing.T) {
	testHelper(t, FENStartPos, []counters{
		{1, 0, 0},
		{20, 0, 0},
		{400, 0, 0},
		{8902, 34, 0},
		{197281, 1576, 0},
		// {4865609, 82719, 0},
		// {119060324, 2812008, 5248},
	})

}

func TestPerftKiwipete(t *testing.T) {
	testHelper(t, FENKiwipete, []counters{
		{1, 0, 0},
		{48, 8, 0},
		{2039, 351, 1},
		{97862, 17102, 45},
		// {4085603, 757163, 1929},
		// {193690690, 35043416, 73365},
	})
}

func TestPerftDuplain(t *testing.T) {
	fen := "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -"
	testHelper(t, fen, []counters{
		{1, 0, 0},
		{14, 1, 0},
		{191, 14, 0},
		{2812, 209, 2},
		{43238, 3348, 123},
		// {674624, 52051, 1165},
		// {11030083, 940350, 33325},
		// {178633661, 14519036, 294874},
	})
}
