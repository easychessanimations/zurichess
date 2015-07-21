package engine

import (
	"fmt"
	"log"
	"strconv"
)

const (
	// No capture, no castling, no promotion.
	Quiet int = 1 << iota
	// Castling and underpromotions (including captures).
	Tactical
	// Captures and queen promotions.
	Violent
	// All moves.
	All = Quiet | Tactical | Violent
)

var (
	// Starting position.
	FENStartPos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

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
	CastlingAbility Castle    // remaining castling rights.
	EnpassantSquare [2]Square // enpassant square (polyglot, fen). If none, then SquareA1.
	IrreversiblePly int       // highest square in which an irreversible move (cannot be part of repetition) was made
	Zobrist         uint64
}

// Position encodes the chess board.
type Position struct {
	ByFigure   [FigureArraySize]Bitboard // bitboards of square occupancy by figure.
	ByColor    [ColorArraySize]Bitboard  // bitboards of square occupancy by color.
	SideToMove Color                     // which side is to move. SideToMove is updated by DoMove and UndoMove.

	HalfMoveClock  int
	FullMoveNumber int
	Ply            int // current Ply

	states []state // a state for each Ply
	curr   *state  // current state
}

// NewPosition returns a new position.
func NewPosition() *Position {
	pos := &Position{
		HalfMoveClock:  0,
		FullMoveNumber: 1,
		states:         make([]state, 1),
	}
	pos.curr = &pos.states[pos.Ply]
	return pos
}

// PositionFromFEN parses fen and returns the position.
//
// fen must contain the position using Forsyth–Edwards Notation
// http://en.wikipedia.org/wiki/Forsyth%E2%80%93Edwards_Notation
func PositionFromFEN(fen string) (*Position, error) {
	// Split fen into 6 fields.
	// Same as string.Fields() but creates much less garbage.
	// The optimization is important when a huge number of positions
	// need to be evaluated.
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
	return pos.curr.EnpassantSquare[1]
}

// CastlingAbility returns kings' castling ability.
func (pos *Position) CastlingAbility() Castle {
	return pos.curr.CastlingAbility
}

// Zobrist returns the zobrist key of the position.
// The returned value is equal to polyglot book key
// (http://hgm.nubati.net/book_format.html).
func (pos *Position) Zobrist() uint64 {
	return pos.curr.Zobrist
}

// Sides return current color to move, and the next color to move.
func (pos *Position) Sides() (Color, Color) {
	return pos.SideToMove, pos.SideToMove.Opposite()
}

// NumNonPawns returns the number of minor and major pieces.
func (pos *Position) NumNonPawns(col Color) int {
	return int((pos.ByColor[col] &^ pos.ByFigure[Pawn] &^ pos.ByFigure[King]).Popcnt())
}

// HasNonPawns returns whether col has at least some minor or major pieces.
func (pos *Position) HasNonPawns(col Color) bool {
	return pos.ByColor[col]&^pos.ByFigure[Pawn]&^pos.ByFigure[King] != 0
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
		sq := bb.Pop()
		if bb != 0 {
			sq2 := bb.Pop()
			return fmt.Errorf("More than one King for %v at %v and %v", col, sq, sq2)
		}
	}

	// Verifies that pieces have the right color.
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for bb := pos.ByColor[col]; bb != 0; {
			sq := bb.Pop()
			pi := pos.Get(sq)
			if pi.Color() != col {
				return fmt.Errorf("Expected color %v, got %v", col, pi)
			}
		}
	}

	// Verifies that no two pieces sit on the same cell.
	for pi1 := PieceMinValue; pi1 <= PieceMaxValue; pi1++ {
		for pi2 := pi1 + 1; pi2 <= PieceMaxValue; pi2++ {
			if pos.ByPiece(pi1.Color(), pi1.Figure())&pos.ByPiece(pi2.Color(), pi2.Figure()) != 0 {
				return fmt.Errorf("%v and %v overlap", pi1, pi2)
			}
		}
	}

	return nil
}

// SetCastlingAbility sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetCastlingAbility(castle Castle) {
	if pos.curr.CastlingAbility == castle {
		return
	}

	pos.curr.Zobrist ^= zobristCastle[pos.curr.CastlingAbility]
	pos.curr.CastlingAbility = castle
	pos.curr.Zobrist ^= zobristCastle[pos.curr.CastlingAbility]
}

// SetSideToMove sets the side to move, correctly updating the Zobrist key.
func (pos *Position) SetSideToMove(col Color) {
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
	pos.SideToMove = col
	pos.curr.Zobrist ^= zobristColor[pos.SideToMove]
}

// SetEnpassantSquare sets the enpassant square correctly updating the Zobrist key.
func (pos *Position) SetEnpassantSquare(sq Square) {
	if sq == pos.curr.EnpassantSquare[1] {
		// In the trivial case both values are SquareA1
		// and zobrist value doesn't change.
		return
	}

	pos.curr.Zobrist ^= zobristEnpassant[pos.curr.EnpassantSquare[0]]
	pos.curr.EnpassantSquare[0] = sq
	pos.curr.EnpassantSquare[1] = sq

	if sq != SquareA1 {
		// In polyglot the hash key for en passant is updated only if the
		// an enpassant capture is possible next move. In other words
		// if there is an enemy pawn next to the end square of the move.
		var theirs Bitboard
		if sq.Rank() == 2 { // White
			theirs, sq = pos.ByPiece(Black, Pawn), RankFile(3, sq.File())
		} else if sq.Rank() == 5 { // Black
			theirs, sq = pos.ByPiece(White, Pawn), RankFile(4, sq.File())
		} else {
			panic("bad en passant square")
		}

		if (sq.File() == 0 || !theirs.Has(sq-1)) && (sq.File() == 7 || !theirs.Has(sq+1)) {
			pos.curr.EnpassantSquare[0] = SquareA1
		}
	}

	pos.curr.Zobrist ^= zobristEnpassant[pos.curr.EnpassantSquare[0]]
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
	}
}

// Remove removes a piece from the table.
// Does nothing if pi is NoPiece. Does not validate input.
func (pos *Position) Remove(sq Square, pi Piece) {
	if pi != NoPiece {
		pos.curr.Zobrist ^= zobristPiece[pi][sq]
		col, fig := pi.Color(), pi.Figure()
		bb := ^sq.Bitboard()

		pos.ByColor[col] &= bb
		pos.ByFigure[fig] &= bb
	}
}

// IsEmpty returns true if there is no piece at sq.
func (pos *Position) IsEmpty(sq Square) bool {
	return (pos.ByColor[White]|pos.ByColor[Black])>>sq&1 == 0
}

// Get returns the piece at sq.
func (pos *Position) Get(sq Square) Piece {
	var col Color
	if pos.ByColor[White].Has(sq) {
		col = White
	} else if pos.ByColor[Black].Has(sq) {
		col = Black
	} else {
		return NoPiece
	}

	for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
		if pos.ByFigure[fig].Has(sq) {
			return ColorFigure(col, fig)
		}
	}
	panic("unreachable")
}

// KnightMobility returns all squares a knigth can reach from sq.
func (pos *Position) KnightMobility(sq Square) Bitboard {
	return bbKnightAttack[sq]
}

// BishopMobility returns the squares a bishop can reach from sq given all pieces.
func (pos *Position) BishopMobility(sq Square, all Bitboard) Bitboard {
	return bishopMagic[sq].Attack(all)
}

// RookMobility returns the squares a rook can reach from sq given all pieces.
func (pos *Position) RookMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all)
}

// QueenMobility returns the squares a queen can reach from sq given all pieces.
func (pos *Position) QueenMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all) | bishopMagic[sq].Attack(all)
}

// KingMobility returns all squares a king can reach from sq.
// Doesn't include castling.
func (pos *Position) KingMobility(sq Square) Bitboard {
	return bbKingAttack[sq]
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
	return pos.GetAttacker(kingSq, side.Opposite()) != NoFigure
}

// PrettyPrint pretty prints the current position to log.
func (pos *Position) PrettyPrint() {
	log.Println("zobrist =", pos.Zobrist())
	log.Println("fen =", pos.String())
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

// DoMove executes a legal move.
func (pos *Position) DoMove(move Move) {
	pos.pushState()

	// Update castling rights.
	pi := move.Piece()
	if pi != NoPiece { // nullmove cannot change castling ability
		pos.SetCastlingAbility(pos.curr.CastlingAbility &^ lostCastleRights[move.From()] &^ lostCastleRights[move.To()])
	}
	// Update IrreversiblePly.
	if move.Capture() != NoPiece || pi.Figure() == Pawn {
		pos.curr.IrreversiblePly = pos.Ply
	}
	// Move rook on castling.
	if move.MoveType() == Castling {
		rook, start, end := CastlingRook(move.To())
		pos.Remove(start, rook)
		pos.Put(end, rook)
	}
	// Set Enpassant square for capturing.
	if pi.Figure() == Pawn &&
		move.From().Bitboard()&BbPawnStartRank != 0 &&
		move.To().Bitboard()&BbPawnDoubleRank != 0 {
		pos.SetEnpassantSquare((move.From() + move.To()) / 2)
	} else {
		pos.SetEnpassantSquare(SquareA1)
	}

	// Update the pieces the chess board.
	pos.Remove(move.From(), pi)
	pos.Remove(move.CaptureSquare(), move.Capture())
	pos.Put(move.To(), move.Target())
	pos.SetSideToMove(pos.SideToMove.Opposite())
}

// UndoMove takes back the last move.
func (pos *Position) UndoMove(move Move) {
	pos.SetCastlingAbility(pos.prev().CastlingAbility)
	pos.SetEnpassantSquare(pos.prev().EnpassantSquare[1])
	pos.SetSideToMove(pos.SideToMove.Opposite())

	// Modify the chess board.
	pi := move.Piece()
	pos.Put(move.From(), pi)
	pos.Remove(move.To(), move.Target())
	pos.Put(move.CaptureSquare(), move.Capture())

	// Move rook on castling.
	if move.MoveType() == Castling {
		rook, start, end := CastlingRook(move.To())
		pos.Put(start, rook)
		pos.Remove(end, rook)
	}

	pos.popState()
}

// PawnThreats returns the set of squares threatened by side's pawns.
func (pos *Position) PawnThreats(side Color) Bitboard {
	pawns := Forward(side, pos.ByPiece(side, Pawn))
	return West(pawns) | East(pawns)
}

func (pos *Position) genPawnPromotions(kind int, moves *[]Move) {
	if kind&(Violent|Tactical) == 0 {
		return
	}

	// Minimum and maximum promotion pieces.
	// Tactical -> Knight - Rook
	// Violent -> Queen
	pMin, pMax := Queen, Rook
	if kind&Violent != 0 {
		pMax = Queen
	}
	if kind&Tactical != 0 {
		pMin = Knight
	}

	us := pos.SideToMove
	them := us.Opposite()

	// Get the pawns that can be promoted.
	all := pos.ByColor[White] | pos.ByColor[Black]
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByColor[them] // their pieces

	forward := Square(0)
	if us == White {
		ours &= BbRank7
		forward = RankFile(+1, 0)
	} else {
		ours &= BbRank2
		forward = RankFile(-1, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward

		if !all.Has(to) { // advance front
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to, NoPiece, ColorFigure(us, p)))
			}
		}
		if to.File() != 0 && theirs.Has(to-1) { // take west
			capt := pos.Get(to - 1)
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to-1, capt, ColorFigure(us, p)))
			}
		}
		if to.File() != 7 && theirs.Has(to+1) { // take east
			capt := pos.Get(to + 1)
			for p := pMin; p <= pMax; p++ {
				*moves = append(*moves, MakeMove(Promotion, from, to+1, capt, ColorFigure(us, p)))
			}
		}
	}
}

// genPawnAdvanceMoves moves pawns one square.
// Does not generate promotions.
func (pos *Position) genPawnAdvanceMoves(kind int, moves *[]Move) {
	if kind&Quiet == 0 {
		return
	}

	ours := pos.ByPiece(pos.SideToMove, Pawn)
	occu := pos.ByColor[White] | pos.ByColor[Black]
	pawn := ColorFigure(pos.SideToMove, Pawn)

	var forward Square
	if pos.SideToMove == White {
		ours = ours &^ South(occu) &^ BbRank7
		forward = RankFile(+1, 0)
	} else {
		ours = ours &^ North(occu) &^ BbRank2
		forward = RankFile(-1, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward
		*moves = append(*moves, MakeMove(Normal, from, to, NoPiece, pawn))
	}
}

// genPawnDoubleAdvanceMoves moves pawns two square.
func (pos *Position) genPawnDoubleAdvanceMoves(kind int, moves *[]Move) {
	if kind&Quiet == 0 {
		return
	}

	ours := pos.ByPiece(pos.SideToMove, Pawn)
	occu := pos.ByColor[White] | pos.ByColor[Black]
	pawn := ColorFigure(pos.SideToMove, Pawn)

	var forward Square
	if pos.SideToMove == White {
		ours &= RankBb(1) &^ South(occu) &^ South(South(occu))
		forward = RankFile(+2, 0)
	} else {
		ours &= RankBb(6) &^ North(occu) &^ North(North(occu))
		forward = RankFile(-2, 0)
	}

	for ours != 0 {
		from := ours.Pop()
		to := from + forward
		*moves = append(*moves, MakeMove(Normal, from, to, NoPiece, pawn))
	}
}

func (pos *Position) pawnCapture(to Square) (MoveType, Piece) {
	if pos.IsEnpassantSquare(to) {
		return Enpassant, ColorFigure(pos.SideToMove.Opposite(), Pawn)
	}
	return Normal, pos.Get(to)
}

// Generate pawn attacks moves.
// Does not generate promotions.
func (pos *Position) genPawnAttackMoves(kind int, moves *[]Move) {
	if kind&Violent == 0 {
		return
	}

	theirs := pos.ByColor[pos.SideToMove.Opposite()]
	if pos.curr.EnpassantSquare[0] != SquareA1 {
		theirs |= pos.curr.EnpassantSquare[0].Bitboard()
	}

	forward := 0
	pawn := ColorFigure(pos.SideToMove, Pawn)
	ours := pos.ByPiece(pos.SideToMove, Pawn)
	if pos.SideToMove == White {
		ours = ours &^ BbRank7
		theirs = South(theirs)
		forward = +1
	} else {
		ours = ours &^ BbRank2
		theirs = North(theirs)
		forward = -1
	}

	// Left
	att := RankFile(forward, -1)
	for bbl := ours & East(theirs); bbl > 0; {
		from := bbl.Pop()
		to := from + att
		mt, capt := pos.pawnCapture(to)
		*moves = append(*moves, MakeMove(mt, from, to, capt, pawn))
	}

	// Right
	att = RankFile(forward, +1)
	for bbr := ours & West(theirs); bbr > 0; {
		from := bbr.Pop()
		to := from + att
		mt, capt := pos.pawnCapture(to)
		*moves = append(*moves, MakeMove(mt, from, to, capt, pawn))
	}
}

func (pos *Position) genBitboardMoves(pi Piece, from Square, att Bitboard, moves *[]Move) {
	for att != 0 {
		to := att.Pop()
		*moves = append(*moves, MakeMove(Normal, from, to, pos.Get(to), pi))
	}
}

func (pos *Position) getMask(kind int) Bitboard {
	mask := Bitboard(0)
	if kind&Violent != 0 {
		// Generate all attacks.
		// Promotions are handled specially.
		mask |= pos.ByColor[pos.SideToMove.Opposite()]
	}
	if kind&Quiet != 0 {
		// Generate all non-attacks.
		mask |= ^(pos.ByColor[White] | pos.ByColor[Black])
	}
	// Tactical is handled specially.
	return mask
}

func (pos *Position) genKnightMoves(kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	pi := ColorFigure(pos.SideToMove, Knight)
	for bb := pos.ByPiece(pos.SideToMove, Knight); bb != 0; {
		from := bb.Pop()
		att := bbKnightAttack[from] & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genBishopMoves(fig Figure, kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := bishopMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genRookMoves(fig Figure, kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	pi := ColorFigure(pos.SideToMove, fig)
	ref := pos.ByColor[White] | pos.ByColor[Black]
	for bb := pos.ByPiece(pos.SideToMove, fig); bb != 0; {
		from := bb.Pop()
		att := rookMagic[from].Attack(ref) & mask
		pos.genBitboardMoves(pi, from, att, moves)
	}
}

func (pos *Position) genKingMovesNear(kind int, moves *[]Move) {
	mask := pos.getMask(kind)
	pi := ColorFigure(pos.SideToMove, King)
	from := pos.ByPiece(pos.SideToMove, King).AsSquare()
	att := bbKingAttack[from] & mask
	pos.genBitboardMoves(pi, from, att, moves)
}

func (pos *Position) genKingCastles(kind int, moves *[]Move) {
	if kind&Tactical == 0 {
		return
	}

	rank := pos.SideToMove.KingHomeRank()
	oo, ooo := WhiteOO, WhiteOOO
	if pos.SideToMove == Black {
		oo, ooo = BlackOO, BlackOOO
	}

	// Castle king side.
	if pos.curr.CastlingAbility&oo != 0 {
		r5 := RankFile(rank, 5)
		r6 := RankFile(rank, 6)
		if !pos.IsEmpty(r5) || !pos.IsEmpty(r6) {
			goto EndCastleOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.GetAttacker(r4, other) != NoFigure ||
			pos.GetAttacker(r5, other) != NoFigure ||
			pos.GetAttacker(r6, other) != NoFigure {
			goto EndCastleOO
		}

		*moves = append(*moves, MakeMove(Castling, r4, r6, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOO:

	// Castle queen side.
	if pos.curr.CastlingAbility&ooo != 0 {
		r3 := RankFile(rank, 3)
		r2 := RankFile(rank, 2)
		r1 := RankFile(rank, 1)
		if !pos.IsEmpty(r3) || !pos.IsEmpty(r2) || !pos.IsEmpty(r1) {
			goto EndCastleOOO
		}

		r4 := RankFile(rank, 4)
		other := pos.SideToMove.Opposite()
		if pos.GetAttacker(r4, other) != NoFigure ||
			pos.GetAttacker(r3, other) != NoFigure ||
			pos.GetAttacker(r2, other) != NoFigure {
			goto EndCastleOOO
		}

		*moves = append(*moves, MakeMove(Castling, r4, r2, NoPiece, ColorFigure(pos.SideToMove, King)))
	}
EndCastleOOO:
}

// GetAttacker returns the smallest figure of color them that attacks sq.
func (pos *Position) GetAttacker(sq Square, them Color) Figure {
	enemy := pos.ByColor[them]
	if enemy&bbPawnAttack[sq]&pos.ByFigure[Pawn] != 0 {
		if att := sq.Bitboard() & pos.PawnThreats(them); att != 0 {
			return Pawn
		}
	}

	// Knight
	if enemy&bbKnightAttack[sq]&pos.ByFigure[Knight] != 0 {
		return Knight
	}
	// Quick test of queen's attack on an empty board.
	if enemy&bbSuperAttack[sq]&^pos.ByFigure[Pawn] == 0 {
		return NoFigure
	}
	// Bishop
	all := pos.ByColor[White] | pos.ByColor[Black]
	bishop := pos.BishopMobility(sq, all)
	if enemy&pos.ByFigure[Bishop]&bishop != 0 {
		return Bishop
	}
	// Rook
	rook := pos.RookMobility(sq, all)
	if enemy&pos.ByFigure[Rook]&rook != 0 {
		return Rook
	}
	// Queen
	if enemy&pos.ByFigure[Queen]&(bishop|rook) != 0 {
		return Queen
	}
	// King.
	if enemy&bbKingAttack[sq]&pos.ByFigure[King] != 0 {
		return King
	}
	return NoFigure
}

// GenerateMoves appends to moves all moves valid from pos.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
// kind is a combination of Quiet, Tactical or Violent.
func (pos *Position) GenerateMoves(kind int, moves *[]Move) {
	// Order of the moves is important because the last quiet
	// moves will be reduced less.  Current order was produced
	// by testing 20 random orders and picking the best.
	pos.genKingMovesNear(kind, moves)
	pos.genPawnDoubleAdvanceMoves(kind, moves)
	pos.genRookMoves(Rook, kind, moves)
	pos.genBishopMoves(Queen, kind, moves)
	pos.genPawnAttackMoves(kind, moves)
	pos.genPawnAdvanceMoves(kind, moves)
	pos.genPawnPromotions(kind, moves)
	pos.genKnightMoves(kind, moves)
	pos.genBishopMoves(Bishop, kind, moves)
	pos.genKingCastles(kind, moves)
	pos.genRookMoves(Queen, kind, moves)
}

// GenerateFigureMoves generate moves for a given figure.
// The generated moves are pseudo-legal, i.e. they can leave the king in check.
// kind is a combination of Quiet, Tactical or Violent.
func (pos *Position) GenerateFigureMoves(fig Figure, kind int, moves *[]Move) {
	switch fig {
	case Pawn:
		pos.genPawnAdvanceMoves(kind, moves)
		pos.genPawnAttackMoves(kind, moves)
		pos.genPawnDoubleAdvanceMoves(kind, moves)
		pos.genPawnPromotions(kind, moves)
	case Knight:
		pos.genKnightMoves(kind, moves)
	case Bishop:
		pos.genBishopMoves(Bishop, kind, moves)
	case Rook:
		pos.genRookMoves(Rook, kind, moves)
	case Queen:
		pos.genBishopMoves(Queen, kind, moves)
		pos.genRookMoves(Queen, kind, moves)
	case King:
		pos.genKingMovesNear(kind, moves)
		pos.genKingCastles(kind, moves)
	}
}
