package engine

import (
	"fmt"
	"log"
	"strconv"
)

var (
	// Which castle rights are lost when pieces are moved.
	lostCastleRights [64]Castle
)

func init() {
	lostCastleRights[SquareA1] = WhiteOOO
	lostCastleRights[SquareE1] = WhiteOOO | WhiteOO
	lostCastleRights[SquareH1] = WhiteOO
	lostCastleRights[SquareA8] = BlackOOO
	lostCastleRights[SquareE8] = BlackOOO | BlackOO
	lostCastleRights[SquareH8] = BlackOO
}

type state struct {
	Castle          Castle // remaining castling rights.
	EnpassantSquare Square // enpassant square. If none, then SquareA1.
	IrreversiblePly int    // highest square in which an irreversible move (cannot be part of repetition) made
	Zobrist         uint64
}

// Position encodes the chess board.
type Position struct {
	ByFigure   [FigureArraySize]Bitboard             // bitboards of square occupancy by figure.
	ByColor    [ColorArraySize]Bitboard              // bitboards of square occupancy by color.
	NumPieces  [ColorArraySize][FigureArraySize]int8 // number of (color, figure) on the board. NoColor/NoFigure means all.
	SideToMove Color                                 // which side is to move. SideToMove is updated by DoMove and UndoMove.

	FullMoveNumber int
	HalfMoveClock  int
	Ply            int // current Ply

	states []state // a state for each Ply
	curr   *state  // current state
}

// NewPosition returns a new position.
func NewPosition() *Position {
	pos := &Position{states: make([]state, 1)}
	pos.curr = &pos.states[pos.Ply]
	return pos
}

func PositionFromFEN(fen string) (*Position, error) {
	// Split fen into 6 fields.
	// Same as string.Fields() but creates much less garbage.
	f, p := [6]string{}, 0
	for i := 0; i < len(fen); {
		// Find the start and end of the token.
		for ; i < len(fen) && fen[i] == ' '; i++ {
		}
		start := i
		for ; i < len(fen) && fen[i] != ' '; i++ {
		}
		limit := i

		if start == limit {
			continue
		}
		if p >= len(f) {
			return nil, fmt.Errorf("fen has too many fields")
		}
		f[p] = fen[start:limit]
		p++
	}
	if p < len(f) {
		return nil, fmt.Errorf("fen has too few fields")
	}

	// Parse each field.
	pos := NewPosition()
	if err := ParsePiecePlacement(f[0], pos); err != nil {
		return nil, err
	}
	if err := ParseSideToMove(f[1], pos); err != nil {
		return nil, err
	}
	if err := ParseCastlingAbility(f[2], pos); err != nil {
		return nil, err
	}
	if err := ParseEnpassantSquare(f[3], pos); err != nil {
		return nil, err
	}
	var err error
	if pos.HalfMoveClock, err = strconv.Atoi(f[4]); err != nil {
		return nil, err
	}
	if pos.FullMoveNumber, err = strconv.Atoi(f[5]); err != nil {
		return nil, err
	}
	return pos, nil
}

// String returns position in FEN format.
// For table format use PrettyPrint.
func (pos *Position) String() string {
	s := FormatPiecePlacement(pos)
	s += " " + FormatSideToMove(pos)
	s += " " + FormatCastlingAbility(pos)
	s += " " + FormatEnpassantSquare(pos)
	s += " " + strconv.Itoa(pos.HalfMoveClock)
	s += " " + strconv.Itoa(pos.FullMoveNumber)
	return s
}

// prev returns state at previous Ply.
func (pos *Position) prev() *state {
	return &pos.states[pos.Ply-1]
}

// popState pops one Ply.
func (pos *Position) popState() {
	pos.states = pos.states[:pos.Ply]
	pos.Ply--
	pos.curr = &pos.states[pos.Ply]
}

// pushState adds one Ply.
func (pos *Position) pushState() {
	pos.states = append(pos.states, pos.states[pos.Ply])
	pos.Ply++
	pos.curr = &pos.states[pos.Ply]
}

// IsEnpassantSquare returns truee if sq is the enpassant square
func (pos *Position) IsEnpassantSquare(sq Square) bool {
	return sq != SquareA1 && sq == pos.EnpassantSquare()
}

// EnpassantSquare returns the enpassant square.
func (pos *Position) EnpassantSquare() Square {
	return pos.curr.EnpassantSquare
}

// CastlingAbility returns kings' castling ability.
func (pos *Position) CastlingAbility() Castle {
	return pos.curr.Castle
}

// Zobrist returns the zobrist key of the position.
func (pos *Position) Zobrist() uint64 {
	return pos.curr.Zobrist
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
	pos.curr.Zobrist ^= zobristCastle[pos.curr.Castle]
	pos.curr.Castle = castle
	pos.curr.Zobrist ^= zobristCastle[pos.curr.Castle]
}

// SetSideToMove sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetSideToMove(col Color) {
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
	pos.SideToMove = col
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
}

// SetEnpassantSquare sets the enpassant square correctly updating the Zobrist key.
func (pos *Position) SetEnpassantSquare(sq Square) {
	pos.curr.Zobrist ^= zobristEnpassant[pos.EnpassantSquare()]
	pos.curr.EnpassantSquare = sq
	pos.curr.Zobrist ^= zobristEnpassant[pos.EnpassantSquare()]
}

// ByPiece is a shortcut for ByColor[col]&ByFigure[fig].
func (pos *Position) ByPiece(col Color, fig Figure) Bitboard {
	return pos.ByColor[col] & pos.ByFigure[fig]
}

// Put puts a piece on the board.
// Does nothing if pi is NoPiece. Does not validate input.
func (pos *Position) Put(sq Square, pi Piece) {
	if pi != NoPiece {
		pos.curr.Zobrist ^= zobristPiece[pi][sq]
		col, fig := pi.Color(), pi.Figure()
		bb := sq.Bitboard()

		pos.ByColor[col] |= bb
		pos.ByFigure[fig] |= bb
		pos.NumPieces[NoColor][NoFigure]++
		pos.NumPieces[NoColor][fig]++
		pos.NumPieces[col][NoFigure]++
		pos.NumPieces[col][fig]++
	}
}

// Remove removes a piece from the table.
// Does nothing if pi is NoPiece. Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	if pi != NoPiece {
		pos.curr.Zobrist ^= zobristPiece[pi][sq]
		col, fig := pi.Color(), pi.Figure()
		bb := ^sq.Bitboard()

		pos.ByColor[pi.Color()] &= bb
		pos.ByFigure[pi.Figure()] &= bb
		pos.NumPieces[NoColor][NoFigure]--
		pos.NumPieces[NoColor][fig]--
		pos.NumPieces[col][NoFigure]--
		pos.NumPieces[col][fig]--
	}
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.ByColor[White]|pos.ByColor[Black])>>sq&1 == 0
}

// Get returns the piece at sq.
func (pos *Position) Get(sq Square) Piece {
	col := White*Color(pos.ByColor[White]>>sq&1) +
		Black*Color(pos.ByColor[Black]>>sq&1)
	if col == NoColor {
		return NoPiece
	}
	for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
		if pos.ByFigure[fig]&sq.Bitboard() != 0 {
			return ColorFigure(col, fig)
		}
	}
	panic("unreachable")
}

// IsThreeFoldRepetition returns whether current position was seen three times already.
func (pos *Position) IsThreeFoldRepetition() bool {
	if pos.Ply-pos.curr.IrreversiblePly < 4 {
		return false
	}

	c, z := 0, pos.Zobrist()
	for i := pos.Ply; i >= pos.curr.IrreversiblePly; i -= 2 {
		if pos.states[i].Zobrist == z {
			if c++; c == 3 {
				return true
			}
		}
	}
	return false
}

// IsChecked returns true if side's king is checked.
func (pos *Position) IsChecked(side Color) bool {
	kingSq := pos.ByPiece(side, King).AsSquare()
	return pos.IsAttackedBy(kingSq, side.Opposite())
}

// PrettyPrint pretty prints the current position to log.
func (pos *Position) PrettyPrint() {
	for r := 7; r >= 0; r-- {
		line := ""
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			if pos.IsEnpassantSquare(sq) {
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

// DoMove performs a move of known piece.
// Expects the move to be valid.
func (pos *Position) DoMove(move Move) {
	pos.pushState()

	// Update castling rights.
	pos.SetCastlingAbility(pos.curr.Castle &^ lostCastleRights[move.From] &^ lostCastleRights[move.To])

	// Update IrreversiblePly.
	if move.Capture() != NoPiece || move.Piece().Figure() == Pawn {
		pos.curr.IrreversiblePly = pos.Ply
	}

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
	pos.Remove(move.CaptureSquare(), move.Capture())
	pos.Put(move.To, move.Target())
	pos.SetSideToMove(pos.SideToMove.Opposite())
}

// UndoMove takes back a move.
// Expects the move to be valid.
func (pos *Position) UndoMove(move Move) {
	pos.SetCastlingAbility(pos.prev().Castle)
	pos.SetEnpassantSquare(pos.prev().EnpassantSquare)
	pos.SetSideToMove(pos.SideToMove.Opposite())

	// Modify the chess board.
	pi := move.Piece()
	pos.Put(move.From, pi)
	pos.Remove(move.To, move.Target())
	pos.Put(move.CaptureSquare(), move.Capture())

	// Move rook on castling.
	if move.MoveType == Castling {
		rook, start, end := CastlingRook(move.To)
		pos.Put(start, rook)
		pos.Remove(end, rook)
	}

	pos.popState()
}

func (pos *Position) genPawnPromotions(from, to Square, capt Piece, violent bool, moves *[]Move) {
	rank := to.Rank()
	if violent && rank != 0 && rank != 7 && capt == NoPiece {
		return
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
	if moveType == Normal && pos.IsEnpassantSquare(to) {
		moveType = Enpassant
	}
	for p := pMin; p <= pMax; p++ {
		*moves = append(*moves, MakeMove(moveType, from, to, capt, ColorFigure(pos.SideToMove, p)))
	}
}

// genPawnAdvanceMoves moves pawns one square.
func (pos *Position) genPawnAdvanceMoves(violent bool, moves *[]Move) {
	var forward Square
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
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
		pos.genPawnPromotions(from, to, NoPiece, violent, moves)
	}
}

// genPawnDoubleAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(moves *[]Move) {
	var forward Square
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	free := ^(pos.ByColor[White] | pos.ByColor[Black])
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
		pos.genPawnPromotions(from, to, NoPiece, false, moves)
	}
}

func (pos *Position) pawnCapture(to Square) Piece {
	if pos.IsEnpassantSquare(to) {
		return ColorFigure(pos.SideToMove.Opposite(), Pawn)
	}
	return pos.Get(to)
}

func (pos *Position) genPawnAttackMoves(violent bool, moves *[]Move) {
	enemy := pos.ByColor[pos.SideToMove.Opposite()]
	if pos.EnpassantSquare() != SquareA1 {
		enemy |= pos.EnpassantSquare().Bitboard()
	}

	forward := 0
	if pos.SideToMove == White {
		enemy >>= 8
		forward = +1
	} else {
		enemy <<= 8
		forward = -1
	}

	// Left
	bb := pos.ByPiece(pos.SideToMove, Pawn)
	att := RankFile(forward, -1)
	for bbl := bb & ((enemy & ^FileBb(7)) << 1); bbl > 0; {
		from := bbl.Pop()
		to := from + att
		capt := pos.pawnCapture(to)
		pos.genPawnPromotions(from, to, capt, violent, moves)
	}

	// Right
	att = RankFile(forward, +1)
	for bbr := bb & ((enemy & ^FileBb(0)) >> 1); bbr > 0; {
		from := bbr.Pop()
		to := from + att
		capt := pos.pawnCapture(to)
		pos.genPawnPromotions(from, to, capt, violent, moves)
	}
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, moves *[]Move) {
	for att != 0 {
		to := att.Pop()
		*moves = append(*moves, MakeMove(Normal, from, to, pos.Get(to), pi))
	}
}

func (pos *Position) violentMask(violent bool) Bitboard {
	if violent {
		// Capture enemy pieces.
		return pos.ByColor[pos.SideToMove.Opposite()]
	}
	// Generate all moves, except capturing own pieces.
	return ^pos.ByColor[pos.SideToMove]
}

func (pos *Position) genKnightMoves(violent bool, moves *[]Move) {
	mask := pos.violentMask(violent)
	pi := ColorFigure(pos.SideToMove, Knight)
	for bb := pos.ByPiece(pos.SideToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := BbKnightAttack[from] & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genBishopMoves(fig Figure, violent bool, moves *[]Move) {
	mask := pos.violentMask(violent)
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := BishopMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genRookMoves(fig Figure, violent bool, moves *[]Move) {
	mask := pos.violentMask(violent)
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := RookMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genKingMovesNear(violent bool, moves *[]Move) {
	mask := pos.violentMask(violent)
	pi := ColorFigure(pos.SideToMove, King)
	from := pos.ByPiece(pos.SideToMove, King).AsSquare()
	att := BbKingAttack[from] & mask
	pos.genBitboardMoves(pi, from, att, moves)
}

func (pos *Position) genKingCastles(moves *[]Move) {
	rank := pos.SideToMove.KingHomeRank()
	oo, ooo := WhiteOO, WhiteOOO
	if pos.SideToMove == Black {
		oo, ooo = BlackOO, BlackOOO
	}

	// Castle king side.
	if pos.curr.Castle&oo != 0 {
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

		*moves = append(*moves, MakeMove(Castling, r4, r6, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOO:

	// Castle queen side.
	if pos.curr.Castle&ooo != 0 {
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

		*moves = append(*moves, MakeMove(Castling, r4, r2, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOOO:
}

// IsAttackedBy returns true if sq is under attacked by side.
func (pos *Position) IsAttackedBy(sq Square, side Color) bool {
	enemy := pos.ByColor[side]
	if BbPawnAttack[sq]&enemy&pos.ByFigure[Pawn] != 0 {
		pawns := pos.ByPiece(side, Pawn)
		if side == White {
			pawns <<= 8
		} else {
			pawns >>= 8
		}
		left := pawns & ^FileBb(7) << 1
		right := pawns & ^FileBb(0) >> 1

		if att := sq.Bitboard() & (left | right); att != 0 {
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
func (pos *Position) GenerateMoves(moves *[]Move) {
	pos.genPawnDoubleAdvanceMoves(moves)
	pos.genBishopMoves(Queen, false, moves)
	pos.genRookMoves(Rook, false, moves)
	pos.genPawnAttackMoves(false, moves)
	pos.genKnightMoves(false, moves)
	pos.genKingMovesNear(false, moves)
	pos.genPawnAdvanceMoves(false, moves)
	pos.genKingCastles(moves)
	pos.genBishopMoves(Bishop, false, moves)
	pos.genRookMoves(Queen, false, moves)
}

// GenerateViolentMoves append to moves all violent moves valid from pos.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
func (pos *Position) GenerateViolentMoves(moves *[]Move) {
	pos.genBishopMoves(Bishop, true, moves)
	pos.genBishopMoves(Queen, true, moves)
	pos.genKingMovesNear(true, moves)
	pos.genKnightMoves(true, moves)
	pos.genPawnAdvanceMoves(true, moves)
	pos.genPawnAttackMoves(true, moves)
	pos.genRookMoves(Queen, true, moves)
	pos.genRookMoves(Rook, true, moves)
}

// GenerateFigureMoves generate moves for a given figure.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
func (pos *Position) GenerateFigureMoves(fig Figure, moves *[]Move) {
	switch fig {
	case Pawn:
		pos.genPawnAttackMoves(false, moves)
		pos.genPawnAdvanceMoves(false, moves)
		pos.genPawnDoubleAdvanceMoves(moves)
	case Knight:
		pos.genKnightMoves(false, moves)
	case Bishop:
		pos.genBishopMoves(Bishop, false, moves)
	case Rook:
		pos.genRookMoves(Rook, false, moves)
	case Queen:
		pos.genBishopMoves(Queen, false, moves)
		pos.genRookMoves(Queen, false, moves)
	case King:
		pos.genKingMovesNear(false, moves)
		pos.genKingCastles(moves)
	}
}
