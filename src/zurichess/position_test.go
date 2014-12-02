package main

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

func checkPiece(t *testing.T, pi Piece, co Color, pt PieceType) {
	if pi.Color() != co && pi.PieceType() != pt {
		t.Errorf("expected %v %v, got %v %v", co, pt, pi.Color(), pi.PieceType())
	}
}

// TestPiece verifies Piece functionality.
func TestPiece(t *testing.T) {
	checkPiece(t, NoPiece, NoColor, NoPieceType)
	for co := ColorMinValue; co < ColorMaxValue; co++ {
		for pt := PieceTypeMinValue; pt < PieceTypeMaxValue; pt++ {
			checkPiece(t, ColorPiece(co, pt), co, pt)
		}
	}
}

// TestPutGetPiece tests PutPiece and GetPiece.
func TestPutGetPiece(t *testing.T) {
	var pi Piece
	pos := &Position{}

	pi = pos.GetPiece(SquareA3)
	checkPiece(t, pi, NoColor, NoPieceType)

	pos.PutPiece(SquareA3, ColorPiece(White, Pawn))
	pi = pos.GetPiece(SquareA3)
	checkPiece(t, pi, White, Pawn)

	pos.PutPiece(SquareH7, ColorPiece(Black, King))
	pi = pos.GetPiece(SquareH7)
	checkPiece(t, pi, Black, King)
}

func TestStartPosition(t *testing.T) {
	pos, err := PositionFromFEN(FENStartPos)
	if err != nil {
		t.Error(err)
	}

	ranks := []struct {
		rank  int
		color Color
	}{
		{0, White},
		{1, White},
		{6, Black},
		{7, Black},
	}
	for _, r := range ranks {
		for f := 0; f < 7; f++ {
			co := pos.GetPiece(RankFile(r.rank, f)).Color()
			if co != r.color {
				t.Errorf("expected %v on rank %d, got %v", r.color, r.rank, co)
			}
		}
	}

	for r := 0; r < 7; r++ {
		for f := 0; f < 7; f++ {
			sq := RankFile(r, f)
			if (r <= 1 || r >= 6) && pos.IsEmpty(sq) {
				t.Errorf("expected piece at %v", sq)
			}
			if (r > 1 && r < 6) && !pos.IsEmpty(sq) {
				t.Errorf("expected no piece at %v", sq)

			}
		}
	}

	pieces := []struct {
		square Square
		piece  Piece
	}{
		{SquareD1, ColorPiece(White, Queen)},
		{SquareE1, ColorPiece(White, King)},
		{SquareD8, ColorPiece(Black, Queen)},
		{SquareE8, ColorPiece(Black, King)},
	}
	for _, p := range pieces {
		actual := pos.GetPiece(p.square)
		if actual != p.piece {
			t.Errorf("expected %v at %v, got %v", p.piece, p.square, actual)
		}
	}

}

func testMoves(t *testing.T, moves []Move, expected []string) {
	seen := make(map[string]bool)
	for _, e := range expected {
		seen[e] = false
	}
	for _, mo := range moves {
		str := mo.String()
		if dup, has := seen[str]; !has {
			t.Error("move", str, "was not expected")
		} else if dup {
			t.Error("move", str, "already seen")
		}
		seen[str] = true
	}
	for mo, has := range seen {
		if !has {
			t.Error("missing move", mo)
		}
	}
}

func TestGenKnightMoves(t *testing.T) {
	var moves []Move
	var expected []string
	pos := &Position{}
	kn := ColorPiece(White, Knight)

	pos.PutPiece(SquareB2, kn)
	pos.PutPiece(SquareE4, kn)
	pos.PutPiece(SquareC4, ColorPiece(White, Pawn))

	moves = pos.genKnightMoves(SquareB2, kn, nil)
	expected = []string{"b2d1", "b2d3", "b2a4"}
	testMoves(t, moves, expected)

	moves = pos.genKnightMoves(SquareF4, kn, nil)
	expected = []string{"f4d3", "f4d5", "f4e6", "f4g6", "f4h5", "f4h3", "f4g2", "f4e2"}
	testMoves(t, moves, expected)
}
