package engine

import (
	"testing"
)

func testFENHelper(t *testing.T, expected *Position, fen string) {
	actual, err := PositionFromFEN(fen)
	if err != nil {
		t.Error(err)
		return
	}

	for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
		epi := expected.Get(sq)
		api := actual.Get(sq)
		if epi != api {
			t.Errorf("expected %v at %v, got %v", epi, sq, api)
		}
	}

	if expected.ToMove != actual.ToMove {
		t.Errorf("expected to move %v, got %v",
			expected.ToMove, actual.ToMove)
	}

	if expected.Castle != actual.Castle {
		t.Errorf("expected Castle rights %v, got %v",
			expected.Castle, actual.Castle)
	}
}

func TestFENStartPosition(t *testing.T) {
	expected := &Position{}
	expected.Put(SquareA1, WhiteRook)
	expected.Put(SquareB1, WhiteKnight)
	expected.Put(SquareC1, WhiteBishop)
	expected.Put(SquareD1, WhiteQueen)
	expected.Put(SquareE1, WhiteKing)
	expected.Put(SquareF1, WhiteBishop)
	expected.Put(SquareG1, WhiteKnight)
	expected.Put(SquareH1, WhiteRook)

	expected.Put(SquareA8, BlackRook)
	expected.Put(SquareB8, BlackKnight)
	expected.Put(SquareC8, BlackBishop)
	expected.Put(SquareD8, BlackQueen)
	expected.Put(SquareE8, BlackKing)
	expected.Put(SquareF8, BlackBishop)
	expected.Put(SquareG8, BlackKnight)
	expected.Put(SquareH8, BlackRook)

	for f := 0; f < 8; f++ {
		expected.Put(RankFile(1, f), WhitePawn)
		expected.Put(RankFile(6, f), BlackPawn)
	}

	expected.ToMove = White
	expected.Castle = AnyCastle
	testFENHelper(t, expected, FENStartPos)
}

func TestFENKiwipete(t *testing.T) {
	expected := &Position{}
	expected.Put(SquareA1, WhiteRook)
	expected.Put(SquareC3, WhiteKnight)
	expected.Put(SquareD2, WhiteBishop)
	expected.Put(SquareF3, WhiteQueen)
	expected.Put(SquareE1, WhiteKing)
	expected.Put(SquareE2, WhiteBishop)
	expected.Put(SquareE5, WhiteKnight)
	expected.Put(SquareH1, WhiteRook)

	expected.Put(SquareA8, BlackRook)
	expected.Put(SquareB6, BlackKnight)
	expected.Put(SquareA6, BlackBishop)
	expected.Put(SquareE7, BlackQueen)
	expected.Put(SquareE8, BlackKing)
	expected.Put(SquareG7, BlackBishop)
	expected.Put(SquareF6, BlackKnight)
	expected.Put(SquareH8, BlackRook)

	expected.Put(SquareA2, WhitePawn)
	expected.Put(SquareB2, WhitePawn)
	expected.Put(SquareC2, WhitePawn)
	expected.Put(SquareD5, WhitePawn)
	expected.Put(SquareE4, WhitePawn)
	expected.Put(SquareF2, WhitePawn)
	expected.Put(SquareG2, WhitePawn)
	expected.Put(SquareH2, WhitePawn)

	expected.Put(SquareA7, BlackPawn)
	expected.Put(SquareB4, BlackPawn)
	expected.Put(SquareC7, BlackPawn)
	expected.Put(SquareD7, BlackPawn)
	expected.Put(SquareE6, BlackPawn)
	expected.Put(SquareF7, BlackPawn)
	expected.Put(SquareG6, BlackPawn)
	expected.Put(SquareH3, BlackPawn)

	expected.ToMove = White
	expected.Castle = AnyCastle
	testFENHelper(t, expected, FENKiwipete)
}
