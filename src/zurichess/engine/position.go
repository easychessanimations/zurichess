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
)

func initCastleRights() {
	lostCastleRights[SquareA1] = WhiteOOO
	lostCastleRights[SquareE1] = WhiteOOO | WhiteOO
	lostCastleRights[SquareH1] = WhiteOO
	lostCastleRights[SquareA8] = BlackOOO
	lostCastleRights[SquareE8] = BlackOOO | BlackOO
	lostCastleRights[SquareH8] = BlackOO
}

func init() {
	initCastleRights()
}

// Position encodes the chess board.
type Position struct {
	ByFigure  [FigureMaxValue]Bitboard
	ByColor   [ColorMaxValue]Bitboard
	ToMove    Color
	Castle    Castle
	Enpassant Square
}

// ByPiece is a shortcut for byColor&byFigure.
func (pos *Position) ByPiece(col Color, fig Figure) Bitboard {
	return pos.ByColor[col] & pos.ByFigure[fig]
}

// Put puts a piece on the board.
// Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	bb := sq.Bitboard()
	pos.ByColor[pi.Color()] |= bb
	pos.ByFigure[pi.Figure()] |= bb
}

// Remove removes a piece from the table.
// Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
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
	for pt := FigureMinValue; pt < FigureMaxValue; pt++ {
		if pos.ByFigure[pt]&sq.Bitboard() != 0 {
			return pt
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
			line += pos.Get(RankFile(r, f)).Symbol()
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
	move.OldCastle = pos.Castle
	move.OldEnpassant = pos.Enpassant
	return move
}

var symbolToFigure = map[byte]Figure{
	'N': Knight,
	'n': Knight,
	'B': Bishop,
	'b': Bishop,
	'R': Rook,
	'r': Rook,
	'Q': Queen,
	'q': Queen,
}

// ParseMove parses a move given in standard algebraic notation.
// s can be "a2a4" or "h7h8Q" (pawn promotion).
func (pos *Position) ParseMove(s string) Move {
	from := SquareFromString(s[0:2])
	to := SquareFromString(s[2:4])

	mvtp := Normal
	capt := pos.Get(to)
	promo := pos.Get(from)

	pi := pos.Get(from)
	if pi.Figure() == Pawn && pos.Enpassant != SquareA1 && to == pos.Enpassant {
		mvtp = Enpassant
		capt = ColorFigure(pos.ToMove.Other(), Pawn)
	}
	if pi == WhiteKing && from == SquareE1 && (to == SquareC1 || to == SquareG1) {
		mvtp = Castling
	}
	if pi == BlackKing && from == SquareE8 && (to == SquareC8 || to == SquareG8) {
		mvtp = Castling
	}
	if pi.Figure() == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		mvtp = Promotion
		promo = ColorFigure(pos.ToMove, symbolToFigure[s[4]])
	}

	return pos.fix(Move{
		MoveType: mvtp,
		From:     from,
		To:       to,
		Capture:  capt,
		Target:   promo,
	})
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

	// Update castling rights based on the source square.
	pos.Castle &= ^lostCastleRights[move.From]
	// Update castling rights based on the captured square.
	pos.Castle &= ^lostCastleRights[move.To]

	// Move rook on castling.
	if move.MoveType == Castling {
		if move.To == SquareC1 {
			pos.Remove(SquareA1, WhiteRook)
			pos.Put(SquareD1, WhiteRook)
		}
		if move.To == SquareG1 {
			pos.Remove(SquareH1, WhiteRook)
			pos.Put(SquareF1, WhiteRook)
		}
		if move.To == SquareC8 {
			pos.Remove(SquareA8, BlackRook)
			pos.Put(SquareD8, BlackRook)
		}
		if move.To == SquareG8 {
			pos.Remove(SquareH8, BlackRook)
			pos.Put(SquareF8, BlackRook)
		}
	}

	// Set Enpassant square for capturing.
	if pi.Figure() == Pawn &&
		move.From.Bitboard()&BbPawnStartRank != 0 &&
		move.To.Bitboard()&BbPawnDoubleRank != 0 {
		pos.Enpassant = (move.From + move.To) / 2
	} else {
		pos.Enpassant = SquareA1
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

	// Modify the chess board.
	pos.Remove(move.From, pi)
	pos.Remove(captSq, move.Capture)
	if move.MoveType != Promotion {
		pos.Put(move.To, pi)
	} else {
		pos.Put(move.To, move.Target)
	}
	pos.ToMove = pos.ToMove.Other()
}

// UndoMovePiece takes back a move.
// Expects the move to be valid.
// pi must be the piece moved, i.e. the pawn in case of promotions.
func (pos *Position) UndoMove(move Move) {
	pos.ToMove = pos.ToMove.Other()
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
		if move.To == SquareC1 {
			pos.Remove(SquareD1, WhiteRook)
			pos.Put(SquareA1, WhiteRook)
		}
		if move.To == SquareG1 {
			pos.Remove(SquareF1, WhiteRook)
			pos.Put(SquareH1, WhiteRook)
		}
		if move.To == SquareC8 {
			pos.Remove(SquareD8, BlackRook)
			pos.Put(SquareA8, BlackRook)
		}
		if move.To == SquareG8 {
			pos.Remove(SquareF8, BlackRook)
			pos.Put(SquareH8, BlackRook)
		}
	}

	pos.Castle = move.OldCastle
	pos.Enpassant = move.OldEnpassant
}

func (pos *Position) genPawnPromotions(from, to Square, capt Piece, moves []Move) []Move {
	pr := to.Rank()
	if pr != 0 && pr != 7 {
		mvtp := Normal
		if to == pos.Enpassant {
			mvtp = Enpassant
		}
		moves = append(moves, pos.fix(Move{
			MoveType: mvtp,
			From:     from,
			To:       to,
			Capture:  capt,
			Target:   ColorFigure(pos.ToMove, Pawn),
		}))
		return moves
	}

	for _, promo := range pawnPromotions {
		moves = append(moves, pos.fix(Move{
			MoveType: Promotion,
			From:     from,
			To:       to,
			Capture:  capt,
			Target:   ColorFigure(pos.ToMove, promo),
		}))
	}
	return moves
}

// GenPawnAdvanceMoves moves pawns one square.
func (pos *Position) genPawnAdvanceMoves(moves []Move) []Move {
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
		moves = pos.genPawnPromotions(from, to, NoPiece, moves)
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
		moves = pos.genPawnPromotions(from, to, NoPiece, moves)
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
		moves = pos.genPawnPromotions(from, to, capt, moves)
	}

	// Right
	bbr := bb & BbPawnRightAttack & (enemy >> 1)
	att = forward.Relative(0, +1)
	for bbr > 0 {
		from := bbr.Pop()
		to := from + att
		capt := pos.Get(to)
		moves = pos.genPawnPromotions(from, to, capt, moves)
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
		moves = pos.genPawnPromotions(from, to, capt, moves)
	}

	// Right
	bbr := bb & BbPawnRightAttack & (enemy >> 1)
	if bbr != 0 {
		from := bbr.AsSquare()
		to := pos.Enpassant
		capt := ColorFigure(pos.ToMove.Other(), Pawn)
		moves = pos.genPawnPromotions(from, to, capt, moves)
	}

	return moves

}

// GenPawnMoves generates pawn moves around from.
func (pos *Position) genPawnMoves(moves []Move) []Move {
	moves = pos.genPawnAdvanceMoves(moves)
	moves = pos.genPawnDoubleAdvanceMoves(moves)
	moves = pos.genPawnEnpassantMoves(moves)
	moves = pos.genPawnAttackMoves(moves)
	return moves
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, moves []Move) []Move {
	for att != 0 {
		to := att.Pop()
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

func (pos *Position) genKnightMoves(moves []Move) []Move {
	pi := ColorFigure(pos.ToMove, Knight)
	for bb := pos.ByPiece(pos.ToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := BbKnightAttack[from] & (^pos.ByColor[pos.ToMove])
		moves = pos.genBitboardMoves(pi, from, att, moves)
	}
	return moves
}

func (pos *Position) genBishopMoves(moves []Move, fig Figure) []Move {
	pi := ColorFigure(pos.ToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.ToMove, fig); bb != 0; {
		from := bb.Pop()
		att := BishopMagic[from].Attack(ref) &^ pos.ByColor[pos.ToMove]
		moves = pos.genBitboardMoves(pi, from, att, moves)
	}
	return moves
}

func (pos *Position) genRookMoves(moves []Move, fig Figure) []Move {
	pi := ColorFigure(pos.ToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.ToMove, fig); bb != 0; {
		from := bb.Pop()
		att := RookMagic[from].Attack(ref) &^ pos.ByColor[pos.ToMove]
		moves = pos.genBitboardMoves(pi, from, att, moves)
	}
	return moves
}

func (pos *Position) genKingMoves(moves []Move) []Move {
	pi := ColorFigure(pos.ToMove, King)
	from := (pos.ByColor[pos.ToMove] & pos.ByFigure[King]).AsSquare()

	// King moves around.
	other := pos.ToMove.Other()
	att := BbKingAttack[from] & (^pos.ByColor[pos.ToMove])
	for att != 0 {
		if to := att.Pop(); !pos.IsAttackedBy(to, other) {
			moves = append(moves, pos.fix(Move{
				MoveType: Normal,
				From:     from,
				To:       to,
				Capture:  pos.Get(to),
				Target:   pi,
			}))
		}
	}

	// King Castles.
	oo, ooo, rank := WhiteOO, WhiteOOO, 0
	if pos.ToMove == Black {
		oo, ooo, rank = BlackOO, BlackOOO, 7
	}

	// Castle king side.
	r5 := RankFile(rank, 5)
	r6 := RankFile(rank, 6)
	if pos.Castle&oo != 0 && pos.IsEmpty(r5) && pos.IsEmpty(r6) {
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

	// Castle queen side.
	r3 := RankFile(rank, 3)
	r2 := RankFile(rank, 2)
	r1 := RankFile(rank, 1)
	if pos.Castle&ooo != 0 && pos.IsEmpty(r3) && pos.IsEmpty(r2) && pos.IsEmpty(r1) {
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

	return moves
}

// IsAttackedBy returns true if sq is under attacked by co.
func (pos *Position) IsAttackedBy(sq Square, co Color) bool {
	enemy := pos.ByColor[co]
	if BbPawnAttack[sq]&enemy&pos.ByFigure[Pawn] != 0 {
		bb := sq.Bitboard()
		pawns := pos.ByPiece(co, Pawn)
		pawnsLeft := BbPawnLeftAttack & pawns
		pawnsRight := BbPawnRightAttack & pawns

		if co == White { // WhitePawn
			if att := (bb>>7)&pawnsLeft | (bb>>9)&pawnsRight; att != 0 {
				return true
			}
		} else { // BlackPawn
			if att := (bb<<9)&pawnsLeft | (bb<<7)&pawnsRight; att != 0 {
				return true
			}
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

// GenerateMoves is a helper to generate moves in stages.
type MoveGenerator struct {
	position *Position
	state    int
}

func NewMoveGenerator(pos *Position) *MoveGenerator {
	return &MoveGenerator{
		position: pos,
		state:    0,
	}
}

// Next generates pseudo-legal moves,i.e. doesn't check for king check.
func (mg *MoveGenerator) Next(moves []Move) (Piece, []Move) {
	mg.state++
	toMove := mg.position.ToMove
	switch mg.state {
	case 1:
		moves = mg.position.genPawnMoves(moves)
		return ColorFigure(toMove, Pawn), moves
	case 2:
		moves = mg.position.genKnightMoves(moves)
		return ColorFigure(toMove, Knight), moves
	case 3:
		moves = mg.position.genBishopMoves(moves, Bishop)
		return ColorFigure(toMove, Bishop), moves
	case 4:
		moves = mg.position.genRookMoves(moves, Rook)
		return ColorFigure(toMove, Rook), moves
	case 5:
		moves = mg.position.genBishopMoves(moves, Queen)
		return ColorFigure(toMove, Queen), moves
	case 6:
		moves = mg.position.genRookMoves(moves, Queen)
		return ColorFigure(toMove, Queen), moves
	case 7:
		moves = mg.position.genKingMoves(moves)
		return ColorFigure(toMove, King), moves
	default:
		return NoPiece, moves
	}
}

// GenerateMoves is a helper to generate all moves.
func (pos *Position) GenerateMoves(moves []Move) []Move {
	moveGen := NewMoveGenerator(pos)
	for piece := WhitePawn; piece != NoPiece; {
		piece, moves = moveGen.Next(moves)
	}
	return moves
}
