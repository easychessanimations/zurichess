package engine

import (
	"fmt"
	"log"
)

var (
	_ = log.Println

	// Which castle rights are lost when pieces are moved.
	lostCastleRights [64]Castle
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
	byFigure  [FigureMaxValue]Bitboard
	byColor   [ColorMaxValue]Bitboard
	ToMove    Color
	Castle    Castle
	Enpassant Square
}

// Put puts a piece on the board.
// Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	bb := sq.Bitboard()
	pos.byColor[pi.Color()] |= bb
	pos.byFigure[pi.Figure()] |= bb
}

// Remove removes a piece from the table.
// Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	bb := ^sq.Bitboard()
	pos.byColor[pi.Color()] &= bb
	pos.byFigure[pi.Figure()] &= bb
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.byColor[White]|pos.byColor[Black])>>sq&1 == 0
}

// GetColor returns the piece's color at sq.
func (pos *Position) GetColor(sq Square) Color {
	return White*Color(pos.byColor[White]>>sq&1) +
		Black*Color(pos.byColor[Black]>>sq&1)
}

// GetFigure returns the piece's type at sq.
func (pos *Position) GetFigure(sq Square) Figure {
	for pt := FigureMinValue; pt < FigureMaxValue; pt++ {
		if pos.byFigure[pt]&sq.Bitboard() != 0 {
			return pt
		}
	}
	return NoFigure
}

// Get returns the piece at sq.
func (pos *Position) Get(sq Square) Piece {
	co := pos.GetColor(sq)
	if co == NoColor {
		return NoPiece
	}
	pt := pos.GetFigure(sq)
	return ColorFigure(co, pt)
}

// IsChecked returns true if co's king is checked.
func (pos *Position) IsChecked(co Color) bool {
	kingSq := (pos.byColor[co] & pos.byFigure[King]).AsSquare()
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
	promo := NoPiece

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
		MoveType:  mvtp,
		From:      from,
		To:        to,
		Capture:   capt,
		Promotion: promo,
	})
}

// DoMove performs a move.
// Expects the move to be valid.
// TODO: promotion
func (pos *Position) DoMove(move Move) {
	pi := pos.Get(move.From)
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
		pos.Put(move.To, move.Promotion)
	}
	pos.ToMove = pos.ToMove.Other()
}

// UndoMove takes back a move.
// Expects the move to be valid.
func (pos *Position) UndoMove(move Move) {
	// log.Println("Taking back", move)

	pos.ToMove = pos.ToMove.Other()

	captSq := move.To
	if move.MoveType == Enpassant {
		captSq = RankFile(move.From.Rank(), move.To.File())
	}

	// Modify the chess board.
	pi := pos.Get(move.To)
	pos.Remove(move.To, pi)
	if move.MoveType != Promotion {
		pos.Put(move.From, pi)
	} else {
		pos.Put(move.From, ColorFigure(pos.ToMove, Pawn))
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

var (
	pawnAttack = []struct {
		delta  int      // difference on delta
		attack Bitboard // allowed start position for attack
	}{
		{-1, BbPawnLeftAttack},
		{+1, BbPawnRightAttack},
	}
	pawnPromotions = []Figure{Knight, Bishop, Rook, Queen}
)

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
		}))
	} else {
		for _, promo := range pawnPromotions {
			moves = append(moves, pos.fix(Move{
				MoveType:  Promotion,
				From:      from,
				To:        to,
				Capture:   capt,
				Promotion: ColorFigure(pos.ToMove, promo),
			}))
		}
	}
	return moves
}

// genPawnAdvanceMoves moves pawns one square.
func (pos *Position) genPawnAdvanceMoves(moves []Move) []Move {
	bb := pos.byColor[pos.ToMove] & pos.byFigure[Pawn]
	free := ^(pos.byColor[White] | pos.byColor[Black])
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

// genPawnAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(moves []Move) []Move {
	bb := pos.byColor[pos.ToMove] & pos.byFigure[Pawn]
	free := ^(pos.byColor[White] | pos.byColor[Black])
	var forward Square

	if pos.ToMove == White {
		bb &= BbRank2
		bb &= free >> 8
		bb &= free >> 16
		forward = RankFile(+2, 0)
	} else {
		bb &= BbRank7
		bb &= free << 8
		bb &= free << 16
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
	bb := pos.byColor[pos.ToMove] & pos.byFigure[Pawn]
	enemy := pos.byColor[pos.ToMove.Other()]
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

// genPawnMoves generates pawn moves around from.
func (pos *Position) genPawnMoves(from Square, moves []Move) []Move {
	advance := 1
	if pos.ToMove == Black {
		advance = -1
	}

	// Attack.
	for _, pa := range pawnAttack {
		if from.Bitboard()&pa.attack != 0 {
			var capt Piece
			to := from.Relative(advance, pa.delta)
			if pos.Enpassant != SquareA1 && to == pos.Enpassant {
				// Captures en passant.
				capt = ColorFigure(pos.ToMove.Other(), Pawn)
				moves = pos.genPawnPromotions(from, to, capt, moves)
			}
		}
	}

	return moves
}

func (pos *Position) genBitboardMoves(from Square, att Bitboard, moves []Move) []Move {
	for att != 0 {
		to := att.Pop()
		moves = append(moves, pos.fix(Move{
			From:    from,
			To:      to,
			Capture: pos.Get(to),
		}))
	}
	return moves
}

// genKnightMoves generates knight moves around from.
func (pos *Position) genKnightMoves(from Square, moves []Move) []Move {
	att := BbKnightAttack[from] & (^pos.byColor[pos.ToMove])
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genRookMoves(from Square, moves []Move) []Move {
	ref := pos.byColor[White] | pos.byColor[Black]
	att := RookMagic[from].Attack(ref) &^ pos.byColor[pos.ToMove]
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genBishopMoves(from Square, moves []Move) []Move {
	ref := pos.byColor[White] | pos.byColor[Black]
	att := BishopMagic[from].Attack(ref) &^ pos.byColor[pos.ToMove]
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genQueenMoves(from Square, moves []Move) []Move {
	moves = pos.genRookMoves(from, moves)
	moves = pos.genBishopMoves(from, moves)
	return moves
}

func (pos *Position) genKingMoves(from Square, moves []Move) []Move {
	// King moves around.
	other := pos.ToMove.Other()
	att := BbKingAttack[from] & (^pos.byColor[pos.ToMove])
	for att != 0 {
		if to := att.Pop(); !pos.IsAttackedBy(to, other) {
			moves = append(moves, pos.fix(Move{
				From:    from,
				To:      to,
				Capture: pos.Get(to),
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
			}))
		}
	}

	return moves
}

// GenerateMoves generates pseudo-legal moves, i.e. doesn't
// check for king check.
func (pos *Position) GenerateMoves(moves []Move) []Move {
	moves = pos.genPawnAdvanceMoves(moves)
	moves = pos.genPawnDoubleAdvanceMoves(moves)
	moves = pos.genPawnAttackMoves(moves)

	for from := SquareMinValue; from < SquareMaxValue; from++ {
		if pos.byColor[pos.ToMove]&from.Bitboard() == 0 {
			continue
		}

		pi := pos.Get(from)
		switch pi.Figure() {
		case Pawn:
			moves = pos.genPawnMoves(from, moves)
		case Knight:
			moves = pos.genKnightMoves(from, moves)
		case Bishop:
			moves = pos.genBishopMoves(from, moves)
		case Rook:
			moves = pos.genRookMoves(from, moves)
		case Queen:
			moves = pos.genQueenMoves(from, moves)
		case King:
			moves = pos.genKingMoves(from, moves)
		}
	}
	return moves
}

// IsAttackedBy returns true if sq is under attacked by co.
func (pos *Position) IsAttackedBy(sq Square, co Color) bool {
	// WhitePawn
	if co == White {
		pawns := pos.byColor[White] & pos.byFigure[Pawn]
		if sq.Bitboard()&(BbPawnLeftAttack<<7) != 0 {
			if sq.Relative(-1, +1).Bitboard()&pawns != 0 {
				return true
			}
		}
		if sq.Bitboard()&(BbPawnRightAttack<<9) != 0 {
			if sq.Relative(-1, -1).Bitboard()&pawns != 0 {
				return true
			}
		}
	}

	// BlackPawn
	if co == Black {
		pawns := pos.byColor[Black] & pos.byFigure[Pawn]
		if sq.Bitboard()&(BbPawnLeftAttack>>9) != 0 {
			if sq.Relative(+1, +1).Bitboard()&pawns != 0 {
				return true
			}
		}
		if sq.Bitboard()&(BbPawnRightAttack>>7) != 0 {
			if sq.Relative(+1, -1).Bitboard()&pawns != 0 {
				return true
			}
		}
	}

	all := pos.byColor[White] | pos.byColor[Black]
	enemy := pos.byColor[co]

	// Quick test of SuperPiece on empty board.
	if BbSuperAttack[sq]&(enemy&^pos.byFigure[Pawn]) == 0 {
		return false
	}

	// Knight
	if BbKnightAttack[sq]&enemy&pos.byFigure[Knight] != 0 {
		return true
	}

	// King.
	if BbKingAttack[sq]&enemy&pos.byFigure[King] != 0 {
		return true
	}

	// Rook&Queen
	rooks := enemy & (pos.byFigure[Rook] | pos.byFigure[Queen])
	if rooks != 0 && rooks&RookMagic[sq].Attack(all) != 0 {
		return true
	}

	// Bishop&Queen
	bishops := enemy & (pos.byFigure[Bishop] | pos.byFigure[Queen])
	if bishops != 0 && bishops&BishopMagic[sq].Attack(all) != 0 {
		return true
	}

	return false
}
