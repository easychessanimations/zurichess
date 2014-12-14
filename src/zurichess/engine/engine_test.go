package engine

import (
	"testing"
)

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		engWhite := &Engine{Position: pos}
		engBlack := &Engine{Position: pos}

		for m := 0; m < 10; m++ {
			if move, err := engWhite.Play(); err != nil {
				b.Log("white error: ", err)
				break
			} else {
				pos.DoMove(move)
			}

			if move, err := engBlack.Play(); err != nil {
				b.Log("black error: ", err)
				break
			} else {
				pos.DoMove(move)
			}
		}
	}
}
