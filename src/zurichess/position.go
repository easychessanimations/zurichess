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

// ParseMove parses a move given in standard algebraic notation.
// s can be "a2a4" or "h7h8Q" (pawn promotion).
func (pos *Position) ParseMove(s string) Move {
	from := SquareFromString(s[0:2])
	to := SquareFromString(s[2:4])

	return Move{
		MoveType:  Normal, // TODO
		From:      from,
		To:        to,
		Capture:   pos.Get(to),
		OldCastle: pos.castle,
	}
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
		panic(fmt.Errorf("expected %v piece at %v, got %v", pos.toMove, move.From, pi))
	}

	// log.Println(pos.Get(move.From), "playing", move, "; castling rights", pos.castle)

	// Update castling rights.
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

	// Modify the chess board.
	pos.Remove(move.From, pi)
	pos.Remove(move.To, move.Capture)
	pos.Put(move.To, pi)
	pos.toMove = pos.toMove.Other()
}

// UndoMove takes back a move.
// Expects the move to be valid.
// TODO: promotion
func (pos *Position) UndoMove(move Move) {
	// log.Println("Takeing back", move)

	// Modify the chess board.
	pi := pos.Get(move.To)
	pos.Remove(move.To, pi)
	pos.Put(move.From, pi)
	pos.Put(move.To, move.Capture)
	pos.toMove = pos.toMove.Other()

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

	// Restore castling rights.
	pos.castle = move.OldCastle
}

// genPawnMoves generates pawn moves around from.
func (pos *Position) genPawnMoves(from Square, pi Piece, moves []Move) []Move {
	advance, pawnRank, lastRank := 1, 1, 6
	if pi.Color() == Black {
		advance, pawnRank, lastRank = -1, 6, 1
	}

	pr := from.Rank()
	f1 := from.Relative(advance, 0)

	// Move forward.
	if pr != lastRank {
		if pos.IsEmpty(f1) {
			moves = append(moves, Move{
				From:      from,
				To:        f1,
				OldCastle: pos.castle,
			})
		}
	}

	// Move forward 2x.
	if pr == pawnRank {
		f2 := from.Relative(advance*2, 0)

		if pos.IsEmpty(f1) && pos.IsEmpty(f2) {
			moves = append(moves, Move{
				From:      from,
				To:        f2,
				OldCastle: pos.castle,
			})
		}
	}

	// Attack left.
	if pr != lastRank && from.File() != 0 {
		to := from.Relative(advance, -1)
		c := pos.Get(to)
		if c.Color() == pi.Color().Other() {
			moves = append(moves, Move{
				From:      from,
				To:        to,
				Capture:   c,
				OldCastle: pos.castle,
			})
		}
	}

	// Attack right.
	if pr != lastRank && from.File() != 7 {
		to := from.Relative(advance, +1)
		c := pos.Get(to)
		if c.Color() == pi.Color().Other() {
			moves = append(moves, Move{
				From:      from,
				To:        to,
				Capture:   c,
				OldCastle: pos.castle,
			})
		}
	}

	// TODO promote

	return moves
}

var (
	knightJump = [8][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
)

// genKnightMoves generates knight moves around from.
func (pos *Position) genKnightMoves(from Square, pi Piece, moves []Move) []Move {
	for _, e := range knightJump {
		r, f := from.Rank()+e[0], from.File()+e[1]
		if 0 > r || r >= 8 || 0 > f || f >= 8 {
			// Cannot jump out of the table.
			continue
		}
		to := RankFile(r, f)

		capture := pos.Get(to)
		if capture.Color() == pi.Color() {
			// Cannot capture same color.
			continue
		}

		// Found a valid Knight move.
		moves = append(moves, Move{
			From:      from,
			To:        to,
			Capture:   capture,
			OldCastle: pos.castle,
		})
	}
	return moves
}

var limit = [3]int{-1, -1, 8}

func (pos *Position) genSlidingMoves(from Square, pi Piece, dr, df int, moves []Move) []Move {
	r, f := from.Rank(), from.File()
	lr := limit[dr+1]
	lf := limit[df+1]

	for {
		r, f = r+dr, f+df
		if r == lr || f == lf {
			// Stop when outside board.
			break
		}
		to := RankFile(r, f)

		// Check the captured piece.
		capture := pos.Get(to)
		if pi.Color() == capture.Color() {
			break
		}

		moves = append(moves, Move{
			From:      from,
			To:        to,
			Capture:   pos.Get(to),
			OldCastle: pos.castle,
		})

		// Stop if there a piece in the way.
		if capture.Color() != NoColor {
			break
		}
	}

	return moves
}

func (pos *Position) genRookMoves(from Square, pi Piece, moves []Move) []Move {
	moves = pos.genSlidingMoves(from, pi, +1, 0, moves)
	moves = pos.genSlidingMoves(from, pi, -1, 0, moves)
	moves = pos.genSlidingMoves(from, pi, 0, +1, moves)
	moves = pos.genSlidingMoves(from, pi, 0, -1, moves)
	return moves
}

func (pos *Position) genBishopMoves(from Square, pi Piece, moves []Move) []Move {
	moves = pos.genSlidingMoves(from, pi, +1, -1, moves)
	moves = pos.genSlidingMoves(from, pi, -1, -1, moves)
	moves = pos.genSlidingMoves(from, pi, +1, +1, moves)
	moves = pos.genSlidingMoves(from, pi, -1, +1, moves)
	return moves
}

func (pos *Position) genQueenMoves(from Square, pi Piece, moves []Move) []Move {
	moves = pos.genRookMoves(from, pi, moves)
	moves = pos.genBishopMoves(from, pi, moves)
	return moves
}

var (
	kingDRank = [8]int{+1, +1, +1, +0, -1, -1, -1, +0}
	kingDFile = [8]int{-1, +0, +1, +1, +1, +0, -1, -1}
)

func (pos *Position) genKingMoves(from Square, pi Piece, moves []Move) []Move {
	if pi.Color() != pos.toMove {
		panic(fmt.Errorf("expected %v to move, got %v",
			pos.toMove, pi.Color()))
	}

	// King moves around.
	for i := 0; i < 8; i++ {
		dr := kingDRank[i]
		df := kingDFile[i]

		r, f := from.Rank()+dr, from.File()+df
		if r == -1 || r == 8 || f == -1 || f == 8 {
			// Stop when outside board.
			continue
		}
		to := RankFile(r, f)

		// Check the captured piece.
		capture := pos.Get(to)
		if pi.Color() == capture.Color() {
			continue
		}

		moves = append(moves, Move{
			From:      from,
			To:        to,
			Capture:   pos.Get(to),
			OldCastle: pos.castle,
		})
	}

	// King castles.
	// TODO: verify checks
	oo, ooo, rank := WhiteOO, WhiteOOO, 0
	if pi.Color() == Black {
		oo, ooo, rank = BlackOO, BlackOOO, 7
	}

	// Castle king side.
	if pos.castle&oo != 0 {
		if pos.IsEmpty(RankFile(rank, 5)) &&
			pos.IsEmpty(RankFile(rank, 6)) {
			moves = append(moves, Move{
				MoveType:  Castling,
				From:      from,
				To:        RankFile(rank, 6),
				OldCastle: pos.castle,
			})
		}
	}

	// Castle queen side.
	if pos.castle&ooo != 0 {
		if pos.IsEmpty(RankFile(rank, 3)) &&
			pos.IsEmpty(RankFile(rank, 2)) &&
			pos.IsEmpty(RankFile(rank, 1)) {
			moves = append(moves, Move{
				MoveType:  Castling,
				From:      from,
				To:        RankFile(rank, 2),
				OldCastle: pos.castle,
			})
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
			moves = pos.genPawnMoves(sq, pi, moves)
		case Knight:
			moves = pos.genKnightMoves(sq, pi, moves)
		case Bishop:
			moves = pos.genBishopMoves(sq, pi, moves)
		case Rook:
			moves = pos.genRookMoves(sq, pi, moves)
		case Queen:
			moves = pos.genQueenMoves(sq, pi, moves)
		case King:
			moves = pos.genKingMoves(sq, pi, moves)
		}
	}
	return moves
}
