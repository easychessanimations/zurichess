package engine

import (
	"fmt"
	"log"
)

var (
	_ = log.Println

	// Which castle rights are lost when pieces are moved.
	lostCastleRights [64]Castle

	// Pieces into which a pawn can be promoted.
	pawnPromotions = []Figure{Knight, Bishop, Rook, Queen}

	// Maps pieces to symbols. ? means invalid.
	pieceToSymbol = ".????pP??nN??bB??rR??qQ??kK?"

	// Maps runes to pieces. It's the reverse of the previous array.
	symbolToPiece = map[rune]Piece{
		'p': BlackPawn,
		'n': BlackKnight,
		'b': BlackBishop,
		'r': BlackRook,
		'q': BlackQueen,
		'k': BlackKing,

		'P': WhitePawn,
		'N': WhiteKnight,
		'B': WhiteBishop,
		'R': WhiteRook,
		'Q': WhiteQueen,
		'K': WhiteKing,
	}

	// Maps runes to figures.
	symbolToFigure = map[rune]Figure{
		'p': Pawn,
		'n': Knight,
		'b': Bishop,
		'r': Rook,
		'q': Queen,
		'k': King,

		'P': Pawn,
		'N': Knight,
		'B': Bishop,
		'R': Rook,
		'Q': Queen,
		'K': King,
	}
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
	ByFigure  [FigureArraySize]Bitboard
	ByColor   [ColorArraySize]Bitboard
	ToMove    Color
	Castle    Castle
	Enpassant Square
	Zobrist   uint64
}

// SetCastle sets the side to move, correctly updating the zobriest key.
func (pos *Position) SetCastlingAbility(castle Castle) {
	pos.Zobrist ^= ZobriestCastle[pos.Castle]
	pos.Castle = castle
	pos.Zobrist ^= ZobriestCastle[pos.Castle]
}

// SetEnpassant sets the side to move, correctly updating the zobriest key.
func (pos *Position) SetSideToMove(col Color) {
	pos.Zobrist ^= ZobriestColor[pos.ToMove]
	pos.ToMove = col
	pos.Zobrist ^= ZobriestColor[pos.ToMove]
}

// SetEnpassant sets the enpassant square correctly updating the zobriest key.
func (pos *Position) SetEnpassantSquare(sq Square) {
	pos.Zobrist ^= ZobriestEnpassant[pos.Enpassant]
	pos.Enpassant = sq
	pos.Zobrist ^= ZobriestEnpassant[pos.Enpassant]
}

// ByPiece is a shortcut for byColor&byFigure.
func (pos *Position) ByPiece(col Color, fig Figure) Bitboard {
	return pos.ByColor[col] & pos.ByFigure[fig]
}

// Put puts a piece on the board.
// Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	pos.Zobrist ^= ZobriestPiece[pi][sq]
	bb := sq.Bitboard()
	pos.ByColor[pi.Color()] |= bb
	pos.ByFigure[pi.Figure()] |= bb
}

// Remove removes a piece from the table.
// Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	pos.Zobrist ^= ZobriestPiece[pi][sq]
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
	col := pos.GetColor(sq)
	if col == NoColor {
		return NoPiece
	}
	fig := pos.GetFigure(sq)
	return ColorFigure(col, fig)
}

// IsChecked returns true if co's king is checked.
func (pos *Position) IsChecked(co Color) bool {
	kingSq := (pos.ByColor[co] & pos.ByFigure[King]).AsSquare()
	return pos.IsAttackedBy(kingSq, co.Other())
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
		if r == 7 && pos.ToMove == Black {
			line += " *"
		}
		if r == 0 && pos.ToMove == White {
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

// DoMovePiece performs a move of known piece.
// Expects the move to be valid.
func (pos *Position) DoMove(move Move) {
	pi := move.Target
	if move.MoveType == Promotion {
		pi = ColorFigure(pos.ToMove, Pawn)
	}

	if pi.Color() != pos.ToMove {
		panic(fmt.Errorf("bad move %v: expected %v piece at %v, got %v",
			move, pos.ToMove, move.From, pi))
	}

	/*
		log.Println(
			pos.Get(move.From), "playing", move,
			"; castling ", pos.Castle,
			"; Enpassant", pos.Enpassant)
	*/

	// Update castling rights based on the source&target squares.
	pos.SetCastlingAbility(pos.Castle & ^lostCastleRights[move.From] & ^lostCastleRights[move.To])

	// Move rook on castling.
	if move.MoveType == Castling {
		rook, start, end := CastlingRook(move.To)
		pos.Remove(start, rook)
		pos.Put(end, rook)
	}

	// Set Enpassant square for capturing.
	if pi.Figure() == Pawn &&
		move.From.Bitboard()&BbPawnStartRank != 0 &&
		move.To.Bitboard()&BbPawnDoubleRank != 0 {
		pos.SetEnpassantSquare((move.From + move.To) / 2)
	} else {
		pos.SetEnpassantSquare(SquareA1)
	}

	// Capture pawn Enpassant.
	captSq := move.To
	if move.MoveType == Enpassant {
		captSq = RankFile(move.From.Rank(), move.To.File())
	}

	if move.Capture != NoPiece && pos.IsEmpty(captSq) {
		panic(fmt.Errorf("invalid capture: expected %v at %v, got %v. move is %+q",
			move.Capture, captSq, pos.Get(captSq), move))
	}

	// Update the pieces the chess board.
	pos.Remove(move.From, pi)
	pos.Remove(captSq, move.Capture)
	pos.Put(move.To, move.Target)
	pos.SetSideToMove(pos.ToMove.Other())
}

// UndoMovePiece takes back a move.
// Expects the move to be valid.
// pi must be the piece moved, i.e. the pawn in case of promotions.
func (pos *Position) UndoMove(move Move) {
	pos.SetSideToMove(pos.ToMove.Other())
	pi := move.Target
	if move.MoveType == Promotion {
		pi = ColorFigure(pos.ToMove, Pawn)
	}

	captSq := move.To
	if move.MoveType == Enpassant {
		captSq = RankFile(move.From.Rank(), move.To.File())
	}

	// Modify the chess board.
	pos.Put(move.From, pi)
	if move.MoveType == Promotion {
		pos.Remove(move.To, move.Target)
	} else {
		pos.Remove(move.To, pi)
	}
	pos.Put(captSq, move.Capture)

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
	pr := to.Rank()
	if pr != 0 && pr != 7 {
		if !violent || capt != NoPiece {
			moveType := Normal
			if pos.Enpassant != SquareA1 && to == pos.Enpassant {
				moveType = Enpassant
			}
			moves = append(moves, pos.fix(Move{
				MoveType: moveType,
				From:     from,
				To:       to,
				Capture:  capt,
				Target:   ColorFigure(pos.ToMove, Pawn),
			}))
		}
	} else {
		for _, promo := range pawnPromotions {
			moves = append(moves, pos.fix(Move{
				MoveType: Promotion,
				From:     from,
				To:       to,
				Capture:  capt,
				Target:   ColorFigure(pos.ToMove, promo),
			}))
		}
	}
	return moves
}

// GenPawnAdvanceMoves moves pawns one square.
func (pos *Position) genPawnAdvanceMoves(violent bool, moves []Move) []Move {
	bb := pos.ByColor[pos.ToMove] & pos.ByFigure[Pawn]
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
	var forward Square

	if pos.ToMove == White {
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

// GenPawnAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(moves []Move) []Move {
	bb := pos.ByColor[pos.ToMove] & pos.ByFigure[Pawn]
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
	var forward Square

	if pos.ToMove == White {
		bb &= BbRank2 & (free >> 8) & (free >> 16)
		forward = RankFile(+2, 0)
	} else {
		bb &= BbRank7 & (free << 8) & (free << 16)
		forward = RankFile(-2, 0)
	}

	for bb != 0 {
		from := bb.Pop()
		to := from + forward
		moves = pos.genPawnPromotions(from, to, NoPiece, false, moves)
	}
	return moves
}

func (pos *Position) genPawnAttackMoves(moves []Move) []Move {
	bb := pos.ByColor[pos.ToMove] & pos.ByFigure[Pawn]
	enemy := pos.ByColor[pos.ToMove.Other()]
	var forward Square

	if pos.ToMove == White {
		enemy >>= 8
		forward = RankFile(+1, 0)
	} else {
		enemy <<= 8
		forward = RankFile(-1, 0)
	}

	// Left
	bbl := bb & BbPawnLeftAttack & (enemy << 1)
	att := forward.Relative(0, -1)
	for bbl > 0 {
		from := bbl.Pop()
		to := from + att
		capt := pos.Get(to)
		moves = pos.genPawnPromotions(from, to, capt, true, moves)
	}

	// Right
	bbr := bb & BbPawnRightAttack & (enemy >> 1)
	att = forward.Relative(0, +1)
	for bbr > 0 {
		from := bbr.Pop()
		to := from + att
		capt := pos.Get(to)
		moves = pos.genPawnPromotions(from, to, capt, true, moves)
	}

	return moves
}

func (pos *Position) genPawnEnpassantMoves(moves []Move) []Move {
	if pos.Enpassant == SquareA1 {
		return moves
	}

	bb := pos.ByColor[pos.ToMove] & pos.ByFigure[Pawn]
	enemy := pos.Enpassant.Bitboard()

	if pos.ToMove == White {
		enemy >>= 8
	} else {
		enemy <<= 8
	}

	// Left
	bbl := bb & BbPawnLeftAttack & (enemy << 1)
	if bbl != 0 {
		from := bbl.AsSquare()
		to := pos.Enpassant
		capt := ColorFigure(pos.ToMove.Other(), Pawn)
		moves = pos.genPawnPromotions(from, to, capt, true, moves)
	}

	// Right
	bbr := bb & BbPawnRightAttack & (enemy >> 1)
	if bbr != 0 {
		from := bbr.AsSquare()
		to := pos.Enpassant
		capt := ColorFigure(pos.ToMove.Other(), Pawn)
		moves = pos.genPawnPromotions(from, to, capt, true, moves)
	}

	return moves

}

// GenPawnMoves generates pawn moves around from.
func (pos *Position) genPawnMoves(moves []Move) []Move {
	moves = pos.genPawnAdvanceMoves(false, moves)
	moves = pos.genPawnDoubleAdvanceMoves(moves)
	moves = pos.genPawnEnpassantMoves(moves)
	moves = pos.genPawnAttackMoves(moves)
	return moves
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, violent bool, moves []Move) []Move {
	if violent {
		att &= pos.ByColor[pos.ToMove.Other()]
	}
	att &= ^pos.ByColor[pos.ToMove]
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
	pi := ColorFigure(pos.ToMove, Knight)
	for bb := pos.ByPiece(pos.ToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := BbKnightAttack[from]
		moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	}
	return moves
}

func (pos *Position) genBishopMoves(fig Figure, violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.ToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.ToMove, fig); bb != 0; {
		from := bb.Pop()
		att := BishopMagic[from].Attack(ref)
		moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	}
	return moves
}

func (pos *Position) genRookMoves(fig Figure, violent bool, moves []Move) []Move {
	pi := ColorFigure(pos.ToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.ToMove, fig); bb != 0; {
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
	pi := ColorFigure(pos.ToMove, King)
	from := pos.ByPiece(pos.ToMove, King).AsSquare()
	att := BbKingAttack[from]
	moves = pos.genBitboardMoves(pi, from, att, violent, moves)
	return moves
}

func (pos *Position) genKingCastles(moves []Move) []Move {
	// King Castles.
	pi, oo, ooo := WhiteKing, WhiteOO, WhiteOOO
	rank, from := 0, SquareE1
	if pos.ToMove == Black {
		pi, oo, ooo = BlackKing, BlackOO, BlackOOO
		rank, from = 7, SquareE8
	}
	if pos.Castle&(oo|ooo) == 0 {
		return moves
	}

	other := pos.ToMove.Other()

	// Castle king side.
	if pos.Castle&oo != 0 {
		r5 := RankFile(rank, 5)
		r6 := RankFile(rank, 6)
		if pos.IsEmpty(r5) && pos.IsEmpty(r6) {
			if !pos.IsAttackedBy(from, other) &&
				!pos.IsAttackedBy(r5, other) &&
				!pos.IsAttackedBy(r6, other) {
				moves = append(moves, pos.fix(Move{
					MoveType: Castling,
					From:     from,
					To:       RankFile(rank, 6),
					Target:   pi,
				}))
			}
		}
	}

	// Castle queen side.
	if pos.Castle&ooo != 0 {
		r3 := RankFile(rank, 3)
		r2 := RankFile(rank, 2)
		r1 := RankFile(rank, 1)
		if pos.IsEmpty(r3) && pos.IsEmpty(r2) && pos.IsEmpty(r1) {
			if !pos.IsAttackedBy(from, other) &&
				!pos.IsAttackedBy(r3, other) &&
				!pos.IsAttackedBy(r2, other) {
				moves = append(moves, pos.fix(Move{
					MoveType: Castling,
					From:     from,
					To:       RankFile(rank, 2),
					Target:   pi,
				}))
			}
		}
	}

	return moves
}

// IsAttackedBy returns true if sq is under attacked by co.
func (pos *Position) IsAttackedBy(sq Square, co Color) bool {
	enemy := pos.ByColor[co]
	if BbPawnAttack[sq]&enemy&pos.ByFigure[Pawn] != 0 {
		bb := sq.Bitboard()
		pawns := pos.ByPiece(co, Pawn)
		pawnsLeft := (BbPawnLeftAttack & pawns) >> 1
		pawnsRight := (BbPawnRightAttack & pawns) << 1

		if co == White { // WhitePawn
			bb >>= 8
		} else { // BlackPawn
			bb <<= 8
		}

		if att := bb & (pawnsLeft | pawnsRight); att != 0 {
			return true
		}
	}

	// Knight
	if BbKnightAttack[sq]&enemy&pos.ByFigure[Knight] != 0 {
		return true
	}

	// Quick test of SuperPiece (Bishop, Rook, Queen, King) on empty board.
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

// GenerateMoves is a helper to generate all moves.
func (pos *Position) GenerateMoves(moves []Move) []Move {
	moves = pos.genPawnAttackMoves(moves)
	moves = pos.genPawnEnpassantMoves(moves)
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

// GenerateMoves is a helper to generate all moves.
func (pos *Position) GenerateViolentMoves(moves []Move) []Move {
	moves = pos.genPawnAdvanceMoves(true, moves)
	moves = pos.genPawnAttackMoves(moves)
	moves = pos.genPawnEnpassantMoves(moves)
	moves = pos.genKnightMoves(true, moves)
	moves = pos.genBishopMoves(Bishop, true, moves)
	moves = pos.genBishopMoves(Queen, true, moves)
	moves = pos.genRookMoves(Rook, true, moves)
	moves = pos.genRookMoves(Queen, true, moves)
	moves = pos.genKingMovesNear(true, moves)
	return moves
}

// GenerateFigureMoves generate moves for a given figure.
func (pos *Position) GenerateFigureMoves(fig Figure, moves []Move) []Move {
	switch fig {
	case Pawn:
		moves = pos.genPawnEnpassantMoves(moves)
		moves = pos.genPawnAttackMoves(moves)
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
