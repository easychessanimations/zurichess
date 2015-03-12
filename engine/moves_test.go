package engine

import (
	"testing"
)

var (
	sanMoves = []struct {
		san  string
		move Move
	}{
		{"Qxf6", MakeMove(Normal, SquareF3, SquareF6, BlackKnight, WhiteQueen)},
		{"hxg2", MakeMove(Normal, SquareH3, SquareG2, WhitePawn, BlackPawn)},
		{"a4", MakeMove(Normal, SquareA2, SquareA4, NoPiece, WhitePawn)},
		{"bxa3e.p.", MakeMove(Enpassant, SquareB4, SquareA3, WhitePawn, BlackPawn)},
		{"Qf5", MakeMove(Normal, SquareF6, SquareF5, NoPiece, WhiteQueen)},
		{"gxh1=Q", MakeMove(Promotion, SquareG2, SquareH1, WhiteRook, BlackQueen)},
		{"Bf1", MakeMove(Normal, SquareE2, SquareF1, NoPiece, WhiteBishop)},
		{"exf5", MakeMove(Normal, SquareE6, SquareF5, WhiteQueen, BlackPawn)},
	}
)

func TestSANToMovePlay(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	for i, test := range sanMoves {
		actual, err := pos.SANToMove(test.san)
		if err != nil {
			t.Fatalf("#%d %s parse error: %v", i, test.san, err)
		} else if test.move != actual {
			t.Fatalf("#%d %s expected %v (%s), got %v (%s)",
				i, test.san, test.move, &test.move, actual, &actual)
		}
		pos.DoMove(actual)
	}
}

func TestSANToMoveFixed(t *testing.T) {
	pos, _ := PositionFromFEN("2r3k1/6pp/4pp2/3bp3/1Pq5/3R1P2/r1PQ2PP/1K1RN3 b - - 0 1")
	actual, err := pos.SANToMove("Ra1+")
	if err != nil {
		t.Fatal("could not parse move:", err)
	}

	expected := MakeMove(Normal, SquareA2, SquareA1, NoPiece, BlackRook)
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
