package engine

import (
	"testing"
)

func TestPositionFromFENAndBack(t *testing.T) {
	data := []string{
		FENKiwipete,
		FENStartPos,
		FENDuplain,
	}

	for _, d := range data {
		pos, err := PositionFromFEN(d)
		if err != nil {
			t.Errorf("%s failed with %v", d, err)
		} else if fen := pos.String(); d != fen {
			t.Errorf("expected %s, got %s", d, fen)
		}
	}
}

func BenchmarkPositionFromFEN(b *testing.B) {
	data := []string{
		FENKiwipete,
		FENStartPos,
		FENDuplain,
	}

	for i := 0; i < b.N; i++ {
		for _, d := range data {
			PositionFromFEN(d)
		}
	}
}
