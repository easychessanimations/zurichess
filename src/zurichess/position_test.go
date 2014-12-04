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

func testFENHelper(t *testing.T, expected *Position, fen string) {
	actual, err := PositionFromFEN(fen)
	if err != nil {
		t.Error(err)
		return
	}

	for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
		epi := expected.GetPiece(sq)
		api := actual.GetPiece(sq)
		if epi != api {
			t.Errorf("expected %v at %v, got %v", epi, sq, api)
		}
	}
}

func TestFENStartPosition(t *testing.T) {
	expected := &Position{}
	expected.PutPiece(SquareA1, ColorPiece(White, Rook))
	expected.PutPiece(SquareB1, ColorPiece(White, Knight))
	expected.PutPiece(SquareC1, ColorPiece(White, Bishop))
	expected.PutPiece(SquareD1, ColorPiece(White, Queen))
	expected.PutPiece(SquareE1, ColorPiece(White, King))
	expected.PutPiece(SquareF1, ColorPiece(White, Bishop))
	expected.PutPiece(SquareG1, ColorPiece(White, Knight))
	expected.PutPiece(SquareH1, ColorPiece(White, Rook))

	expected.PutPiece(SquareA8, ColorPiece(Black, Rook))
	expected.PutPiece(SquareB8, ColorPiece(Black, Knight))
	expected.PutPiece(SquareC8, ColorPiece(Black, Bishop))
	expected.PutPiece(SquareD8, ColorPiece(Black, Queen))
	expected.PutPiece(SquareE8, ColorPiece(Black, King))
	expected.PutPiece(SquareF8, ColorPiece(Black, Bishop))
	expected.PutPiece(SquareG8, ColorPiece(Black, Knight))
	expected.PutPiece(SquareH8, ColorPiece(Black, Rook))

	for f := 0; f < 8; f++ {
		expected.PutPiece(RankFile(1, f), ColorPiece(White, Pawn))
		expected.PutPiece(RankFile(6, f), ColorPiece(Black, Pawn))
	}

	testFENHelper(t, expected, FENStartPos)
}

func TestFENKiwipete(t *testing.T) {
	expected := &Position{}
	expected.PutPiece(SquareA1, ColorPiece(White, Rook))
	expected.PutPiece(SquareC3, ColorPiece(White, Knight))
	expected.PutPiece(SquareD2, ColorPiece(White, Bishop))
	expected.PutPiece(SquareF3, ColorPiece(White, Queen))
	expected.PutPiece(SquareE1, ColorPiece(White, King))
	expected.PutPiece(SquareE2, ColorPiece(White, Bishop))
	expected.PutPiece(SquareE5, ColorPiece(White, Knight))
	expected.PutPiece(SquareH1, ColorPiece(White, Rook))

	expected.PutPiece(SquareA8, ColorPiece(Black, Rook))
	expected.PutPiece(SquareB6, ColorPiece(Black, Knight))
	expected.PutPiece(SquareA6, ColorPiece(Black, Bishop))
	expected.PutPiece(SquareE7, ColorPiece(Black, Queen))
	expected.PutPiece(SquareE8, ColorPiece(Black, King))
	expected.PutPiece(SquareG7, ColorPiece(Black, Bishop))
	expected.PutPiece(SquareF6, ColorPiece(Black, Knight))
	expected.PutPiece(SquareH8, ColorPiece(Black, Rook))

	expected.PutPiece(SquareA2, ColorPiece(White, Pawn))
	expected.PutPiece(SquareB2, ColorPiece(White, Pawn))
	expected.PutPiece(SquareC2, ColorPiece(White, Pawn))
	expected.PutPiece(SquareD5, ColorPiece(White, Pawn))
	expected.PutPiece(SquareE4, ColorPiece(White, Pawn))
	expected.PutPiece(SquareF2, ColorPiece(White, Pawn))
	expected.PutPiece(SquareG2, ColorPiece(White, Pawn))
	expected.PutPiece(SquareH2, ColorPiece(White, Pawn))

	expected.PutPiece(SquareA7, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareB4, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareC7, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareD7, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareE6, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareF7, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareG6, ColorPiece(Black, Pawn))
	expected.PutPiece(SquareH3, ColorPiece(Black, Pawn))

	testFENHelper(t, expected, FENKiwipete)
}

func testMoves(t *testing.T, moves []Move, expected []string) {
	t.Logf("expected = %v", expected)
	t.Logf("actual = %v", moves)
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
	kn := ColorPiece(White, Knight)
	pos := &Position{}
	pos.PutPiece(SquareB2, kn)
	pos.PutPiece(SquareE4, kn)
	pos.PutPiece(SquareC4, ColorPiece(White, Pawn))

	moves := pos.genKnightMoves(SquareB2, kn, nil)
	expected := []string{"b2d1", "b2d3", "b2a4"}
	testMoves(t, moves, expected)

	moves = pos.genKnightMoves(SquareF4, kn, nil)
	expected = []string{"f4d3", "f4d5", "f4e6", "f4g6", "f4h5", "f4h3", "f4g2", "f4e2"}
	testMoves(t, moves, expected)
}

func TestGenRookMoves(t *testing.T) {
	rk := ColorPiece(White, Rook)
	pos := &Position{}
	pos.PutPiece(SquareB2, rk)
	pos.PutPiece(SquareF2, ColorPiece(White, King))
	pos.PutPiece(SquareB6, ColorPiece(Black, King))

	moves := pos.genRookMoves(SquareB2, rk, nil)
	expected := []string{"b2b1", "b2b3", "b2b4", "b2b5", "b2b6", "b2a2", "b2c2", "b2d2", "b2e2"}
	testMoves(t, moves, expected)
}

func TestGenKingMoves(t *testing.T) {
	// King is alone.
	kg := ColorPiece(White, Rook)
	pos := &Position{}
	pos.PutPiece(SquareA2, kg)

	moves := pos.genKingMoves(SquareA2, kg, nil)
	expected := []string{"a2a3", "a2b3", "a2b2", "a2b1", "a2a1"}
	testMoves(t, moves, expected)

	// King is surrounded by black and white pieces.
	pos.PutPiece(SquareA3, ColorPiece(White, Pawn))
	pos.PutPiece(SquareB3, ColorPiece(Black, Knight))
	pos.PutPiece(SquareB2, ColorPiece(White, Queen))

	moves = pos.genKingMoves(SquareA2, kg, nil)
	expected = []string{"a2b3", "a2b1", "a2a1"}
	testMoves(t, moves, expected)
}

func TestGenBishopMoves(t *testing.T) {
	bs := ColorPiece(White, Bishop)
	pos := &Position{}
	pos.PutPiece(SquareB1, ColorPiece(Black, Rook))
	pos.PutPiece(SquareD1, ColorPiece(White, Queen))
	pos.PutPiece(SquareE1, ColorPiece(White, King))
	pos.PutPiece(SquareG1, ColorPiece(White, Knight))
	pos.PutPiece(SquareC2, ColorPiece(White, Knight))
	pos.PutPiece(SquareF2, ColorPiece(White, Knight))
	pos.PutPiece(SquareE3, ColorPiece(White, Knight))
	pos.PutPiece(SquareF3, bs)
	pos.PutPiece(SquareD5, ColorPiece(Black, Rook))

	moves := pos.genBishopMoves(SquareF3, bs, nil)
	expected := []string{"f3e2", "f3e4", "f3d5", "f3g2", "f3h1", "f3g4", "f3h5"}
	testMoves(t, moves, expected)
}

func TestGenQueenMoves(t *testing.T) {
	qn := ColorPiece(White, Queen)
	pos := &Position{}
	pos.PutPiece(SquareB1, ColorPiece(Black, Rook))
	pos.PutPiece(SquareD1, qn)
	pos.PutPiece(SquareE1, ColorPiece(White, King))
	pos.PutPiece(SquareG1, ColorPiece(White, Knight))
	pos.PutPiece(SquareC2, ColorPiece(White, Knight))
	pos.PutPiece(SquareF2, ColorPiece(White, Knight))
	pos.PutPiece(SquareE3, ColorPiece(White, Knight))
	pos.PutPiece(SquareF3, ColorPiece(White, Bishop))
	pos.PutPiece(SquareD5, ColorPiece(Black, Rook))

	moves := pos.genQueenMoves(SquareD1, qn, nil)
	expected := []string{"d1b1", "d1c1", "d1d2", "d1d3", "d1d4", "d1d5", "d1e2"}
	testMoves(t, moves, expected)
}
