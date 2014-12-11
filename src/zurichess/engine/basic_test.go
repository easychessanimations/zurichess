package engine

import (
	"testing"
)

func TestSquareFromString(t *testing.T) {
	data := []struct {
		sq  Square
		str string
	}{
		{SquareF4, "f4"},
		{SquareA3, "a3"},
		{SquareC1, "c1"},
		{SquareH8, "h8"},
	}

	for _, d := range data {
		if d.sq.String() != d.str {
			t.Errorf("expected %v, got %v", d.str, d.sq.String())
		}
		if d.sq != SquareFromString(d.str) {
			t.Errorf("expected %v, got %v", d.sq, SquareFromString(d.str))
		}
	}
}

func TestRankFile(t *testing.T) {
	for r := 0; r < 7; r++ {
		for f := 0; f < 7; f++ {
			sq := RankFile(r, f)
			if sq.Rank() != r || sq.File() != f {
				t.Errorf("expected (rank, file) (%d, %d), got (%d, %d)",
					r, f, sq.Rank(), sq.File())
			}
		}
	}
}

func checkPiece(t *testing.T, pi Piece, co Color, pt Figure) {
	if pi.Color() != co && pi.Figure() != pt {
		t.Errorf("expected %v %v, got %v %v", co, pt, pi.Color(), pi.Figure())
	}
}

// TestPiece verifies Piece functionality.
func TestPiece(t *testing.T) {
	checkPiece(t, NoPiece, NoColor, NoFigure)
	for co := ColorMinValue; co < ColorMaxValue; co++ {
		for pt := FigureMinValue; pt < FigureMaxValue; pt++ {
			checkPiece(t, ColorFigure(co, pt), co, pt)
		}
	}
}
