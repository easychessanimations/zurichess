package engine

import (
	"testing"
)

var (
	sanMoves = []struct {
		san  string
		move Move
	}{
		{
			"Qxf6",
			Move{
				MoveType:    Normal,
				From:        SquareF3,
				To:          SquareF6,
				Capture:     BlackKnight,
				Target:      WhiteQueen,
				SavedCastle: AnyCastle,
			},
		},
		{
			"hxg2",
			Move{
				MoveType:    Normal,
				From:        SquareH3,
				To:          SquareG2,
				Capture:     WhitePawn,
				Target:      BlackPawn,
				SavedCastle: AnyCastle,
			},
		},
		{
			"a4",
			Move{
				MoveType:    Normal,
				From:        SquareA2,
				To:          SquareA4,
				Target:      WhitePawn,
				SavedCastle: AnyCastle,
			},
		},
		{
			"bxa3e.p.",
			Move{
				MoveType:       Enpassant,
				From:           SquareB4,
				To:             SquareA3,
				Capture:        WhitePawn,
				Target:         BlackPawn,
				SavedCastle:    AnyCastle,
				SavedEnpassant: SquareA3,
			},
		},
		{
			"Qf5",
			Move{
				MoveType:    Normal,
				From:        SquareF6,
				To:          SquareF5,
				Target:      WhiteQueen,
				SavedCastle: AnyCastle,
			},
		},
		{
			"gxh1=Q",
			Move{
				MoveType:    Promotion,
				From:        SquareG2,
				To:          SquareH1,
				Capture:     WhiteRook,
				Target:      BlackQueen,
				SavedCastle: AnyCastle,
			},
		},
		{
			"Bf1",
			Move{
				MoveType:    Normal,
				From:        SquareE2,
				To:          SquareF1,
				Target:      WhiteBishop,
				SavedCastle: WhiteOOO | BlackOO | BlackOOO,
			},
		},
		{
			"exf5",
			Move{
				MoveType:    Normal,
				From:        SquareE6,
				To:          SquareF5,
				Capture:     WhiteQueen,
				Target:      BlackPawn,
				SavedCastle: WhiteOOO | BlackOO | BlackOOO,
			},
		},
	}
)

func TestSANToMovePlay(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	for i, test := range sanMoves {
		actual, err := pos.SANToMove(test.san)
		if err != nil {
			t.Fatalf("#%d %s parse error: %v", i, test.san, err)
		} else if test.move != actual {
			t.Fatalf("#%d %s expected %#v, got %#v", i, test.san, &test.move, &actual)
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

	expected := Move{
		MoveType: Normal,
		From:     SquareA2,
		To:       SquareA1,
		Target:   BlackRook,
	}
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
