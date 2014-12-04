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

	pos.PutPiece(SquareA3, WhitePawn)
	pi = pos.GetPiece(SquareA3)
	checkPiece(t, pi, White, Pawn)

	pos.PutPiece(SquareH7, BlackKing)
	pi = pos.GetPiece(SquareH7)
	checkPiece(t, pi, Black, King)
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
	kn := WhiteKnight
	pos := &Position{}
	pos.PutPiece(SquareB2, kn)
	pos.PutPiece(SquareE4, kn)
	pos.PutPiece(SquareC4, WhitePawn)

	moves := pos.genKnightMoves(SquareB2, kn, nil)
	expected := []string{"b2d1", "b2d3", "b2a4"}
	testMoves(t, moves, expected)

	moves = pos.genKnightMoves(SquareF4, kn, nil)
	expected = []string{"f4d3", "f4d5", "f4e6", "f4g6", "f4h5", "f4h3", "f4g2", "f4e2"}
	testMoves(t, moves, expected)
}

func TestGenRookMoves(t *testing.T) {
	rk := WhiteRook
	pos := &Position{}
	pos.PutPiece(SquareB2, rk)
	pos.PutPiece(SquareF2, WhiteKing)
	pos.PutPiece(SquareB6, BlackKing)

	moves := pos.genRookMoves(SquareB2, rk, nil)
	expected := []string{"b2b1", "b2b3", "b2b4", "b2b5", "b2b6", "b2a2", "b2c2", "b2d2", "b2e2"}
	testMoves(t, moves, expected)
}

func TestGenKingMoves(t *testing.T) {
	// King is alone.
	kg := WhiteRook
	pos := &Position{}
	pos.PutPiece(SquareA2, kg)

	moves := pos.genKingMoves(SquareA2, kg, nil)
	expected := []string{"a2a3", "a2b3", "a2b2", "a2b1", "a2a1"}
	testMoves(t, moves, expected)

	// King is surrounded by black and white pieces.
	pos.PutPiece(SquareA3, WhitePawn)
	pos.PutPiece(SquareB3, BlackKnight)
	pos.PutPiece(SquareB2, WhiteQueen)

	moves = pos.genKingMoves(SquareA2, kg, nil)
	expected = []string{"a2b3", "a2b1", "a2a1"}
	testMoves(t, moves, expected)
}

func TestCastle(t *testing.T) {
	pos := &Position{}
	pos.PutPiece(SquareD2, WhitePawn)
	pos.PutPiece(SquareE2, WhitePawn)
	pos.PutPiece(SquareF2, WhitePawn)
	pos.PutPiece(SquareE1, WhiteKing)
	pos.PutPiece(SquareA1, WhiteRook)
	pos.PutPiece(SquareA8, WhiteRook)

	type testData struct {
		castle   Castle   // castle rights
		expected []string // expected moves
	}

	// Simple.
	data := []testData{
		// No castle rights.
		{NoCastle, []string{"e1d1", "e1f1"}},
		// Castle rights for black.
		{BlackOO | BlackOOO, []string{"e1d1", "e1f1"}},
		// Castle on king side.
		{WhiteOO, []string{"e1d1", "e1f1", "e1g1"}},
		// Castle on queen side.
		{WhiteOOO, []string{"e1d1", "e1f1", "e1c1"}},
		// Castle on both sides.
		{WhiteOO | WhiteOOO, []string{"e1d1", "e1f1", "e1g1", "e1c1"}},
	}
	for _, d := range data {
		pos.castle = d.castle
		moves := pos.genKingMoves(SquareE1, WhiteKing, nil)
		testMoves(t, moves, d.expected)
	}

	// Put a piece to block castling on OOO
	pos.PutPiece(SquareC1, WhiteBishop)
	data = []testData{
		// No castle rights.
		{NoCastle, []string{"e1d1", "e1f1"}},
		// Castle rights for black.
		{BlackOO | BlackOOO, []string{"e1d1", "e1f1"}},
		// Castle on king side.
		{WhiteOO, []string{"e1d1", "e1f1", "e1g1"}},
		// Castle on queen side.
		{WhiteOOO, []string{"e1d1", "e1f1"}},
		// Castle on both sides.
		{WhiteOO | WhiteOOO, []string{"e1d1", "e1f1", "e1g1"}},
	}
	for _, d := range data {
		pos.castle = d.castle
		moves := pos.genKingMoves(SquareE1, WhiteKing, nil)
		testMoves(t, moves, d.expected)
	}
}

func TestGenBishopMoves(t *testing.T) {
	pos := &Position{}
	pos.PutPiece(SquareB1, BlackRook)
	pos.PutPiece(SquareD1, WhiteQueen)
	pos.PutPiece(SquareE1, WhiteKing)
	pos.PutPiece(SquareG1, WhiteKnight)
	pos.PutPiece(SquareC2, WhiteKnight)
	pos.PutPiece(SquareF2, WhiteKnight)
	pos.PutPiece(SquareE3, WhiteKnight)
	pos.PutPiece(SquareF3, WhiteBishop)
	pos.PutPiece(SquareD5, BlackRook)

	moves := pos.genBishopMoves(SquareF3, WhiteBishop, nil)
	expected := []string{"f3e2", "f3e4", "f3d5", "f3g2", "f3h1", "f3g4", "f3h5"}
	testMoves(t, moves, expected)
}

func TestGenQueenMoves(t *testing.T) {
	pos := &Position{}
	pos.PutPiece(SquareB1, BlackRook)
	pos.PutPiece(SquareD1, WhiteQueen)
	pos.PutPiece(SquareE1, WhiteKing)
	pos.PutPiece(SquareG1, WhiteKnight)
	pos.PutPiece(SquareC2, WhiteKnight)
	pos.PutPiece(SquareF2, WhiteKnight)
	pos.PutPiece(SquareE3, WhiteKnight)
	pos.PutPiece(SquareF3, WhiteBishop)
	pos.PutPiece(SquareD5, BlackRook)

	moves := pos.genQueenMoves(SquareD1, WhiteQueen, nil)
	expected := []string{"d1b1", "d1c1", "d1d2", "d1d3", "d1d4", "d1d5", "d1e2"}
	testMoves(t, moves, expected)
}
