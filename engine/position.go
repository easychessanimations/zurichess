package engine

import (
	"fmt"
	"log"
)

var (
	// Which castle rights are lost when pieces are moved.
	lostCastleRights [64]Castle
	// Pieces into which a pawn can be promoted.
	pawnPromotions = []Figure{Knight, Bishop, Rook, Queen}
	// Maps pieces to symbols. ? means invalid.
	pieceToSymbol = ".????pP??nN??bB??rR??qQ??kK?"
)

func init() {
	lostCastleRights[SquareA1] = WhiteOOO
	lostCastleRights[SquareE1] = WhiteOOO | WhiteOO
	lostCastleRights[SquareH1] = WhiteOO
	lostCastleRights[SquareA8] = BlackOOO
	lostCastleRights[SquareE8] = BlackOOO | BlackOO
	lostCastleRights[SquareH8] = BlackOO
}

// Position encodes the chess board.
type Position struct {
	ByFigure   [FigureArraySize]Bitboard
	ByColor    [ColorArraySize]Bitboard
	SideToMove Color
	Castle     Castle
	Enpassant  Square
	Zobrist    uint64
}

// Verify check the validity of the position.
// Mostly used for debugging purposes.
func (pos *Position) Verify() error {
	if bb := pos.ByColor[White] & pos.ByColor[Black]; bb != 0 {
		sq := bb.Pop()
		return fmt.Errorf("Square %v is both White and Black", sq)
	}
	// Check that there is at most one king.
	// Catches castling issues.
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		bb := pos.ByPiece(col, King)
		if bb != bb.LSB() {
			return fmt.Errorf("More than one King for %v", col)
		}
	}
	return nil
}

// SetCastlingAbility sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetCastlingAbility(castle Castle) {
	pos.Zobrist ^= ZobristCastle[pos.Castle]
	pos.Castle = castle
	pos.Zobrist ^= ZobristCastle[pos.Castle]
}

// SetSideToMove sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetSideToMove(col Color) {
	pos.Zobrist ^= ZobristColor[pos.SideToMove]
	pos.SideToMove = col
	pos.Zobrist ^= ZobristColor[pos.SideToMove]
}

// SetEnpassantSquare sets the enpassant square correctly updating the Zobrist key.
func (pos *Position) SetEnpassantSquare(sq Square) {
	pos.Zobrist ^= ZobristEnpassant[pos.Enpassant]
	pos.Enpassant = sq
	pos.Zobrist ^= ZobristEnpassant[pos.Enpassant]
}

// ByPiece is a shortcut for ByColor[col]&ByFigure[fig].
func (pos *Position) ByPiece(col Color, fig Figure) Bitboard {
	return pos.ByColor[col] & pos.ByFigure[fig]
}

// NumPieces returns the number of pieces on the board.
func (pos *Position) NumPieces() int {
	return (pos.ByColor[White] | pos.ByColor[Black]).Popcnt()
}

// Put puts a piece on the board.
// Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	pos.Zobrist ^= ZobristPiece[pi][sq]
	bb := sq.Bitboard()
	pos.ByColor[pi.Color()] |= bb
	pos.ByFigure[pi.Figure()] |= bb
}

// Remove removes a piece from the table.
// Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	pos.Zobrist ^= ZobristPiece[pi][sq]
	bb := ^sq.Bitboard()
	pos.ByColor[pi.Color()] &= bb
	pos.ByFigure[pi.Figure()] &= bb
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.ByColor[White]|pos.ByColor[Black])>>sq&1 == 0
}

// GetColor returns the piece's color at sq.
func (pos *Position) GetColor(sq Square) Color {
	return White*Color(pos.ByColor[White]>>sq&1) +
		Black*Color(pos.ByColor[Black]>>sq&1)
}

// GetFigure returns the piece's type at sq.
func (pos *Position) GetFigure(sq Square) Figure {
	for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
		if pos.ByFigure[fig]&sq.Bitboard() != 0 {
			return fig
		}
	}
	return NoFigure
}

// Get returns the piece at sq.
func (pos *Position) Get(sq Square) Piece {
	if col := pos.GetColor(sq); col != NoColor {
		return ColorFigure(col, pos.GetFigure(sq))
	}
	return NoPiece
}

// IsChecked returns true if co's king is checked.
func (pos *Position) IsChecked(col Color) bool {
	kingSq := pos.ByPiece(col, King).AsSquare()
	return pos.IsAttackedBy(kingSq, col.Opposite())
}

// PrettyPrint pretty prints the current position.
func (pos *Position) PrettyPrint() {
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			if sq == pos.Enpassant {
				line += ","
			} else {
				line += string(pieceToSymbol[pos.Get(sq)])
			}
		}
		if r == 7 && pos.SideToMove == Black {
			line += " *"
		}
		if r == 0 && pos.SideToMove == White {
			line += " *"
		}
		log.Println(line)
	}

}

// fix updates move so that it can be undone.
// Does not set capture.
func (pos *Position) fix(move Move) Move {
	move.SavedCastle = pos.Castle
	move.SavedEnpassant = pos.Enpassant
	return move
}

// DoMove performs a move of known piece.
// Expects the move to be valid.
func (pos *Position) DoMove(move Move) {
	if move.SideToMove() != pos.SideToMove {
		panic(fmt.Errorf("bad move %v: expected %v piece at %v, got %v",
			move, pos.SideToMove, move.From, move.Piece()))
	}

	// Update castling rights based on the source&target squares.
	pos.SetCastlingAbility(pos.Castle & ^lostCastleRights[move.From] & ^lostCastleRights[move.To])

	// Move rook on castling.
	if move.MoveType == Castling {
		rook, start, end := CastlingRook(move.To)
		pos.Remove(start, rook)
		pos.Put(end, rook)
	}

	// Set Enpassant square for capturing.
	pi := move.Piece()
	if pi.Figure() == Pawn &&
		move.From.Bitboard()&BbPawnStartRank != 0 &&
		move.To.Bitboard()&BbPawnDoubleRank != 0 {
		pos.SetEnpassantSquare((move.From + move.To) / 2)
	} else {
		pos.SetEnpassantSquare(SquareA1)
	}

	// Update the pieces the chess board.
	pos.Remove(move.From, pi)
	pos.Remove(move.CaptureSquare(), move.Capture)
	pos.Put(move.To, move.Target)
	pos.SetSideToMove(pos.SideToMove.Opposite())
}

// UndoMove takes back a move.
// Expects the move to be valid.
func (pos *Position) UndoMove(move Move) {
	pos.SetSideToMove(pos.SideToMove.Opposite())

	// Modify the chess board.
	pi := move.Piece()
	pos.Put(move.From, pi)
	pos.Remove(move.To, move.Target)
	pos.Put(move.CaptureSquare(), move.Capture)

	// Move rook on castling.
	if move.MoveType == Castling {
		rook, start, end := CastlingRook(move.To)
		pos.Put(start, rook)
		pos.Remove(end, rook)
	}

	pos.SetCastlingAbility(move.SavedCastle)
	pos.SetEnpassantSquare(move.SavedEnpassant)
}

func (pos *Position) genPawnPromotions(from, to Square, capt Piece, violent bool, moves []Move) []Move {
	rank := to.Rank()
	if violent && rank != 0 && rank != 7 && capt == NoPiece {
		return moves
	}

	moveType := Normal
	pMin, pMax := Pawn, Pawn
	if rank == 0 || rank == 7 {
		moveType = Promotion
		if violent {
			pMin, pMax = Queen, Queen
		} else {
			pMin, pMax = Knight, Queen
		}
	}
	if moveType == Normal && pos.Enpassant != SquareA1 && pos.Enpassant == to {
		moveType = Enpassant
	}

	for p := pMin; p <= pMax; p++ {
		moves = append(moves, pos.fix(Move{
			MoveType: moveType,
			From:     from,
			To:       to,
			Capture:  capt,
			Target:   ColorFigure(pos.SideToMove, p),
		}))
	}
	return moves
}

// genPawnAdvanceMoves moves pawns one square.
func (pos *Position) genPawnAdvanceMoves(violent bool, moves []Move) []Move {
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
	var forward Square

	if pos.SideToMove == White {
		bb &= free >> 8
		forward = RankFile(+1, 0)
	} else {
		bb &= free << 8
		forward = RankFile(-1, 0)
	}

	for bb != 0 {
		from := bb.Pop()
		to := from + forward
		moves = pos.genPawnPromotions(from, to, NoPiece, violent, moves)
	}
	return moves
}

// genPawnDoubleAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(moves []Move) []Move {
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
	var forward Square

	if pos.SideToMove == White {
		bb &= RankBb(1) & (free >> 8) & (free >> 16)
		forward = RankFile(+2, 0)
	} else {
		bb &= RankBb(6) & (free << 8) & (free << 16)
		forward = RankFile(-2, 0)
	}

	for bb != 0 {
		from := bb.Pop()
		to := from + forward
		moves = pos.genPawnPromotions(from, to, NoPiece, false, moves)
	}
	return moves
}

func (pos *Position) pawnCapture(to Square) Piece {
	if pos.Enpassant != SquareA1 && to == pos.Enpassant {
		return ColorFigure(pos.SideToMove.Opposite(), Pawn)
	}
	return pos.Get(to)
}

func (pos *Position) genPawnAttackMoves(violent bool, moves []Move) []Move {
	enemy := pos.ByColor[pos.SideToMove.Opposite()]
	if pos.Enpassant != SquareA1 {
		enemy |= pos.Enpassant.Bitboard()
	}

	var forward Square
	if pos.SideToMove == White {
		enemy >>= 8
		forward = RankFile(+1, 0)
	} else {
		enemy <<= 8
		forward = RankFile(-1, 0)
	}

	// Left
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	att := forward.Relative(0, -1)
	for bbl := bb & BbPawnLeftAttack & (enemy << 1); bbl > 0; {
		from := bbl.Pop()
		to := from + att
		capt := pos.pawnCapture(to)
		moves = pos.genPawnPromotions(from, to, capt, violent, moves)
	}

	// Right
	att = forward.Relative(0, +1)
	for bbr := bb & BbPawnRightAttack & (enemy >> 1); bbr > 0; {
		from := bbr.Pop()
		to := from + att
		capt := pos.pawnCapture(to)
		moves = pos.genPawnPromotions(from, to, capt, violent, moves)
	}

	return moves
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, violent bool, moves []Move) []Move {
	if violent {
		// Capture enemy pieces.
		att &= pos.ByColor[pos.SideToMove.Opposite()]
	} else {
		// Generate all moves, except capturing own pieces.
		att &= ^pos.ByColor[pos.SideToMove]
	}

	for bb := att; bb != 0; {
		to := bb.Pop()
		moves = append(moves, pos.fix(Move{
			MoveType: Normal,
			From:     from,
			To:       to,
			Capture:  pos.Get(to),
			Target:   pi,
		}))
	}
	return moves
}

func (pos *Position) genKnightMoves(violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.SideToMove, Knight)
	for bb := pos.ByPiece(pos.SideToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := BbKnightAttack[from]
		moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	}
	return moves
}

func (pos *Position) genBishopMoves(fig Figure, violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := BishopMagic[from].Attack(ref)
		moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	}
	return moves
}

func (pos *Position) genRookMoves(fig Figure, violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := RookMagic[from].Attack(ref)
		moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	}
	return moves
}

// Like other gen*Moves functions it might leave the king in check.
func (pos *Position) genKingMoves(moves []Move) []Move {
	moves = pos.genKingMovesNear(false, moves)
	moves = pos.genKingCastles(moves)
	return moves
}

func (pos *Position) genKingMovesNear(violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.SideToMove, King)
	from := pos.ByPiece(pos.SideToMove, King).AsSquare()
	att := BbKingAttack[from]
	moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	return moves
}

func (pos *Position) genKingCastles(moves []Move) []Move {
	rank := pos.SideToMove.KingHomeRank()
	oo, ooo := WhiteOO, WhiteOOO
	if pos.SideToMove == Black {
		oo, ooo = BlackOO, BlackOOO
	}

	// Castle king side.
	if pos.Castle&oo != 0 {
		r5 := RankFile(rank, 5)
		r6 := RankFile(rank, 6)
		if !pos.IsEmpty(r5) || !pos.IsEmpty(r6) {
			goto EndCastleOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.IsAttackedBy(r4, other) ||
			pos.IsAttackedBy(r5, other) ||
			pos.IsAttackedBy(r6, other) {
			goto EndCastleOO
		}

		moves = append(moves, pos.fix(Move{
			MoveType: Castling,
			From:     r4,
			To:       r6,
			Target:   ColorFigure(pos.SideToMove, King),
		}))
	}
EndCastleOO:

	// Castle queen side.
	if pos.Castle&ooo != 0 {
		r3 := RankFile(rank, 3)
		r2 := RankFile(rank, 2)
		r1 := RankFile(rank, 1)
		if !pos.IsEmpty(r3) || !pos.IsEmpty(r2) || !pos.IsEmpty(r1) {
			goto EndCastleOOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.IsAttackedBy(r4, other) ||
			pos.IsAttackedBy(r3, other) ||
			pos.IsAttackedBy(r2, other) {
			goto EndCastleOOO
		}

		moves = append(moves, pos.fix(Move{
			MoveType: Castling,
			From:     r4,
			To:       r2,
			Target:   ColorFigure(pos.SideToMove, King),
		}))
	}
EndCastleOOO:

	return moves
}

// IsAttackedBy returns true if sq is under attacked by co.
func (pos *Position) IsAttackedBy(sq Square, co Color) bool {
	enemy := pos.ByColor[co]
	if BbPawnAttack[sq]&enemy&pos.ByFigure[Pawn] != 0 {
		bb := sq.Bitboard()
		if co == White { // WhitePawn
			bb >>= 8
		} else { // BlackPawn
			bb <<= 8
		}

		pawns := pos.ByPiece(co, Pawn)
		pawnsLeft := (BbPawnLeftAttack & pawns) >> 1
		pawnsRight := (BbPawnRightAttack & pawns) << 1
		if att := bb & (pawnsLeft | pawnsRight); att != 0 {
			return true
		}
	}

	// Knight
	if BbKnightAttack[sq]&enemy&pos.ByFigure[Knight] != 0 {
		return true
	}

	// Quick test of queen's attack on an empty board.
	if BbSuperAttack[sq]&(enemy&^pos.ByFigure[Pawn]) == 0 {
		return false
	}

	// King.
	if BbKingAttack[sq]&enemy&pos.ByFigure[King] != 0 {
		return true
	}

	// Bishop&Queen
	all := pos.ByColor[White] | pos.ByColor[Black]
	bishops := enemy & (pos.ByFigure[Bishop] | pos.ByFigure[Queen])
	if bishops != 0 && bishops&BishopMagic[sq].Attack(all) != 0 {
		return true
	}

	// Rook&Queen
	rooks := enemy & (pos.ByFigure[Rook] | pos.ByFigure[Queen])
	if rooks != 0 && rooks&RookMagic[sq].Attack(all) != 0 {
		return true
	}

	return false
}

// GenerateMoves appends to moves all moves valid from pos.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
func (pos *Position) GenerateMoves(moves []Move) []Move {
	moves = pos.genPawnAttackMoves(false, moves)
	moves = pos.genKnightMoves(false, moves)
	moves = pos.genBishopMoves(Bishop, false, moves)
	moves = pos.genRookMoves(Rook, false, moves)
	moves = pos.genBishopMoves(Queen, false, moves)
	moves = pos.genRookMoves(Queen, false, moves)

	moves = pos.genPawnAdvanceMoves(false, moves)
	moves = pos.genKingMovesNear(false, moves)
	moves = pos.genKingCastles(moves)
	moves = pos.genPawnDoubleAdvanceMoves(moves)
	return moves
}

// GenerateViolentMoves append to moves all violent moves valid from pos.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
func (pos *Position) GenerateViolentMoves(moves []Move) []Move {
	moves = pos.genPawnAdvanceMoves(true, moves)
	moves = pos.genPawnAttackMoves(true, moves)
	moves = pos.genKnightMoves(true, moves)
	moves = pos.genBishopMoves(Bishop, true, moves)
	moves = pos.genBishopMoves(Queen, true, moves)
	moves = pos.genRookMoves(Rook, true, moves)
	moves = pos.genRookMoves(Queen, true, moves)
	moves = pos.genKingMovesNear(true, moves)
	return moves
}

// GenerateFigureMoves generate moves for a given figure.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
func (pos *Position) GenerateFigureMoves(fig Figure, moves []Move) []Move {
	switch fig {
	case Pawn:
		moves = pos.genPawnAttackMoves(false, moves)
		moves = pos.genPawnAdvanceMoves(false, moves)
		moves = pos.genPawnDoubleAdvanceMoves(moves)
	case Knight:
		moves = pos.genKnightMoves(false, moves)
	case Bishop:
		moves = pos.genBishopMoves(Bishop, false, moves)
	case Rook:
		moves = pos.genRookMoves(Rook, false, moves)
	case Queen:
		moves = pos.genBishopMoves(Queen, false, moves)
		moves = pos.genRookMoves(Queen, false, moves)
	case King:
		moves = pos.genKingMovesNear(false, moves)
		moves = pos.genKingCastles(moves)
	}
	return moves
}
