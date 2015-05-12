package engine

import (
	"math"
	"testing"
)

func TestStack(t *testing.T) {
	ms := &moveStack{}
	for _, fen := range []string{FENStartPos, FENKiwipete, FENDuplain} {
		pos, _ := PositionFromFEN(fen)
		ms.GenerateMoves(pos, Move(0), [2]Move{})

		limit := int16(math.MaxInt16)
		for move := Move(0); ms.PopMove(&move); {
			if curr := mvvlva(move); curr > limit {
				t.Errorf("moves not sorted")
			} else {
				limit = curr
			}
		}
	}
}
