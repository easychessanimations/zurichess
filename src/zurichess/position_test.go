package main

import (
	"testing"
)

func TestPutGet(t *testing.T) {
	var pi Piece
	pos := &Position{}

	pi = pos.Get(SquareA3)
	checkPiece(t, pi, NoColor, NoFigure)

	pos.Put(SquareA3, WhitePawn)
	pi = pos.Get(SquareA3)
	checkPiece(t, pi, White, Pawn)

	pos.Put(SquareH7, BlackKing)
	pi = pos.Get(SquareH7)
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
	pos := &Position{toMove: White}
	pos.Put(SquareB2, WhiteKnight)
	pos.Put(SquareE4, WhiteKnight)
	pos.Put(SquareC4, WhitePawn)

	moves := pos.genKnightMoves(SquareB2, nil)
	expected := []string{"b2d1", "b2d3", "b2a4"}
	testMoves(t, moves, expected)

	moves = pos.genKnightMoves(SquareF4, nil)
	expected = []string{"f4d3", "f4d5", "f4e6", "f4g6", "f4h5", "f4h3", "f4g2", "f4e2"}
	testMoves(t, moves, expected)
}

func TestGenRookMoves(t *testing.T) {
	pos := &Position{toMove: White}
	pos.Put(SquareB2, WhiteRook)
	pos.Put(SquareF2, WhiteKing)
	pos.Put(SquareB6, BlackKing)

	moves := pos.genRookMoves(SquareB2, nil)
	expected := []string{"b2b1", "b2b3", "b2b4", "b2b5", "b2b6", "b2a2", "b2c2", "b2d2", "b2e2"}
	testMoves(t, moves, expected)
}

func TestGenKingMoves(t *testing.T) {
	// King is alone.
	pos := &Position{toMove: White}
	pos.Put(SquareA2, WhiteKing)

	moves := pos.genKingMoves(SquareA2, nil)
	expected := []string{"a2a3", "a2b3", "a2b2", "a2b1", "a2a1"}
	testMoves(t, moves, expected)

	// King is surrounded by black and white pieces.
	pos.Put(SquareA3, WhitePawn)
	pos.Put(SquareB3, BlackKnight)
	pos.Put(SquareB2, WhiteQueen)

	moves = pos.genKingMoves(SquareA2, nil)
	expected = []string{"a2b3", "a2b1", "a2a1"}
	testMoves(t, moves, expected)
}

type castleTestData struct {
	castle   Castle   // castle rights, 255 to ignore
	expected []string // expected moves
}

func testCastleHelper(t *testing.T, pos *Position, data []castleTestData) {
	if pos.Get(SquareE1) != WhiteKing {
		t.Errorf("expected %v on %v, got %v",
			WhiteKing, SquareE1, pos.Get(SquareE1))
		return
	}

	for _, d := range data {
		if d.castle != 255 {
			pos.castle = d.castle
		}
		moves := pos.genKingMoves(SquareE1, nil)
		testMoves(t, moves, d.expected)
	}
}

func TestCastle(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)

	// Simple.
	testCastleHelper(t, pos, []castleTestData{
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
	})

	// Put a piece to block castling on OOO
	pos.Put(SquareC1, WhiteBishop)
	testCastleHelper(t, pos, []castleTestData{
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
	})
}

func TestCastleAfterUnrelatedMove(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	pos.toMove = Black

	// Move bishop which doesn't change castle rights.
	m1 := Move{
		From:      SquareF7,
		To:        SquareF6,
		OldCastle: pos.castle,
	}

	data := []castleTestData{
		// Castle on both sides.
		{255, []string{"e1d1", "e1f1", "e1g1", "e1c1"}},
	}

	pos.DoMove(m1)
	testCastleHelper(t, pos, data)
}

func testPiece(t *testing.T, pos *Position, sq Square, pi Piece) {
	if pos.Get(sq) != pi {
		t.Errorf("expected %v at %v, got %v",
			pi, sq, pos.Get(sq))
	}
}

func TestCastleMovesPieces(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)

	// White
	pos.toMove = White
	m1 := Move{
		MoveType:  Castling,
		From:      SquareE1,
		To:        SquareC1,
		OldCastle: pos.castle,
	}

	pos.DoMove(m1)
	testPiece(t, pos, SquareA1, NoPiece)
	testPiece(t, pos, SquareC1, WhiteKing)
	testPiece(t, pos, SquareD1, WhiteRook)
	testPiece(t, pos, SquareE1, NoPiece)

	pos.UndoMove(m1)
	testPiece(t, pos, SquareA1, WhiteRook)
	testPiece(t, pos, SquareC1, NoPiece)
	testPiece(t, pos, SquareD1, NoPiece)
	testPiece(t, pos, SquareE1, WhiteKing)

	// Black
	pos.toMove = Black
	m2 := Move{
		MoveType:  Castling,
		From:      SquareE8,
		To:        SquareC8,
		OldCastle: pos.castle,
	}

	pos.DoMove(m2)
	testPiece(t, pos, SquareA8, NoPiece)
	testPiece(t, pos, SquareC8, BlackKing)
	testPiece(t, pos, SquareD8, BlackRook)
	testPiece(t, pos, SquareE8, NoPiece)

	pos.UndoMove(m2)
	testPiece(t, pos, SquareA8, BlackRook)
	testPiece(t, pos, SquareC8, NoPiece)
	testPiece(t, pos, SquareD8, NoPiece)
	testPiece(t, pos, SquareE8, BlackKing)
}

func TestCastleRightsAreUpdated(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	pos.castle = WhiteOOO

	good := []castleTestData{
		{255, []string{"e1d1", "e1f1", "e1c1"}},
	}
	fail := []castleTestData{
		{255, []string{"e1d1", "e1f1"}},
	}

	// Check that king can castle queen side.
	testCastleHelper(t, pos, good)

	// Move rook.
	m1 := pos.ParseMove("a1a4")
	pos.DoMove(m1)
	b1 := pos.ParseMove("a8a5")
	pos.DoMove(b1)
	testCastleHelper(t, pos, fail)

	m2 := pos.ParseMove("a4a1")
	pos.DoMove(m2)
	b2 := pos.ParseMove("a5a8")
	pos.DoMove(b2)
	testCastleHelper(t, pos, fail)

	// Undo rook's moves.
	pos.UndoMove(b2)
	pos.UndoMove(m2)
	testCastleHelper(t, pos, fail)

	pos.UndoMove(b1)
	pos.UndoMove(m1)
	testCastleHelper(t, pos, good)

	// Move king.
	m3 := pos.ParseMove("e1d1")
	pos.DoMove(m3)
	pos.DoMove(b1)
	moves := pos.genKingMoves(SquareD1, nil)
	testMoves(t, moves, []string{"d1c1", "d1c2", "d1e1"})

	m4 := pos.ParseMove("d1e1")
	pos.DoMove(m4)
	pos.DoMove(b2)
	testCastleHelper(t, pos, fail)

	// Undo king's move.
	pos.UndoMove(b2)
	pos.UndoMove(m4)
	moves = pos.genKingMoves(SquareD1, nil)
	testMoves(t, moves, []string{"d1c1", "d1c2", "d1e1"})

	pos.UndoMove(b1)
	pos.UndoMove(m3)
	testCastleHelper(t, pos, good)
}

func TestGenBishopMoves(t *testing.T) {
	pos := &Position{toMove: White}
	pos.Put(SquareB1, BlackRook)
	pos.Put(SquareD1, WhiteQueen)
	pos.Put(SquareE1, WhiteKing)
	pos.Put(SquareG1, WhiteKnight)
	pos.Put(SquareC2, WhiteKnight)
	pos.Put(SquareF2, WhiteKnight)
	pos.Put(SquareE3, WhiteKnight)
	pos.Put(SquareF3, WhiteBishop)
	pos.Put(SquareD5, BlackRook)

	moves := pos.genBishopMoves(SquareF3, nil)
	expected := []string{"f3e2", "f3e4", "f3d5", "f3g2", "f3h1", "f3g4", "f3h5"}
	testMoves(t, moves, expected)
}

func TestGenQueenMoves(t *testing.T) {
	pos := &Position{toMove: White}
	pos.Put(SquareB1, BlackRook)
	pos.Put(SquareD1, WhiteQueen)
	pos.Put(SquareE1, WhiteKing)
	pos.Put(SquareG1, WhiteKnight)
	pos.Put(SquareC2, WhiteKnight)
	pos.Put(SquareF2, WhiteKnight)
	pos.Put(SquareE3, WhiteKnight)
	pos.Put(SquareF3, WhiteBishop)
	pos.Put(SquareD5, BlackRook)

	moves := pos.genQueenMoves(SquareD1, nil)
	expected := []string{"d1b1", "d1c1", "d1d2", "d1d3", "d1d4", "d1d5", "d1e2"}
	testMoves(t, moves, expected)
}

func TestGenPawnMoves(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)

	moves := pos.genPawnMoves(SquareE5, nil)
	expected := []string{"e5e6"}
	testMoves(t, moves, expected)

	moves = pos.genPawnMoves(SquareE2, nil)
	expected = []string{"e2e3", "e2e4", "e2f3"}
	testMoves(t, moves, expected)
}

func TestPawnEnpassant(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	pos.toMove = Black

	m1 := pos.ParseMove("d7d5")
	pos.DoMove(m1)
	moves := pos.genPawnMoves(SquareE5, nil)
	expected := []string{"e5e6", "e5d6"}
	testMoves(t, moves, expected)

	m2 := pos.ParseMove("e2e3")
	pos.DoMove(m2)
	m3 := pos.ParseMove("f7f5")
	pos.DoMove(m3)
	moves = pos.genPawnMoves(SquareE5, nil)
	expected = []string{"e5e6", "e5f6"}
	testMoves(t, moves, expected)

	m4 := pos.ParseMove("e3e4")
	pos.DoMove(m4)
	m5 := pos.ParseMove("f3f4")
	pos.DoMove(m5)
	moves = pos.genPawnMoves(SquareE5, nil)
	expected = []string{"e5e6"}
	testMoves(t, moves, expected)
}
