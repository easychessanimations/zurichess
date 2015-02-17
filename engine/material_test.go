package engine

import (
	"strings"
	"testing"
)

// Test evaluation is the same if we start with the position or move.
func TestEvaluate(t *testing.T) {
	for _, game := range games {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := MidGameMaterial.EvaluatePosition(pos)

		moves := strings.Fields(game)
		for _, str := range moves {
			move := pos.UCIToMove(str)

			pos.DoMove(move)
			dynamic += MidGameMaterial.EvaluateMove(move)
			static := MidGameMaterial.EvaluatePosition(pos)

			t.Log("move", move, "piece", move.Target, "capture", move.Capture)
			if static != dynamic {
				t.Logf("expected static score %v, got dynamic score %v", static, dynamic)
				t.FailNow()
			}
		}
	}
}
