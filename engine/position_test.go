package engine

import (
	"log"
	"testing"
)

var (
	_          = log.Println
	testBoard1 = "r3k2r/3ppp2/1BB3B1/pp2P1pp/PP4PP/5b2/3PPP2/R3K2R w KQkq - 0 1"
	testBoard2 = "3k4/8/8/p1P2p2/PpP1pP2/pPPpP3/2P2pp1/3K3R w - - 0 1"
)

// testEngine is an simple engine to simplify move testing.
type testEngine struct {
	T     *testing.T
	Pos   *Position
	moves []Move
}

// Move does uci move (e.g. a1h8).
// If m == "", then it does the null move.
func (te *testEngine) Move(m string) {
	move := Move(0)
	if m != "" {
		move = te.Pos.UCIToMove(m)
	}
	if te.Pos.SideToMove == move.Capture().Color() {
		te.T.Fatalf("%v cannot capture its own color (move %v)",
			te.Pos.SideToMove, move)
	}
	te.moves = append(te.moves, move)
	te.Pos.DoMove(move)
}

func (te *testEngine) Undo() {
	l := len(te.moves) - 1
	te.Pos.UndoMove(te.moves[l])
	te.moves = te.moves[:l]
}

func (te *testEngine) Attacked(sq Square, co Color, is bool) {
	if is && !te.Pos.IsAttackedBy(sq, co) {
		te.T.Errorf("expected %v to be attacked by %v", sq, co)
	}
	if !is && te.Pos.IsAttackedBy(sq, co) {
		te.T.Errorf("expected %v not to be attacked by %v", sq, co)
	}
}

func (te *testEngine) Piece(sq Square, expected Piece) {
	if te.Pos.Get(sq) != expected {
		te.T.Errorf("expected %v at %v, got %v", expected, sq, te.Pos.Get(sq))
	}
}

// Filter pawn moves not starting at sq.
func (te *testEngine) Pawn(sq Square, expected []string) {
	end := 0
	var actual []Move
	te.Pos.GenerateFigureMoves(Pawn, All, &actual)
	for i := range actual {
		if actual[i].From() == sq {
			actual[end] = actual[i]
			end++
		}
	}
	testMoves(te.T, actual[:end], expected)
}

func (te *testEngine) Knight(expected []string) {
	var actual []Move
	te.Pos.genKnightMoves(All, &actual)
	testMoves(te.T, actual, expected)
}

func (te *testEngine) Bishop(expected []string) {
	var actual []Move
	te.Pos.genBishopMoves(Bishop, All, &actual)
	testMoves(te.T, actual, expected)
}

func (te *testEngine) Rook(expected []string) {
	var actual []Move
	te.Pos.genRookMoves(Rook, All, &actual)
	testMoves(te.T, actual, expected)
}

func (te *testEngine) Queen(expected []string) {
	var actual []Move
	te.Pos.genBishopMoves(Queen, All, &actual)
	te.Pos.genRookMoves(Queen, All, &actual)
	testMoves(te.T, actual, expected)
}

func (te *testEngine) King(expected []string) {
	var actual []Move
	te.Pos.genKingMovesNear(All, &actual)
	te.Pos.genKingCastles(All, &actual)
	testMoves(te.T, actual, expected)
}

func testMoves(t *testing.T, moves []Move, expected []string) {
	seen := make(map[string]bool)
	for _, e := range expected {
		seen[e] = false
	}
	for _, mo := range moves {
		str := mo.UCI()
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

func TestPutGetRemove(t *testing.T) {
	pos := NewPosition()
	te := &testEngine{T: t, Pos: pos}

	te.Piece(SquareA3, NoPiece)

	pos.Put(SquareA3, WhitePawn)
	te.Piece(SquareA3, WhitePawn)
	pos.Remove(SquareA3, WhitePawn)
	te.Piece(SquareA3, NoPiece)

	pos.Put(SquareH7, BlackKing)
	te.Piece(SquareH7, BlackKing)
	pos.Remove(SquareH7, BlackKing)
	te.Piece(SquareH7, NoPiece)
}

func TestKnightMoves(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	pos.Put(SquareB2, WhiteKnight)
	pos.Put(SquareF4, WhiteKnight)
	pos.Put(SquareC4, WhitePawn)

	te := &testEngine{T: t, Pos: pos}
	te.Knight([]string{"b2d1", "b2d3", "b2a4", "f4d3", "f4d5", "f4e6", "f4g6", "f4h5", "f4h3", "f4g2", "f4e2"})
}

func TestRookMoves(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	pos.Put(SquareB2, WhiteRook)
	pos.Put(SquareF2, WhiteKing)
	pos.Put(SquareB6, BlackKing)

	te := &testEngine{T: t, Pos: pos}
	te.Rook([]string{"b2b1", "b2b3", "b2b4", "b2b5", "b2b6", "b2a2", "b2c2", "b2d2", "b2e2"})
}

func TestKingMoves1(t *testing.T) {
	// King is alone.
	pos := NewPosition()
	pos.SetSideToMove(White)
	te := &testEngine{T: t, Pos: pos}

	pos.Put(SquareA2, WhiteKing)
	te.King([]string{"a2a3", "a2b3", "a2b2", "a2b1", "a2a1"})

	// King is surrounded by black and white pieces.
	pos.Put(SquareA3, WhitePawn)
	pos.Put(SquareB3, BlackPawn)
	pos.Put(SquareB2, WhiteQueen)
	te.King([]string{"a2b3", "a2b1", "a2a1"})
}

func TestKingMoves2(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	te := &testEngine{T: t, Pos: pos}
	te.Move("f2f4")
	te.Move("h7h5")
	te.Move("e1f2")
	te.Move("h5h4")

	te.King([]string{"f2e1", "f2e3", "f2f3", "f2g3"})
}

type CastleTestData struct {
	Castle   Castle   // Castle rights, 255 to ignore
	expected []string // expected moves
}

func testCastleHelper(t *testing.T, pos *Position, data []CastleTestData) {
	if pos.Get(SquareE1) != WhiteKing {
		t.Errorf("expected %v on %v, got %v",
			WhiteKing, SquareE1, pos.Get(SquareE1))
		return
	}

	for _, d := range data {
		if d.Castle != 255 {
			pos.SetCastlingAbility(d.Castle)
		}
		te := &testEngine{T: t, Pos: pos}
		te.King(d.expected)
	}
}

func TestCastle(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)

	// Simple.
	testCastleHelper(t, pos, []CastleTestData{
		// No Castle rights.
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
	testCastleHelper(t, pos, []CastleTestData{
		// No Castle rights.
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

	// Black bishop attacks one of the squares.
	pos.Remove(SquareC1, WhiteBishop)
	pos.Put(SquareA3, BlackBishop)
	testCastleHelper(t, pos, []CastleTestData{
		// No Castle rights.
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

func TestKingCannotCastleWhenUnderAttack(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	te := &testEngine{T: t, Pos: pos}

	te.Move("f3f5")
	// Black King can castle both sides.
	te.King([]string{"e8d8", "e8f8", "e8g8", "e8c8"})
	te.Move("d7d6")
	te.Move("e2b5")
	// Bishop attacks Black King.
	te.King([]string{"e8d8", "e8f8", "e8d7"})
}

func TestCastleMovesPieces(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	te := &testEngine{T: t, Pos: pos}

	// White
	pos.SideToMove = White
	te.Move("e1c1")
	te.Piece(SquareA1, NoPiece)
	te.Piece(SquareC1, WhiteKing)
	te.Piece(SquareD1, WhiteRook)
	te.Piece(SquareE1, NoPiece)

	te.Undo()
	te.Piece(SquareA1, WhiteRook)
	te.Piece(SquareC1, NoPiece)
	te.Piece(SquareD1, NoPiece)
	te.Piece(SquareE1, WhiteKing)

	// Black
	pos.SideToMove = Black
	te.Move("e8c8")
	te.Piece(SquareA8, NoPiece)
	te.Piece(SquareC8, BlackKing)
	te.Piece(SquareD8, BlackRook)
	te.Piece(SquareE8, NoPiece)

	te.Undo()
	te.Piece(SquareA8, BlackRook)
	te.Piece(SquareC8, NoPiece)
	te.Piece(SquareD8, NoPiece)
	te.Piece(SquareE8, BlackKing)
}

func TestCastleRightsAreUpdated(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	pos.SetCastlingAbility(WhiteOOO)
	te := &testEngine{T: t, Pos: pos}

	good := []CastleTestData{
		{255, []string{"e1d1", "e1f1", "e1c1"}},
	}
	fail := []CastleTestData{
		{255, []string{"e1d1", "e1f1"}},
	}

	// Check that king can Castle queen side.
	testCastleHelper(t, pos, good)

	// Move rook.
	te.Move("a1a2")
	te.Move("a8a7")
	testCastleHelper(t, pos, fail)

	te.Move("a2a1")
	te.Move("a7a8")
	testCastleHelper(t, pos, fail)

	// Undo rook's moves.
	te.Undo()
	te.Undo()
	testCastleHelper(t, pos, fail)

	te.Undo()
	te.Undo()
	testCastleHelper(t, pos, good)

	// Move king.
	te.Move("e1d1")
	te.Move("a8a7")
	te.King([]string{"d1c1", "d1c2", "d1e1"})

	te.Move("d1e1")
	te.Move("a7a8")
	testCastleHelper(t, pos, fail)

	// Undo king's move.
	te.Undo()
	te.Undo()
	te.King([]string{"d1c1", "d1c2", "d1e1"})

	te.Undo()
	te.Undo()
	testCastleHelper(t, pos, good)
}

func TestCannotCastleAfterRookCapture(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	te := &testEngine{T: t, Pos: pos}

	te.Move("f3f5")
	te.Move("h3g2")
	te.Move("a1b1")
	te.Move("g2h1N") // Pawn takes white rook.
	te.King([]string{"e1d1", "e1f1"})
}

func TestgenBishopMoves(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	pos.Put(SquareB1, BlackRook)
	pos.Put(SquareD1, WhiteQueen)
	pos.Put(SquareE1, WhiteKing)
	pos.Put(SquareG1, WhiteKnight)
	pos.Put(SquareC2, WhiteKnight)
	pos.Put(SquareF2, WhiteKnight)
	pos.Put(SquareE3, WhiteKnight)
	pos.Put(SquareF3, WhiteBishop)
	pos.Put(SquareD5, BlackRook)

	te := &testEngine{T: t, Pos: pos}
	te.Bishop([]string{"f3e2", "f3e4", "f3d5", "f3g2", "f3h1", "f3g4", "f3h5"})
}

func TestgenQueenMoves(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	pos.Put(SquareB1, BlackRook)
	pos.Put(SquareD1, WhiteQueen)
	pos.Put(SquareE1, WhiteKing)
	pos.Put(SquareG1, WhiteKnight)
	pos.Put(SquareC2, WhiteKnight)
	pos.Put(SquareF2, WhiteKnight)
	pos.Put(SquareE3, WhiteKnight)
	pos.Put(SquareF3, WhiteBishop)
	pos.Put(SquareD5, BlackRook)

	te := &testEngine{T: t, Pos: pos}
	te.Queen([]string{"d1b1", "d1c1", "d1d2", "d1d3", "d1d4", "d1d5", "d1e2"})
}

func TestgenPawnAdvanceMoves(t *testing.T) {
	var moves []Move
	pos, _ := PositionFromFEN(testBoard1)

	pos.SideToMove = White
	pos.genPawnAdvanceMoves(All, &moves)
	testMoves(t, moves, []string{"d2d3", "e2e3", "e5e6"})

	pos.SideToMove = Black
	pos.genPawnAdvanceMoves(All, &moves)
	testMoves(t, moves, []string{"d7d6", "e7e6", "f7f6"})
}

func TestgenPawnDoubleAdvanceMoves(t *testing.T) {
	var moves []Move
	pos, _ := PositionFromFEN(testBoard1)

	pos.SideToMove = White
	pos.genPawnDoubleAdvanceMoves(All, &moves)
	testMoves(t, moves, []string{"d2d4", "e2e4"})

	pos.SideToMove = Black
	pos.genPawnDoubleAdvanceMoves(All, &moves)
	testMoves(t, moves, []string{"d7d5", "f7f5"})
}

func TestGenPawnAttackMoves1(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)

	var moves []Move
	pos.SideToMove = White
	pos.genPawnAttackMoves(All, &moves)
	testMoves(t, moves, []string{"e2f3", "a4b5", "b4a5", "g4h5", "h4g5"})

	moves = moves[:0]
	pos.SideToMove = Black
	pos.genPawnAttackMoves(All, &moves)
	testMoves(t, moves, []string{"d7c6", "f7g6", "a5b4", "b5a4", "h5g4", "g5h4"})
}

func TestGenPawnAttackMoves2(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)

	var moves []Move
	pos.SideToMove = White
	pos.genPawnAttackMoves(All, &moves)
	testMoves(t, moves, []string{"d5e6", "g2h3"})

	moves = moves[:0]
	pos.SideToMove = Black
	pos.genPawnAttackMoves(All, &moves)
	testMoves(t, moves, []string{"b4c3", "h3g2", "e6d5"})
}

func TestGenPawnEnpassant(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	pos.Put(SquareH1, WhiteKing)
	pos.Put(SquareH8, BlackKing)

	pos.Put(SquareA3, WhitePawn)
	pos.Put(SquareA4, BlackPawn)
	pos.Put(SquareB2, WhitePawn)
	pos.Put(SquareC3, WhitePawn)
	pos.Put(SquareC4, BlackPawn)

	move := pos.UCIToMove("b2b4")

	pos.DoMove(move)
	if SquareB3 != pos.EnpassantSquare() {
		t.Fatalf("expected enpassant square %v, got %v",
			SquareB3, pos.EnpassantSquare())
	}

	var moves []Move
	pos.GenerateFigureMoves(Pawn, All, &moves)
	if 2 != len(moves) {
		t.Fatalf("expected 2 moves, got %d", len(moves))
	}

	for _, m := range moves {
		if Enpassant != m.MoveType() {
			t.Fatalf("expected move typ %v, got %v", Enpassant, m.MoveType())
		}
		if SquareB3 != m.To() {
			t.Fatalf("expected to at %v, got at %v", SquareB3, m.To())
		}
		if SquareB4 != m.CaptureSquare() {
			t.Fatalf("expected capture at %v, got at %v", SquareB4, m.CaptureSquare())
		}
	}

	pos.UndoMove(move)
	if SquareA1 != pos.EnpassantSquare() {
		t.Fatalf("expected enpassant square %v, got %v",
			SquareA1, pos.EnpassantSquare())
	}
}

func TestPawnAttacks(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard2)
	te := &testEngine{T: t, Pos: pos}

	te.Attacked(SquareA4, White, true)
	te.Attacked(SquareB4, White, true)
	te.Attacked(SquareC4, White, true)
	te.Attacked(SquareD4, White, true)
	te.Attacked(SquareE4, White, false)
	te.Attacked(SquareF4, White, true)
	te.Attacked(SquareG4, White, false)
	te.Attacked(SquareB6, White, true)
	te.Attacked(SquareC6, White, false)
	te.Attacked(SquareD6, White, true)

	te.Attacked(SquareA1, Black, false)
	te.Attacked(SquareB1, Black, false)
	te.Attacked(SquareC1, Black, false)
	te.Attacked(SquareD1, Black, false)
	te.Attacked(SquareE1, Black, true)
	te.Attacked(SquareF1, Black, true)
	te.Attacked(SquareG1, Black, true)
	te.Attacked(SquareH1, Black, true)
	te.Attacked(SquareE4, Black, true)
	te.Attacked(SquareG4, Black, true)
}

func TestPawnPromotions(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard2)
	pos.SideToMove = Black
	te := &testEngine{T: t, Pos: pos}

	te.Pawn(SquareF2, []string{"f2f1N", "f2f1B", "f2f1R", "f2f1Q"})
	te.Pawn(SquareG2, []string{
		"g2g1N", "g2g1B", "g2g1R", "g2g1Q",
		"g2h1N", "g2h1B", "g2h1R", "g2h1Q"})

	te.Move("g2h1N")
	te.Piece(SquareG1, NoPiece)
	te.Piece(SquareH1, BlackKnight)

	te.Undo()
	te.Piece(SquareG2, BlackPawn)
	te.Piece(SquareH1, WhiteRook)

	te.Move("f2f1Q")
	te.Piece(SquareF2, NoPiece)
	te.Piece(SquareF1, BlackQueen)

	te.Undo()
	te.Piece(SquareG2, BlackPawn)
	te.Piece(SquareH1, WhiteRook)
}

func TestPawnPromotions2(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	te := &testEngine{T: t, Pos: pos}

	te.Piece(SquareF5, NoPiece)
	te.Move("f3f5")
	te.Move("h3g2")
	te.Move("a1b1")
	te.Move("g2h1N")
	te.Piece(SquareH1, BlackKnight)
	te.Move("e2f1")
	te.Piece(SquareH1, BlackKnight)
	te.Move("h1f2")
}

func TestPawnAttacksEnpassant(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	te := &testEngine{T: t, Pos: pos}

	te.Move("a1d1")
	te.Move("d7d5")
	te.Pawn(SquareE5, []string{"e5e6", "e5d6"})

	te.Move("e2e3")
	te.Move("f7f5")
	te.Pawn(SquareE5, []string{"e5e6", "e5f6"})

	te.Move("e3e4")
	te.Move("f3f4")
	te.Pawn(SquareE5, []string{"e5e6"})

	te.Undo()
	te.Undo()
	te.Pawn(SquareE5, []string{"e5e6", "e5f6"})

	te.Undo()
	te.Undo()
	te.Pawn(SquareE5, []string{"e5e6", "e5d6"})
}

func TestPawnTakesEnpassant(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	te := &testEngine{T: t, Pos: pos}

	te.Move("a1d1")
	te.Move("d7d5")
	te.Pawn(SquareE5, []string{"e5e6", "e5d6"})
	te.Piece(SquareD6, NoPiece)
	te.Piece(SquareD5, BlackPawn)

	te.Move("e5d6")
	te.Piece(SquareD6, WhitePawn)
	te.Piece(SquareD5, NoPiece)

	te.Undo()
	te.Pawn(SquareE5, []string{"e5e6", "e5d6"})
	te.Piece(SquareD6, NoPiece)
	te.Piece(SquareD5, BlackPawn)

	// Makes sure that black pawn at A2/B2 doesn't take Enpassant.
	pos.SideToMove = Black
	pos.SetEnpassantSquare(SquareA1)
	pos.Remove(SquareB2, pos.Get(SquareB2))
	pos.Remove(SquareA1, pos.Get(SquareA1))
	pos.Put(SquareB2, BlackPawn)
	te.Pawn(SquareB2, []string{"b2b1N", "b2b1B", "b2b1R", "b2b1Q"})
	pos.Put(SquareA2, BlackPawn)
	te.Pawn(SquareA2, []string{"a2a1N", "a2a1B", "a2a1R", "a2a1Q"})
}

func TestSquareIsAttackedByKnight(t *testing.T) {
	testBoard2 := "4K3/8/3n4/8/4N3/3n4/8/4k3 w - - 0 1"
	pos, _ := PositionFromFEN(testBoard2)
	te := &testEngine{T: t, Pos: pos}

	te.Attacked(SquareE8, Black, true)
	te.Attacked(SquareC4, Black, true)
	te.Attacked(SquareE1, Black, true)
	te.Attacked(SquareH8, Black, false)
}

func TestIsAttackedByBishop(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	te := &testEngine{T: t, Pos: pos}

	te.Move("e2e4")
	te.Move("d7d5")
	te.Move("f1b5")
	te.Attacked(SquareE8, White, true)

	te.Move("e8d7")
	te.Attacked(SquareD7, White, true)
	te.Undo()

	te.Move("c7c6")
	te.Attacked(SquareE8, White, false)
	te.Attacked(SquareC6, White, true)
}

func TestIsAttackedByKing(t *testing.T) {
	pos := NewPosition()
	pos.SetSideToMove(White)
	te := &testEngine{T: t, Pos: pos}

	pos.Put(SquareE1, WhiteKing)
	pos.Put(SquareD2, BlackPawn)
	pos.Put(SquareE2, BlackPawn)
	pos.Put(SquareF2, BlackPawn)

	te.Attacked(SquareD1, White, true)
	te.Attacked(SquareD2, White, true)
	te.Attacked(SquareE2, White, true)
	te.Attacked(SquareF2, White, true)
	te.Attacked(SquareF1, White, true)
}

func TestPanicPosition(t *testing.T) {
	var moves []Move
	fen := "8/7P/4R3/p4pk1/P2p1r2/3P4/1R6/b1bK4 b - - 1 111"
	pos, _ := PositionFromFEN(fen)
	pos.GenerateMoves(All, &moves)
	for _, m := range moves {
		pos.DoMove(m)
		pos.UndoMove(m)
	}
}

func TestNumPiece(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)

	data := []struct {
		piece Piece
		num   int8
	}{
		{NoPiece, 32},
		{WhitePawn, 8},
		{WhiteKing, 1},
		{BlackKing, 1},
	}

	for _, d := range data {
		actual := pos.NumPieces[d.piece.Color()][d.piece.Figure()]
		if d.num != actual {
			t.Errorf("for %v expected %d, got %d", d.piece, d.num, actual)
		}
	}
}

func TestIsThreeFoldRepetition(t *testing.T) {
	pos, _ := PositionFromFEN(testBoard1)
	te := &testEngine{T: t, Pos: pos}

	te.Move("b1c3")
	te.Move("b8c6")
	te.Move("c3b1")
	te.Move("c6b8")
	if pos.IsThreeFoldRepetition() { // Knights.
		t.Errorf("three fold repetition not expected")
	}

	te.Move("b1c3")
	te.Move("b8c6")
	te.Move("c3b1")
	te.Move("c6b8")
	if !pos.IsThreeFoldRepetition() { // Knights, knights.
		t.Errorf("three fold repetition expected")
	}

	te.Undo()
	te.Undo()
	te.Undo()
	te.Undo()
	te.Move("d2d4")
	te.Move("d7d5")
	if pos.IsThreeFoldRepetition() { // Knights, pawns.
		t.Errorf("three fold repetition not expected")
	}

	te.Move("b1c3")
	te.Move("b8c6")
	te.Move("c3b1")
	te.Move("c6b8")
	if pos.IsThreeFoldRepetition() { // Knights, pawns, knights.
		t.Errorf("three fold repetition not expected")
	}

	te.Move("b1c3")
	te.Move("b8c6")
	te.Move("c3b1")
	te.Move("c6b8")
	if !pos.IsThreeFoldRepetition() { // Knights, pawns, knights, knights.
		t.Errorf("three fold repetition expected")
	}
}

func TestNullMoveSimple(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	te := &testEngine{T: t, Pos: pos}

	te.Move("")
	if Black != pos.SideToMove {
		t.Fatalf("bad nullmove SideToMove. expected %v, got %v", Black, pos.SideToMove)
	}
	te.Undo()
	if FENStartPos != pos.String() {
		t.Fatalf("bad nullmove undo. expected %s, got %s", FENStartPos, pos.String())
	}
}

func TestNullMoveEnpassantSquare(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	te := &testEngine{T: t, Pos: pos}

	te.Move("d2d4")
	te.Move("")
	if White != pos.SideToMove {
		t.Fatalf("bad nullmove SideToMove. expected %v, got %v", White, pos.SideToMove)
	}
	if SquareA1 != pos.EnpassantSquare() {
		t.Fatalf("bad nullmove EnpassantSquare. expected none, got %v", pos.EnpassantSquare())
	}

	te.Undo()
	if SquareD3 != pos.EnpassantSquare() {
		t.Fatalf("bad nullmove EnpassantSquare. expected %v, got %v", SquareD3, pos.EnpassantSquare())
	}
	te.Undo()
}

func TestNullMoveCastlingAbility(t *testing.T) {
	pos, _ := PositionFromFEN(FENKiwipete)
	te := &testEngine{T: t, Pos: pos}

	te.Move("")
	if Black != pos.SideToMove {
		t.Fatalf("bad nullmove SideToMove. expected %v, got %v", Black, pos.SideToMove)
	}
	if AnyCastle != pos.CastlingAbility() {
		t.Fatalf("bad nullmove CastlingAbility. expected %v, got %v", AnyCastle, pos.CastlingAbility())
	}

	te.Undo()
	if FENKiwipete != pos.String() {
		t.Fatalf("bad nullmove undo. expected %s, got %s", FENStartPos, pos.String())
	}
	if AnyCastle != pos.CastlingAbility() {
		t.Fatalf("bad nullmove CastlingAbility. expected %v, got %v", AnyCastle, pos.CastlingAbility())
	}
}

func TestGenerateMovesKind(t *testing.T) {
	for _, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)

		v := make(map[Move]int)
		for _, k := range []int{Violent, Tactical, Quiet} {
			var moves []Move
			pos.GenerateMoves(k, &moves)
			for _, m := range moves {
				v[m] |= k
			}
		}

		for m, k := range v {
			if k != Violent && k != Tactical && k != Quiet {
				t.Errorf("invalid kind for move %v, fen %s", m, fen)
			}
		}

		var all []Move
		pos.GenerateMoves(All, &all)
		if len(all) != len(v) {
			t.Errorf("Expected %d moves, got %d", len(all), len(v))
		}
	}
}

func TestGenerateMovesQuiet(t *testing.T) {
	for _, fen := range testFENs {
		var all []Move
		pos, _ := PositionFromFEN(fen)
		pos.GenerateMoves(Quiet, &all)
		for _, m := range all {
			if m.MoveType() != Normal || m.Capture() != NoPiece {
				t.Errorf("Expected quiet move, got %v", m)
			}
		}
	}
}

func TestGenerateMovesTactical(t *testing.T) {
	for _, fen := range testFENs {
		var all []Move
		pos, _ := PositionFromFEN(fen)
		pos.GenerateMoves(Tactical, &all)
		for _, m := range all {
			if m.MoveType() != Castling && (m.MoveType() != Promotion || m.Target().Figure() == Queen) {
				t.Errorf("Expected tactical move, got %v", m)
			}
		}
	}
}

func TestGenerateMovesColor(t *testing.T) {
	for _, fen := range testFENs {
		var all []Move
		pos, _ := PositionFromFEN(fen)
		pos.GenerateMoves(All, &all)
		for _, m := range all {
			if m.Piece().Color() != pos.SideToMove {
				t.Errorf("for move %v and fen %v", m, fen)
				t.Errorf("expected piece color %v, got %v", pos.SideToMove, m.Piece().Color())
			}
			if m.Target().Color() != pos.SideToMove {
				t.Errorf("for move %v and fen %v", m, fen)
				t.Errorf("expected target color %v, got %v", pos.SideToMove, m.Piece().Color())
			}
			if m.MoveType() == Promotion && m.Promotion().Color() != pos.SideToMove {
				t.Errorf("for move %v and fen %v", m, fen)
				t.Errorf("expected target color %v, got %v", pos.SideToMove, m.Piece().Color())
			}
		}
	}
}
