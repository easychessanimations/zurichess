package engine

import (
	"errors"
	"fmt"
	"log"
	"time"
	"unsafe"
)

var _ = log.Println
var _ = fmt.Println

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")
)

type EngineOptions struct {
	HashSizeMB  uint64 // Hash size in mega bytes
	AnalyseMode bool   // True to display info strings.
}

var DefaultEngineOptions = EngineOptions{
	HashSizeMB:  32,
	AnalyseMode: false,
}

// SetDefaults sets the default values for opt.
// TODO: If this is getting large, consider using "reflect"
func (opt *EngineOptions) SetDefaults() {
	if opt.HashSizeMB == 0 {
		opt.HashSizeMB = DefaultEngineOptions.HashSizeMB
	}
	if opt.AnalyseMode == false {
		opt.AnalyseMode = DefaultEngineOptions.AnalyseMode
	}
}

type hashKind uint8

const (
	NoKind hashKind = iota
	Exact
)

type HashEntry struct {
	Lock  uint64   // normally position's zobrist key
	Kind  hashKind // type of hash
	Depth int      // searched depth
	Score int      // score
	Move  Move     // best move found. TODO: remove me
}

// HashTable is a transposition table.
// Engine uses such a table to cache moves so it doesn't recompute them again.
type HashTable struct {
	table     []HashEntry
	mask      uint64 // mask is used to determine the index in the table.
	hit, miss uint64
}

// NewHashTable builds transposition table that takes up to hashSizeMB megabytes.
func NewHashTable(hashSizeMB uint64) HashTable {
	// Choose hashSize such that it is a power of two.
	hashEntrySize := uint64(unsafe.Sizeof(HashEntry{}))
	hashSize := (hashSizeMB << 20) / hashEntrySize
	for hashSize&(hashSize-1) != 0 {
		hashSize &= hashSize - 1
	}

	log.Printf("hashEntrySize %d * hashSize %d = %d MB <= HashSizeMB %d",
		hashEntrySize, hashSize, hashEntrySize*hashSize>>20, hashSizeMB)

	return HashTable{
		table: make([]HashEntry, hashSize),
		mask:  hashSize - 1,
	}
}

// Put puts a new entry in the database.
// Current strategy is to always replace.
func (ht *HashTable) Put(entry HashEntry) {
	key := entry.Lock & ht.mask
	if ht.table[key].Kind == NoKind || entry.Depth <= ht.table[key].Depth+1 {
		ht.table[key] = entry
	}
}

// Get returns an entry from the database.
// Lock of the returned entry matches lock.
func (ht *HashTable) Get(lock uint64) (HashEntry, bool) {
	key := lock & ht.mask
	if ht.table[key].Kind != NoKind && ht.table[key].Lock == lock {
		ht.hit++
		return ht.table[key], true
	} else {
		ht.miss++
		return HashEntry{}, false
	}
}

type Engine struct {
	Options  EngineOptions
	Position *Position // current Position

	moves []Move    // moves stack
	nodes uint64    // number of nodes evaluated
	hash  HashTable // transposition table

	pieces        [ColorMaxValue][FigureMaxValue]int
	pieceScore    [2]int // score for pieces for mid and end game.
	positionScore [2]int // score for position for mid and end game.
	maxPly        int    // max ply currently searching at.
}

// Init initializes the engine.
func NewEngine(pos *Position, opt EngineOptions) *Engine {
	opt.SetDefaults()

	eng := &Engine{
		Options:  opt,
		Position: pos,
		moves:    make([]Move, 0, 1024),
		hash:     NewHashTable(opt.HashSizeMB),
	}
	eng.countMaterial()
	return eng
}

// ParseMove parses the move from a string.
func (eng *Engine) ParseMove(move string) Move {
	return eng.Position.ParseMove(move)
}

// put adjusts score after puting piece on sq.
// mask is which side is to move.
// delta is -1 if the piece is taken (including undo), 1 otherwise.
func (eng *Engine) put(sq Square, piece Piece, delta int) {
	col := piece.Color()
	fig := piece.Figure()
	weight := ColorWeight[col]
	mask := ColorMask[col]

	eng.pieces[NoColor][NoFigure] += delta
	eng.pieces[col][NoFigure] += delta
	eng.pieces[NoColor][fig] += delta
	eng.pieces[col][fig] += delta

	eng.pieceScore[MidGame] += delta * weight * FigureBonus[fig][MidGame]
	eng.pieceScore[EndGame] += delta * weight * FigureBonus[fig][EndGame]
	eng.positionScore[MidGame] += delta * weight * PieceSquareTable[fig][mask^sq][MidGame]
	eng.positionScore[EndGame] += delta * weight * PieceSquareTable[fig][mask^sq][EndGame]
}

// adjust updates score after making a move.
// delta is -1 if the move is taken back, 1 otherwise.
// Position.ToMove must have not been updated already.
// TODO: enpassant.
func (eng *Engine) adjust(move Move, delta int) {
	color := eng.Position.ToMove

	if move.MoveType == Promotion {
		eng.put(move.From, ColorFigure(color, Pawn), -delta)
	} else {
		eng.put(move.From, move.Target, -delta)
	}
	eng.put(move.To, move.Target, delta)

	if move.MoveType == Castling {
		rookStart := RookStartSquare(move.To)
		rookEnd := RookEndSquare(move.To)
		rook := CastlingRook(move.To)
		eng.put(rookStart, rook, -delta)
		eng.put(rookEnd, rook, delta)
	}

	if move.Capture != NoPiece {
		eng.put(move.To, move.Capture, -delta)
	}
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.adjust(move, 1)
	eng.Position.DoMove(move)
}

// UndoMove undoes a move. Must be the last move.
func (eng *Engine) UndoMove(move Move) {
	eng.Position.UndoMove(move)
	eng.adjust(move, -1)
}

// countMaterial counts pieces and updates the eng.pieceMgScore
func (eng *Engine) countMaterial() {
	eng.pieceScore[MidGame] = 0
	eng.pieceScore[EndGame] = 0
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			bb := eng.Position.ByPiece(col, fig)
			for bb > 0 {
				eng.put(bb.Pop(), ColorFigure(col, fig), 1)
			}
		}
	}
}

// phase returns current phase and total phase.
// phase is determined by the number of pieces left in the game where
// pawn has score 0, knight and bishop 1, rook 2, queen 2.
// See Tapered Eval:
// https://chessprogramming.wikispaces.com/Tapered+Eval
func (eng *Engine) phase() (int, int) {
	totalPhase := 16*0 + 4*1 + 4*1 + 4*2 + 2*4
	currPhase := totalPhase
	currPhase -= eng.pieces[NoColor][Pawn] * 0
	currPhase -= eng.pieces[NoColor][Knight] * 1
	currPhase -= eng.pieces[NoColor][Bishop] * 1
	currPhase -= eng.pieces[NoColor][Rook] * 2
	currPhase -= eng.pieces[NoColor][Queen] * 4
	currPhase = (currPhase*256 + totalPhase/2) / totalPhase
	return currPhase, 256
}

// Evaluate current Position from white's POV.
// Figure values and bonuses are taken from:
// http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) Score() int {
	eng.nodes++

	// Piece score is something between MidGame and EndGame
	// depending on the pieces on the table.
	scoreMg := eng.pieceScore[MidGame] + eng.positionScore[MidGame]
	scoreEg := eng.pieceScore[EndGame] + eng.positionScore[EndGame]
	currPhase, totalPhase := eng.phase()
	score := (scoreMg*(totalPhase-currPhase) + scoreEg*currPhase) / totalPhase

	// Give bonus for connected bishops.
	if eng.pieces[White][Bishop] >= 2 {
		score += BishopPairBonus
	}
	if eng.pieces[Black][Bishop] >= 2 {
		score -= BishopPairBonus
	}

	// Give bonuses based on number of pawns.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		numPawns := eng.pieces[col][Pawn]
		if numPawns > 5 {
			adjust := KnightPawnBonus * eng.pieces[col][Knight]
			adjust -= RookPawnPenalty * eng.pieces[col][Rook]
			score += adjust * ColorWeight[col] * (numPawns - 5)
		}
	}

	return score
}

// EndPosition determines whether this is and end game
// Position based on the number of pieces.
// Returns score and a bool if the game has ended.
func (eng *Engine) EndPosition() (int, bool) {
	if eng.pieces[White][King] == 0 {
		return -MateScore, true
	}
	if eng.pieces[Black][King] == 0 {
		return +MateScore, true
	}

	// K vs K is draw.
	if eng.pieces[White][NoPiece]+eng.pieces[Black][NoPiece] == 2 {
		return 0, true
	}

	// KN vs K and KB vs K are draws
	if eng.pieces[White][NoPiece]+eng.pieces[Black][NoPiece] == 3 {
		if eng.pieces[White][Knight]+eng.pieces[White][Bishop]+
			eng.pieces[Black][Knight]+eng.pieces[Black][Bishop] == 1 {
			return 0, true
		}
	}

	return 0, false
}

// popMove pops last move.
func (eng *Engine) popMove() Move {
	last := len(eng.moves) - 1
	move := eng.moves[last]
	eng.moves = eng.moves[:last]
	return move
}

// quiesce searches a quite move.
func (eng *Engine) quiesce(alpha, beta int, ply int) int {
	color := eng.Position.ToMove
	score := ColorWeight[color] * eng.Score()
	if score >= beta {
		return beta
	}
	if ply == eng.maxPly {
		return score
	}
	if score > alpha {
		alpha = score
	}

	start := len(eng.moves)
	moveGen := NewMoveGenerator(eng.Position, true)
	for piece := WhitePawn; piece != NoPiece; {
		if len(eng.moves) == start {
			piece, eng.moves = moveGen.Next(eng.moves)
			continue
		}

		move := eng.popMove()
		if move.Capture == NoPiece {
			// For quiscence search only captures are considered.
			continue
		}

		eng.DoMove(move)
		if !eng.Position.IsChecked(color) {
			score := -eng.quiesce(-beta, -alpha, ply+1)
			if score >= beta {
				eng.UndoMove(move)
				eng.moves = eng.moves[:start]
				return beta
			}
			if score > alpha {
				alpha = score
			}
		}
		eng.UndoMove(move)
	}

	return alpha
}

// negamax implements negamax framework with fail-soft.
// http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
// alpha, beta represent lower and upper bounds.
// ply is the move number (increasing).
func (eng *Engine) negamax(alpha, beta int, ply int) (Move, int) {
	// Check the transposition table.
	if entry, ok := eng.hash.Get(eng.Position.Zobrist); ok {
		if eng.maxPly-ply <= entry.Depth {
			return entry.Move, entry.Score
		}
	}

	color := eng.Position.ToMove
	if score, done := eng.EndPosition(); done {
		return Move{}, ColorWeight[color] * score
	}
	if ply == eng.maxPly {
		return Move{}, eng.quiesce(alpha, beta, 0)
	}

	localAlpha := alpha
	bestMove, bestScore := Move{}, -InfinityScore
	moveGen := NewMoveGenerator(eng.Position, false)
	for start, piece := len(eng.moves), WhitePawn; piece != NoPiece; {
		if len(eng.moves) == start {
			piece, eng.moves = moveGen.Next(eng.moves)
			continue
		}

		move := eng.popMove()
		eng.DoMove(move)
		if !eng.Position.IsChecked(color) {
			_, score := eng.negamax(-beta, -localAlpha, ply+1)
			score = -score
			if score > KnownWinScore {
				score--
			}
			if score >= beta {
				eng.UndoMove(move)
				eng.moves = eng.moves[:start]
				return Move{}, beta
			}
			if score > bestScore {
				bestMove, bestScore = move, score
				if score > localAlpha {
					localAlpha = score
				}
			}
		}
		eng.UndoMove(move)
	}

	if bestMove.MoveType == NoMove {
		if eng.Position.IsChecked(color) {
			bestMove, bestScore = Move{}, -MateScore
		} else {
			bestMove, bestScore = Move{}, 0
		}
	}

	if alpha <= bestScore && bestScore < beta {
		// Update the transposition table with the new entry.
		eng.hash.Put(HashEntry{
			Lock:  eng.Position.Zobrist,
			Kind:  Exact,
			Depth: eng.maxPly - ply,
			Score: bestScore,
			Move:  bestMove,
		})
	}

	return bestMove, bestScore
}

func (eng *Engine) alphaBeta() (Move, int) {
	move, score := eng.negamax(-InfinityScore, +InfinityScore, 0)
	score *= ColorWeight[eng.Position.ToMove]
	return move, score
}

// Play find the next move.
// tc should already be started.
func (eng *Engine) Play(tc TimeControl) (Move, error) {
	var move Move
	var score int

	eng.nodes = 0

	start := time.Now()
	for depth := tc.NextDepth(); depth != 0; depth = tc.NextDepth() {
		eng.maxPly = depth
		move, score = eng.alphaBeta()
		elapsed := time.Now().Sub(start)
		_, _ = score, elapsed
		if eng.Options.AnalyseMode {
			fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d pv %v\n",
				depth, score, eng.nodes, elapsed/time.Millisecond,
				eng.nodes*uint64(time.Second)/uint64(elapsed+1),
				move)
		}
	}

	if eng.Options.AnalyseMode {
		log.Printf("hash: hit = %d, miss = %d, ratio %.2f%%",
			eng.hash.hit, eng.hash.miss, float32(eng.hash.hit)/float32(eng.hash.hit+eng.hash.miss)*100)
	}

	if move.MoveType == NoMove {
		// If there is no valid move, then it's a stalement or a checkmate.
		if eng.Position.IsChecked(eng.Position.ToMove) {
			return Move{}, ErrorCheckMate
		} else {
			return Move{}, ErrorStaleMate
		}
	}

	return move, nil
}
