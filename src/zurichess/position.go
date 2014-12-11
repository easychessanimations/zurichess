package main

import (
	"fmt"
	"log"
)

var _ = log.Println

// Position encodes the chess board.
type Position struct {
	byFigure  [FigureMaxValue]Bitboard
	byColor   [ColorMaxValue]Bitboard
	toMove    Color
	castle    Castle
	enpassant Square
}

// Put puts a piece on the board.
// Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	pos.byColor[pi.Color()] |= sq.Bitboard()
	pos.byFigure[pi.Figure()] |= sq.Bitboard()
}

// Remove removes a piece from the table.
// Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	pos.byColor[pi.Color()] &= ^sq.Bitboard()
	pos.byFigure[pi.Figure()] &= ^sq.Bitboard()
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.byColor[White]|pos.byColor[Black])&sq.Bitboard() == 0
}

// GetColor returns the piece's color at sq.
func (pos *Position) GetColor(sq Square) Color {
	if pos.byColor[White]&sq.Bitboard() != 0 {
		return White
	}
	if pos.byColor[Black]&sq.Bitboard() != 0 {
		return Black
	}
	return NoColor
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
		if r == 7 && pos.toMove == Black {
			line += " *"
		}
		if r == 0 && pos.toMove == White {
			line += " *"
		}
		log.Println(line)
	}

}

// fix updates move so that it can be undone.
// Does not set capture.
func (pos *Position) fix(move Move) Move {
	move.OldCastle = pos.castle
	move.OldEnpassant = pos.enpassant
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
	if pi.Figure() == Pawn && pos.enpassant != SquareA1 && to == pos.enpassant {
		mvtp = Enpassant
		capt = ColorFigure(pos.toMove.Other(), Pawn)
	}
	if pi == WhiteKing && from == SquareE1 && (to == SquareC1 || to == SquareG1) {
		mvtp = Castling
	}
	if pi == BlackKing && from == SquareE8 && (to == SquareC8 || to == SquareG8) {
		mvtp = Castling
	}
	if pi.Figure() == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		mvtp = Promotion
		promo = ColorFigure(pos.toMove, symbolToFigure[s[4]])
	}

	return pos.fix(Move{
		MoveType:  mvtp,
		From:      from,
		To:        to,
		Capture:   capt,
		Promotion: promo,
	})
}

var castleRights = map[Square]Castle{
	SquareA1: WhiteOOO,
	SquareE1: WhiteOOO | WhiteOO,
	SquareH1: WhiteOO,
	SquareA8: BlackOOO,
	SquareE8: BlackOOO | BlackOO,
	SquareH8: BlackOO,
}

// DoMove performs a move.
// Expects the move to be valid.
// TODO: promotion
func (pos *Position) DoMove(move Move) {
	pi := pos.Get(move.From)
	if pi.Color() != pos.toMove {
		panic(fmt.Errorf("bad move: expected %v piece at %v, got %v",
			pos.toMove, move.From, pi))
	}

	/*
		log.Println(
			pos.Get(move.From), "playing", move,
			"; castling ", pos.castle,
			"; enpassant", pos.enpassant)
	*/

	// Update castling rights based on the source square.
	pos.castle &= ^castleRights[move.From]

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

	// Set enpassant square for capturing.
	if pi.Figure() == Pawn &&
		move.From.Bitboard()&BbPawnStartRank != 0 &&
		move.To.Bitboard()&BbPawnDoubleRank != 0 {
		pos.enpassant = (move.From + move.To) / 2
	} else {
		pos.enpassant = SquareA1
	}

	// Capture pawn enpassant.
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
	pos.toMove = pos.toMove.Other()
}

// UndoMove takes back a move.
// Expects the move to be valid.
func (pos *Position) UndoMove(move Move) {
	// log.Println("Taking back", move)

	pos.toMove = pos.toMove.Other()

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
		pos.Put(move.From, ColorFigure(pos.toMove, Pawn))
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

	pos.castle = move.OldCastle
	pos.enpassant = move.OldEnpassant
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
		if to == pos.enpassant {
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
				Promotion: ColorFigure(pos.toMove, promo),
			}))
		}
	}
	return moves
}

// genPawnMoves generates pawn moves around from.
func (pos *Position) genPawnMoves(from Square, moves []Move) []Move {
	advance, pawnRank := 1, 1
	if pos.toMove == Black {
		advance, pawnRank = -1, 6
	}

	pr := from.Rank()

	// Move forward.
	f1 := from.Relative(advance, 0)
	if pos.IsEmpty(f1) {
		moves = pos.genPawnPromotions(from, f1, NoPiece, moves)
	}

	// Move forward 2x.
	if pr == pawnRank {
		f2 := from.Relative(advance*2, 0)
		if pos.IsEmpty(f1) && pos.IsEmpty(f2) {
			moves = append(moves, pos.fix(Move{
				From: from,
				To:   f2,
			}))
		}
	}

	for _, pa := range pawnAttack {
		if from.Bitboard()&pa.attack != 0 {
			var capt Piece
			to := from.Relative(advance, pa.delta)
			if pos.enpassant != SquareA1 && to == pos.enpassant {
				// Captures en passant.
				capt = ColorFigure(pos.toMove.Other(), Pawn)
			} else {
				// Regular capture.
				capt = pos.Get(to)
				if capt.Color() != pos.toMove.Other() {
					continue
				}
			}
			moves = pos.genPawnPromotions(from, to, capt, moves)
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
	att := BbKnightAttack[from] & (^pos.byColor[pos.toMove])
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genRookMoves(from Square, moves []Move) []Move {
	ref := pos.byColor[White] | pos.byColor[Black]
	att := RookMagic[from].Attack(ref) &^ pos.byColor[pos.toMove]
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genBishopMoves(from Square, moves []Move) []Move {
	ref := pos.byColor[White] | pos.byColor[Black]
	att := BishopMagic[from].Attack(ref) &^ pos.byColor[pos.toMove]
	return pos.genBitboardMoves(from, att, moves)
}

func (pos *Position) genQueenMoves(from Square, moves []Move) []Move {
	moves = pos.genRookMoves(from, moves)
	moves = pos.genBishopMoves(from, moves)
	return moves
}

func (pos *Position) genKingMoves(from Square, moves []Move) []Move {
	// King moves around.
	att := BbKingAttack[from] & (^pos.byColor[pos.toMove])
	moves = pos.genBitboardMoves(from, att, moves)

	// King castles.
	// TODO: verify checks
	oo, ooo, rank := WhiteOO, WhiteOOO, 0
	if pos.toMove == Black {
		oo, ooo, rank = BlackOO, BlackOOO, 7
	}

	// Castle king side.
	r5 := RankFile(rank, 5)
	r6 := RankFile(rank, 6)
	if pos.castle&oo != 0 {
		if pos.IsEmpty(r5) && pos.IsEmpty(r6) &&
			!pos.IsAttackedBy(r5, pos.toMove.Other()) &&
			!pos.IsAttackedBy(r6, pos.toMove.Other()) {
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
	if pos.castle&ooo != 0 {
		if pos.IsEmpty(r3) && pos.IsEmpty(r2) && pos.IsEmpty(r1) &&
			!pos.IsAttackedBy(r3, pos.toMove.Other()) &&
			!pos.IsAttackedBy(r2, pos.toMove.Other()) {
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
func (pos *Position) GenerateMoves() []Move {
	moves := make([]Move, 0, 8)
	for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
		if pos.byColor[pos.toMove]&sq.Bitboard() == 0 {
			continue
		}

		pi := pos.Get(sq)
		switch pi.Figure() {
		case Pawn:
			moves = pos.genPawnMoves(sq, moves)
		case Knight:
			moves = pos.genKnightMoves(sq, moves)
		case Bishop:
			moves = pos.genBishopMoves(sq, moves)
		case Rook:
			moves = pos.genRookMoves(sq, moves)
		case Queen:
			moves = pos.genQueenMoves(sq, moves)
		case King:
			moves = pos.genKingMoves(sq, moves)
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

	// Knight
	if BbKnightAttack[sq]&pos.byColor[co]&pos.byFigure[Knight] != 0 {
		return true
	}

	// King
	if BbKingAttack[sq]&pos.byColor[co]&pos.byFigure[King] != 0 {
		return true
	}

	// Rook&Queen
	ref := pos.byColor[White] | pos.byColor[Black]
	att := RookMagic[sq].Attack(ref)
	if att&pos.byColor[co]&(pos.byFigure[Rook]|pos.byFigure[Queen]) != 0 {
		return true
	}

	// Bishop&Queen
	ref = pos.byColor[White] | pos.byColor[Black]
	att = BishopMagic[sq].Attack(ref)
	if att&pos.byColor[co]&(pos.byFigure[Bishop]|pos.byFigure[Queen]) != 0 {
		return true
	}

	return false
}
