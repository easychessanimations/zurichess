package engine

import (
	"testing"
)

func testFENHelper(t *testing.T, expected *Position, fen string) {
	epd, err := ParseFEN(fen)
	if err != nil {
		t.Error(err)
		return
	}

	actual := epd.Position
	for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
		epi := expected.Get(sq)
		api := actual.Get(sq)
		if epi != api {
			t.Errorf("expected %v at %v, got %v", epi, sq, api)
		}
	}
	if expected.SideToMove != actual.SideToMove {
		t.Errorf("expected to move %v, got %v",
			expected.SideToMove, actual.SideToMove)
	}
	if expected.Castle != actual.Castle {
		t.Errorf("expected Castle rights %v, got %v",
			expected.Castle, actual.Castle)
	}
	if expected.EnpassantSquare != actual.EnpassantSquare {
		t.Errorf("expected enpassant square %v, got %v",
			expected.EnpassantSquare, actual.EnpassantSquare)
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

	expected.SideToMove = White
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

	expected.SideToMove = White
	expected.Castle = AnyCastle
	testFENHelper(t, expected, FENKiwipete)
}

func TestEPDParser(t *testing.T) {
	// An EPD taken from http://www.stmintz.com/ccc/index.php?id=20631
	line := "rnb2r1k/pp2p2p/2pp2p1/q2P1p2/8/1Pb2NP1/PB2PPBP/R2Q1RK1 w - - bm Qd2 Qe1; id \"BK.14\";"
	epd, err := ParseEPD(line)
	if err != nil {
		t.Fatal(err)
	}

	// Verify id.
	expecteId := "\"BK.14\""
	if expecteId != epd.Id {
		t.Fatalf("expected id %s, got %s", expecteId, epd.Id)
	}

	// Verify bm.
	expectedBestMove := []Move{
		{
			MoveType: Normal,
			Target:   WhiteQueen,
			From:     SquareD1,
			To:       SquareD2,
		},
		{
			MoveType: Normal,
			Target:   WhiteQueen,
			From:     SquareD1,
			To:       SquareE1,
		},
	}
	if len(expectedBestMove) != len(epd.BestMove) {
		t.Fatalf("expected 2 best moves, got %d", len(epd.BestMove))
	}
	for i, bm := range expectedBestMove {
		if bm != epd.BestMove[i] {
			t.Errorf("#%d expected best move 0 %v, got %v", i, bm, epd.BestMove[i])
		}
	}
}
