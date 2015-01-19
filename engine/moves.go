// moves.go deals with move parsing.

package engine

import (
	"fmt"
)

var (
	errorWrongLength       = fmt.Errorf("SAN string is too short")
	errorUnknownFigure     = fmt.Errorf("unknown figure symbol")
	errorBadDisambiguation = fmt.Errorf("bad disambiguation")
	errorBadPromotion      = fmt.Errorf("only pawns on the last rank can be promoted")
	errorNoSuchMove        = fmt.Errorf("no such move")
)

// SANToMove conversts a move to SAN format.
// SAN stand for standard algebraic notation and
// its description can be found in FIDE handbook.
//
// The set of strings accepted is a slightly different.
//   x (capture) presence or correctness is ignored.
//   + (check) and # (checkmate) is ignored.
//   e.p. (enpassant) is ignored
func (pos *Position) SANToMove(s string) (Move, error) {
	piece := NoPiece
	move := Move{MoveType: Normal}
	r, f := -1, -1

	// s[b:e] is the part that still needs to be parsed.
	b, e := 0, len(s)
	if b == e {
		return Move{}, errorWrongLength
	}
	// Skip + (check) and # (checkmate) at the end.
	for e > b && (s[e-1] == '#' || s[e-1] == '+') {
		e--
	}

	if s[b:e] == "o-o" || s[b:e] == "O-O" { // king side castling
		if pos.ToMove == White {
			move = Move{
				MoveType: Castling,
				From:     SquareE1,
				To:       SquareG1,
				Target:   WhiteKing,
			}
		} else {
			move = Move{
				MoveType: Castling,
				From:     SquareE8,
				To:       SquareG8,
				Target:   BlackKing,
			}
		}
		piece = move.Target
	} else if s[b:e] == "o-o-o" || s[b:e] == "O-O-O" { // queen side castling
		if pos.ToMove == White {
			move = Move{
				MoveType: Castling,
				From:     SquareE1,
				To:       SquareC1,
				Target:   WhiteKing,
			}
		} else {
			move = Move{
				MoveType: Castling,
				From:     SquareE8,
				To:       SquareC8,
				Target:   BlackKing,
			}
		}
		piece = move.Target
	} else { // all other moves
		// Get the piece.
		if ('a' <= s[b] && s[b] <= 'h') || s[b] == 'x' {
			piece = ColorFigure(pos.ToMove, Pawn)
		} else {
			if fig := symbolToFigure[rune(s[b])]; fig == NoFigure {
				return Move{}, errorUnknownFigure
			} else {
				piece = ColorFigure(pos.ToMove, fig)
			}
			b++
		}
		move.Target = piece

		// Skip e.p. when enpassant.
		if e-4 > b && s[e-4:e] == "e.p." {
			e -= 4
		}

		// Check pawn promotion.
		if e-1 < b {
			return Move{}, errorWrongLength
		}
		if !('1' <= s[e-1] && s[e-1] <= '8') {
			// Not a rank, but a promotion.
			if piece.Figure() != Pawn {
				return Move{}, errorBadPromotion
			}
			if fig := symbolToFigure[rune(s[e-1])]; fig == NoFigure {
				return Move{}, errorUnknownFigure
			} else {
				move.MoveType = Promotion
				move.Target = ColorFigure(pos.ToMove, fig)
			}
			e--
			if e-1 >= b && s[e-1] == '=' {
				// Sometimes = is inserted before promotion figure.
				e--
			}
		}

		// Handle destination square.
		if e-2 < b {
			return Move{}, errorWrongLength
		}
		var err error
		move.To, err = SquareFromString(s[e-2 : e])
		if err != nil {
			return Move{}, err
		}
		if move.To != SquareA1 && move.To == pos.Enpassant {
			move.MoveType = Enpassant
			move.Capture = ColorFigure(pos.ToMove.Other(), Pawn)
		} else {
			move.Capture = pos.Get(move.To)
		}
		e -= 2

		// Ignore 'x' (capture) or '-' (no capture) if present.
		if e-1 >= b && (s[e-1] == 'x' || s[e-1] == '-') {
			e--
		}

		// Parse disambiguation.
		if e-b > 2 {
			return Move{}, errorBadDisambiguation
		}
		for ; b < e; b++ {
			switch {
			case 'a' <= s[b] && s[b] <= 'h':
				f = int(s[b] - 'a')
			case '1' <= s[b] && s[b] <= '8':
				r = int(s[b] - '1')
			default:
				return Move{}, errorBadDisambiguation
			}
		}
	}

	// Loop through all moves and find out one that matches.
	moves := pos.GenerateFigureMoves(piece.Figure(), nil)
	for _, pm := range moves {
		if pm.MoveType != move.MoveType || pm.Capture != move.Capture {
			continue
		}
		if pm.To != move.To || pm.Target != move.Target {
			continue
		}
		if r != -1 && pm.From.Rank() != r {
			continue
		}
		if f != -1 && pm.From.File() != f {
			continue
		}
		return pm, nil
	}
	return Move{}, errorNoSuchMove
}

// MoveToUCI converts a move to UCI format.
// The protocol specification at http://wbec-ridderkerk.nl/html/UCIProtocol.html
// incorrectly states that this is the long algebraic notation (LAN).
func (pos *Position) MoveToUCI(move Move) string {
	r := move.From.String() + move.To.String()
	if move.MoveType == Promotion {
		r += string(pieceToSymbol[move.Target])
	}
	return r
}

// UCIToMove parses a move given in UCI format.
// s can be "a2a4" or "h7h8Q" for pawn promotion.
func (pos *Position) UCIToMove(s string) Move {
	from, _ := SquareFromString(s[0:2])
	to, _ := SquareFromString(s[2:4])

	moveType := Normal
	capt := pos.Get(to)
	promo := pos.Get(from)

	pi := pos.Get(from)
	if pi.Figure() == Pawn && pos.Enpassant != SquareA1 && to == pos.Enpassant {
		moveType = Enpassant
		capt = ColorFigure(pos.ToMove.Other(), Pawn)
	}
	if pi == WhiteKing && from == SquareE1 && (to == SquareC1 || to == SquareG1) {
		moveType = Castling
	}
	if pi == BlackKing && from == SquareE8 && (to == SquareC8 || to == SquareG8) {
		moveType = Castling
	}
	if pi.Figure() == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		moveType = Promotion
		promo = ColorFigure(pos.ToMove, symbolToFigure[rune(s[4])])
	}

	return pos.fix(Move{
		MoveType: moveType,
		From:     from,
		To:       to,
		Capture:  capt,
		Target:   promo,
	})
}
