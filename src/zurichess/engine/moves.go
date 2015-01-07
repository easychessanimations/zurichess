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
// * x (capture) presence or correctness is ignored.
// * + (check) and # (checkmate) is ignored.
// * e.p. (enpassant) is ignored
//
// TODO: Handle castling.
func (pos *Position) SANToMove(s string) (Move, error) {
	b, e := 0, len(s)
	m := Move{MoveType: Normal}

	// Get the piece.
	if b == e {
		return Move{}, errorWrongLength
	}
	piece := NoPiece
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
	m.Target = piece

	// Skips + (check) and # (checkmate) at the end.
	for e > b && (s[e-1] == '#' || s[e-1] == '+') {
		e--
	}
	// Skips e.p. when enpassant.
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
			m.MoveType = Promotion
			m.Target = ColorFigure(pos.ToMove, fig)
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
	m.To = SquareFromString(s[e-2 : e])
	if m.To == pos.Enpassant {
		m.MoveType = Enpassant
		m.Capture = ColorFigure(pos.ToMove.Other(), Pawn)
	} else {
		m.Capture = pos.Get(m.To)
	}
	e -= 2

	// Ignore capture info.
	if e-1 >= b && s[e-1] == 'x' {
		e--
	}

	// Parse disambiguation.
	r, f := -1, -1
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

	// Loop through all moves and find out one that matches.
	moves := pos.GenerateFigureMoves(piece.Figure(), nil)
	for _, pm := range moves {
		if pm.MoveType != m.MoveType || pm.Capture != m.Capture {
			continue
		}
		if pm.To != m.To || pm.Target != m.Target {
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
// incorrectly states that this is long algebraic notation (LAN).
func (pos *Position) MoveToUCI(m Move) string {
	r := m.From.String() + m.To.String()
	if m.MoveType == Promotion {
		r += string(pieceToSymbol[m.Target])
	}
	return r
}

// UCIToMove parses a move given in UCI format.
// s can be "a2a4" or "h7h8Q" (pawn promotion).
func (pos *Position) UCIToMove(s string) Move {
	from := SquareFromString(s[0:2])
	to := SquareFromString(s[2:4])

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
