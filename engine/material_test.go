package engine

import (
	"strings"
	"testing"
)

func TestStaticScore(t *testing.T) {
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := MakeEvaluation(pos, &GlobalMaterial)

		moves := strings.Fields(game)
		for i, move := range moves {
			m := pos.UCIToMove(move)
			pos.DoMove(m)
			dynamic.DoMove(m)
			static := MakeEvaluation(pos, &GlobalMaterial)
			if static.Static != dynamic.Static {
				t.Errorf("move #%d, expected %v got %v", i, static.Static, dynamic.Static)
			}
		}
	}
}
