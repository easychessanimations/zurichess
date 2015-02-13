package engine

import (
	"strings"
	"testing"
)

func TestSortMoves(t *testing.T) {
	for _, fen := range []string{FENStartPos, FENKiwipete, FENDuplain} {
		pos, _ := PositionFromFEN(fen)
		moves := pos.GenerateMoves(nil)
		sortMoves(moves)
		for i := range moves {
			if i > 0 && score(&moves[i]) < score(&moves[i-1]) {
				t.Errorf("invalid move ordering")
			}
		}
	}
}

func BenchmarkSortMoves(b *testing.B) {
	var moves []Move
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		for i, move := range strings.Fields(game) {
			m := pos.UCIToMove(move)
			pos.DoMove(m)
			if 10 < i && i < 15 { // test only few midgame positions
				for j := 0; j < b.N; j++ {
					moves = pos.GenerateMoves(moves)
					sortMoves(moves)
				}
			}
		}
	}
}
