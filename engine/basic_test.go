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
		if sq, err := SquareFromString(d.str); err != nil {
			t.Errorf("parse error: %v", err)
		} else if d.sq != sq {
			t.Errorf("expected %v, got %v", d.sq, sq)
		}
	}
}

func TestRookSquare(t *testing.T) {
	data := []struct {
		kingEnd, rookStart, rookEnd Square
	}{
		{SquareC1, SquareA1, SquareD1},
		{SquareC8, SquareA8, SquareD8},
		{SquareG1, SquareH1, SquareF1},
		{SquareG8, SquareH8, SquareF8},
	}

	for _, d := range data {
		_, rookStart, rookEnd := CastlingRook(d.kingEnd)
		if rookStart != d.rookStart || rookEnd != d.rookEnd {
			t.Errorf("for king to %v, expected rook from %v to %v, got rook from %v to %v",
				d.kingEnd, d.rookStart, d.rookEnd, rookStart, rookEnd)
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

func checkPiece(t *testing.T, pi Piece, co Color, fig Figure) {
	if pi.Color() != co || pi.Figure() != fig {
		t.Errorf("for %v expected %v %v, got %v %v", pi, co, fig, pi.Color(), pi.Figure())
	}
}

// TestPiece verifies Piece functionality.
func TestPiece1(t *testing.T) {
	checkPiece(t, NoPiece, NoColor, NoFigure)
	for co := ColorMinValue; co < ColorMaxValue; co++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			checkPiece(t, ColorFigure(co, fig), co, fig)
		}
	}
}

func TestPiece2(t *testing.T) {
	checkPiece(t, WhitePawn, White, Pawn)
	checkPiece(t, WhiteKnight, White, Knight)
	checkPiece(t, WhiteRook, White, Rook)
	checkPiece(t, WhiteKing, White, King)
	checkPiece(t, BlackPawn, Black, Pawn)
	checkPiece(t, BlackBishop, Black, Bishop)
}

func TestCastlingRook(t *testing.T) {
	data := []struct {
		kingEnd Square
		rook    Piece
	}{
		{SquareC1, WhiteRook},
		{SquareC8, BlackRook},
		{SquareG1, WhiteRook},
		{SquareG8, BlackRook},
	}

	for _, d := range data {
		rook, _, _ := CastlingRook(d.kingEnd)
		if rook != d.rook {
			t.Errorf("for king to %v, expected %v, got %v", d.kingEnd, d.rook, rook)
		}
	}
}
