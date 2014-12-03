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
	return co.nodes == ot.nodes && co.captures == ot.captures
	// && co.enpassant == ot.enpassant
}

func perftHelper(pos *Position, depth int) counters {
	if depth == 0 {
		return counters{1, 0, 0}
	}

	co := counters{}
	moves := pos.GenerateMoves()
	for _, mo := range moves {
		if mo.Capture() != NoPiece {
			co.captures++
		}
		if mo.MoveType == Enpassant {
			// co.enpassant++
		}

		pos.DoMove(mo)
		co.Add(perftHelper(pos, depth-1))
		pos.UndoMove(mo)
	}
	return co
}

func perft(fen string, depth int) counters {
	pos, err := PositionFromFEN(fen)
	if err != nil {
		panic("invalid FEN: " + fen)
	}
	return perftHelper(pos, depth)
}

func testHelper(t *testing.T, fen string, testData []counters) {
	for depth, expected := range testData {
		actual := perft(fen, depth)
		if !expected.Equal(actual) {
			t.Errorf("at depth %d expected %+v got %+v",
				depth, expected, actual)
		}
	}
}

func TestPerftInitial(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	testHelper(t, fen, []counters{
		{1, 0, 0},
		{20, 0, 0},
		{400, 0, 0},
		{8902, 34, 0},
		{197281, 1576, 0},
		// {4865609, 82719, 0},
	})

}

func TestPerftKiwipete(t *testing.T) {
	fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -"
	testHelper(t, fen, []counters{
		{1, 0, 0},
		{48, 8, 0},
		{2039, 351, 1},
		{4085603, 757163, 1929},
		// {193690690, 35043416, 73365},
	})
}
