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
	expected.PutPiece(SquareA1, WhiteRook)
	expected.PutPiece(SquareB1, WhiteKnight)
	expected.PutPiece(SquareC1, WhiteBishop)
	expected.PutPiece(SquareD1, WhiteQueen)
	expected.PutPiece(SquareE1, WhiteKing)
	expected.PutPiece(SquareF1, WhiteBishop)
	expected.PutPiece(SquareG1, WhiteKnight)
	expected.PutPiece(SquareH1, WhiteRook)

	expected.PutPiece(SquareA8, BlackRook)
	expected.PutPiece(SquareB8, BlackKnight)
	expected.PutPiece(SquareC8, BlackBishop)
	expected.PutPiece(SquareD8, BlackQueen)
	expected.PutPiece(SquareE8, BlackKing)
	expected.PutPiece(SquareF8, BlackBishop)
	expected.PutPiece(SquareG8, BlackKnight)
	expected.PutPiece(SquareH8, BlackRook)

	for f := 0; f < 8; f++ {
		expected.PutPiece(RankFile(1, f), WhitePawn)
		expected.PutPiece(RankFile(6, f), BlackPawn)
	}

	testFENHelper(t, expected, FENStartPos)
}

func TestFENKiwipete(t *testing.T) {
	expected := &Position{}
	expected.PutPiece(SquareA1, WhiteRook)
	expected.PutPiece(SquareC3, WhiteKnight)
	expected.PutPiece(SquareD2, WhiteBishop)
	expected.PutPiece(SquareF3, WhiteQueen)
	expected.PutPiece(SquareE1, WhiteKing)
	expected.PutPiece(SquareE2, WhiteBishop)
	expected.PutPiece(SquareE5, WhiteKnight)
	expected.PutPiece(SquareH1, WhiteRook)

	expected.PutPiece(SquareA8, BlackRook)
	expected.PutPiece(SquareB6, BlackKnight)
	expected.PutPiece(SquareA6, BlackBishop)
	expected.PutPiece(SquareE7, BlackQueen)
	expected.PutPiece(SquareE8, BlackKing)
	expected.PutPiece(SquareG7, BlackBishop)
	expected.PutPiece(SquareF6, BlackKnight)
	expected.PutPiece(SquareH8, BlackRook)

	expected.PutPiece(SquareA2, WhitePawn)
	expected.PutPiece(SquareB2, WhitePawn)
	expected.PutPiece(SquareC2, WhitePawn)
	expected.PutPiece(SquareD5, WhitePawn)
	expected.PutPiece(SquareE4, WhitePawn)
	expected.PutPiece(SquareF2, WhitePawn)
	expected.PutPiece(SquareG2, WhitePawn)
	expected.PutPiece(SquareH2, WhitePawn)

	expected.PutPiece(SquareA7, BlackPawn)
	expected.PutPiece(SquareB4, BlackPawn)
	expected.PutPiece(SquareC7, BlackPawn)
	expected.PutPiece(SquareD7, BlackPawn)
	expected.PutPiece(SquareE6, BlackPawn)
	expected.PutPiece(SquareF7, BlackPawn)
	expected.PutPiece(SquareG6, BlackPawn)
	expected.PutPiece(SquareH3, BlackPawn)

	testFENHelper(t, expected, FENKiwipete)
}
