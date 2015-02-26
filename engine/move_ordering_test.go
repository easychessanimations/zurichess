package engine

import (
	"math"
	"testing"
)

func TestStack(t *testing.T) {
	ms := &moveStack{}
	for _, fen := range []string{FENStartPos, FENKiwipete, FENDuplain} {
		pos, _ := PositionFromFEN(fen)
		ms.Stack(pos.GenerateMoves, mvvlva)

		limit := int16(math.MaxInt16)
		var move Move
		for ms.PopMove(&move) {
			if curr := mvvlva(move); curr > limit {
				t.Errorf("moves not sorted")
			} else {
				limit = curr
			}
		}
	}
}
