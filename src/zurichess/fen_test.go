package main

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
