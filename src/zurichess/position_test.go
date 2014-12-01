package main

import (
	"testing"
)

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
