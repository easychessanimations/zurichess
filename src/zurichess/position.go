package main

import (
	"fmt"
	"strings"
)

type Position struct {
	byPieceType [PieceTypeMaxValue]Bitboard
	byColor     [ColorMaxValue]Bitboard
	toMove      Color
	castle      Castle
	enpassant   Square
}

func PositionFromFEN(fen string) (*Position, error) {
	fld := strings.Fields(fen)
	if len(fld) < 4 {
		return nil, fmt.Errorf("expected at least 4 fields, got %d", len(fld))
	}

	pos := &Position{}

	// Parse position.
	ranks := strings.Split(fld[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("expected 8 rows, got %d", len(ranks))
	}
	for r := range ranks {
		sq := RankFile(7-r, 0) // FEN describes the table from 8th rank.
		for _, p := range ranks[r] {
			pi := NoPiece
			switch p {
			case 'p':
				pi = ColorPiece(Black, Pawn)
			case 'n':
				pi = ColorPiece(Black, Knight)
			case 'b':
				pi = ColorPiece(Black, Bishop)
			case 'r':
				pi = ColorPiece(Black, Rook)
			case 'q':
				pi = ColorPiece(Black, Queen)
			case 'k':
				pi = ColorPiece(Black, King)

			case 'P':
				pi = ColorPiece(White, Pawn)
			case 'N':
				pi = ColorPiece(White, Knight)
			case 'B':
				pi = ColorPiece(White, Bishop)
			case 'R':
				pi = ColorPiece(White, Rook)
			case 'Q':
				pi = ColorPiece(White, Queen)
			case 'K':
				pi = ColorPiece(White, King)

			case '1', '2', '3', '4', '5', '6', '7', '8':
				sq = sq.Relative(0, int(p)-int('0')-1)

			default:
				return nil, fmt.Errorf("unhandled '%c'", p)
			}
			pos.PutPiece(sq, pi)
			sq = sq.Relative(0, 1)
		}
	}

	// Parse next to move.
	switch fld[1] {
	case "w":
		pos.toMove = White
	case "b":
		pos.toMove = Black
	default:
		return nil, fmt.Errorf("unknown color %s", fld[1])
	}

	// Parse castling rights.
	for _, p := range fld[2] {
		switch p {
		case 'K':
			pos.castle |= WhiteOO
		case 'Q':
			pos.castle |= WhiteOOO
		case 'k':
			pos.castle |= BlackOO
		case 'q':
			pos.castle |= BlackOOO
		}
	}

	// Parse enpassant.
	// TODO: handle error
	if fld[3] != "-" {
		pos.enpassant = SquareFromString(fld[3])
	}

	// TODO: halfmove, fullmove
	return pos, nil
}

// PutPiece puts a piece on the board.
// Does not do any checks.
func (pos *Position) PutPiece(sq Square, pi Piece) {
	pos.byColor[pi.Color()] |= sq.Bitboard()
	pos.byPieceType[pi.PieceType()] |= sq.Bitboard()
}

// RemovePiece removes a piece from the table.
func (pos *Position) RemovePiece(sq Square, pi Piece) {
	pos.byColor[pi.Color()] &= ^sq.Bitboard()
	pos.byPieceType[pi.PieceType()] &= ^sq.Bitboard()
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.byColor[White]|pos.byColor[Black])&sq.Bitboard() == 0
}

// GetPiece returns the piece at sq.
func (pos *Position) GetPiece(sq Square) Piece {
	var co Color
	var pt PieceType

	for co = ColorMinValue; co < ColorMaxValue; co++ {
		if pos.byColor[co]&sq.Bitboard() != 0 {
			break
		}
	}
	if co == ColorMaxValue {
		return ColorPiece(NoColor, NoPieceType)
	}

	for pt = PieceTypeMinValue; pt < PieceTypeMaxValue; pt++ {
		if pos.byPieceType[pt]&sq.Bitboard() != 0 {
			break
		}
	}
	if pt == PieceTypeMaxValue {
		panic("expected piece, got nothing")
	}

	return ColorPiece(co, pt)
}

// PrettyPrints pretty prints the current position.
func (pos *Position) PrettyPrint() {
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 0; f < 8; f++ {
			line += pos.GetPiece(RankFile(r, f)).Symbol()
		}
		if r == 7 && pos.toMove == Black {
			line += " *"
		}
		if r == 0 && pos.toMove == White {
			line += " *"
		}
	}

}

func (pos *Position) ParseMove(s string) Move {
	from := SquareFromString(s[0:2])
	to := SquareFromString(s[2:4])

	return Move{
		From:     from,
		To:       to,
		Capture:  pos.GetPiece(to),
		MoveType: Normal, // TODO
	}
}

// DoMove performs a move.
// Expects the move to be valid.
// TODO: castling, promotion
func (pos *Position) DoMove(mo Move) {
	// log.Println("Playing", mo)
	piece := pos.GetPiece(mo.From)
	pos.RemovePiece(mo.From, piece)
	pos.RemovePiece(mo.To, mo.Capture)
	pos.PutPiece(mo.To, piece)
	pos.toMove = pos.toMove.Other()
}

// UndoMove takes back a move.
// Expects the move to be valid.
// TODO: castling, promotion
func (pos *Position) UndoMove(mo Move) {
	// log.Println("Takeing back", mo)
	piece := pos.GetPiece(mo.To)
	pos.RemovePiece(mo.To, piece)
	pos.PutPiece(mo.From, piece)
	pos.PutPiece(mo.To, mo.Capture)
	pos.toMove = pos.toMove.Other()
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
				From: from,
				To:   f1,
			})
		}
	}

	// Move forward 2x.
	if pr == pawnRank {
		f2 := from.Relative(advance*2, 0)

		if pos.IsEmpty(f1) && pos.IsEmpty(f2) {
			moves = append(moves, Move{
				From: from,
				To:   f2,
			})
		}
	}

	// Attack left.
	if pr != lastRank && from.File() != 0 {
		to := from.Relative(advance, -1)
		c := pos.GetPiece(to)
		if c.Color() == pi.Color().Other() {
			moves = append(moves, Move{
				From:    from,
				To:      to,
				Capture: c,
			})
		}
	}

	// Attack right.
	if pr != lastRank && from.File() != 7 {
		to := from.Relative(advance, +1)
		c := pos.GetPiece(to)
		if c.Color() == pi.Color().Other() {
			moves = append(moves, Move{
				From:    from,
				To:      to,
				Capture: c,
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

		capture := pos.GetPiece(to)
		if capture.Color() == pi.Color() {
			// Cannot capture same color.
			continue
		}

		// Found a valid Knight move.
		moves = append(moves, Move{
			From:     from,
			To:       to,
			Capture:  capture,
			MoveType: Normal,
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
		capture := pos.GetPiece(to)
		if pi.Color() == capture.Color() {
			break
		}

		moves = append(moves, Move{
			From:    from,
			To:      to,
			Capture: pos.GetPiece(to),
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
		capture := pos.GetPiece(to)
		if pi.Color() == capture.Color() {
			continue
		}

		moves = append(moves, Move{
			From:    from,
			To:      to,
			Capture: pos.GetPiece(to),
		})
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

		pi := pos.GetPiece(sq)
		switch pi.PieceType() {
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
