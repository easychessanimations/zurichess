package engine

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"
)

var _ = log.Println
var _ = fmt.Println

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")
)

// sorterByMvvLva implements sort.Interface.
// Compares moves by Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
type sorterByMvvLva []Move

func score(m Move) int {
	c := m.Capture.Figure()
	t := m.Target.Figure()
	if m.MoveType != Promotion {
		return MvvLva(t, c)
	} else {
		return MvvLva(Pawn, c) + MvvLva(NoFigure, t)
	}
}

func (c sorterByMvvLva) Len() int {
	return len(c)
}

func (c sorterByMvvLva) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c sorterByMvvLva) Less(i, j int) bool {
	si := score(c[i])
	sj := score(c[j])
	return si < sj
}

// pivotByViolent moves captures and promotions last.
// moves is Engine.moves which is a stack and needs best moves last for early high-cutoffs.
// Returns position of the first violent move.
func pivotByViolent(moves []Move) int {
	i, j := 0, len(moves)-1
	for i < j {
		for ; i < j && !moves[i].IsViolent(); i++ {
		}
		for ; j > i && moves[j].IsViolent(); j-- {
		}
		if i < j {
			moves[i], moves[j] = moves[j], moves[i]
			i++
			j--
		}
	}
	return i
}

// EngineOptions keeps engine's optins.
type EngineOptions struct {
	AnalyseMode bool // True to display info strings.
}

// EngineStats stores some basic stats on the engine.
type EngineStats struct {
	CacheHit  uint64
	CacheMiss uint64
	Nodes     uint64
}

func (es *EngineStats) CacheHitRatio() float32 {
	return float32(es.CacheHit) / float32(es.CacheHit+es.CacheMiss)
}

type Engine struct {
	Options  EngineOptions
	Position *Position // current Position
	Stats    EngineStats

	moves []Move    // moves stack
	root  HashEntry // transposition table

	pieces        [ColorArraySize][FigureArraySize]int
	pieceScore    [2]int // score for pieces for mid and end game.
	positionScore [2]int // score for position for mid and end game.
	maxPly        int16  // max ply currently searching at.
}

// Init initializes the engine.
func NewEngine(pos *Position, opt EngineOptions) *Engine {
	eng := &Engine{
		Options: opt,
		moves:   make([]Move, 0, 1024),
	}
	eng.SetPosition(pos)
	return eng
}

// SetPosition sets current position.
// If pos == nil, the starting position is set.
func (eng *Engine) SetPosition(pos *Position) {
	if pos != nil {
		eng.Position = pos
	} else {
		eng.Position, _ = PositionFromFEN(FENStartPos)
	}
	eng.countMaterial()
}

// UCIToMove parses the move from a string.
func (eng *Engine) UCIToMove(move string) Move {
	return eng.Position.UCIToMove(move)
}

// put adjusts score after puting piece on sq.
// mask is which side is to move.
// delta is -1 if the piece is taken (including undo), 1 otherwise.
func (eng *Engine) put(sq Square, piece Piece, delta int) {
	col := piece.Color()
	fig := piece.Figure()
	mask := ColorMask[col]
	weight := delta * ColorWeight[col]

	eng.pieces[NoColor][NoFigure] += delta
	eng.pieces[col][NoFigure] += delta
	eng.pieces[NoColor][fig] += delta
	eng.pieces[col][fig] += delta

	eng.pieceScore[MidGame] += weight * FigureBonus[MidGame][fig]
	eng.pieceScore[EndGame] += weight * FigureBonus[EndGame][fig]
	eng.positionScore[MidGame] += weight * PieceSquareTable[fig][mask^sq][MidGame]
	eng.positionScore[EndGame] += weight * PieceSquareTable[fig][mask^sq][EndGame]
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
		rook, start, end := CastlingRook(move.To)
		eng.put(start, rook, -delta)
		eng.put(end, rook, delta)
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
	eng.positionScore[MidGame] = 0
	eng.pieceScore[EndGame] = 0
	eng.positionScore[EndGame] = 0
	for col := NoColor; col <= ColorMaxValue; col++ {
		for fig := NoFigure; fig <= FigureMaxValue; fig++ {
			eng.pieces[col][fig] = 0
		}
	}

	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
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
func (eng *Engine) Score() int16 {
	eng.Stats.Nodes++

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
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		numPawns := eng.pieces[col][Pawn]
		adjust := KnightPawnBonus * eng.pieces[col][Knight]
		adjust -= RookPawnPenalty * eng.pieces[col][Rook]
		score += adjust * ColorWeight[col] * (numPawns - 5)
	}

	return int16(score)
}

// EndPosition determines whether this is and end game
// Position based on the number of pieces.
// Returns score and a bool if the game has ended.
func (eng *Engine) EndPosition() (int16, bool) {
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

// retrieveHash gets from GlobalHashTable the current position.
func (eng *Engine) retrieveHash() (HashEntry, bool) {
	entry, ok := GlobalHashTable.Get(eng.Position.Zobrist)
	if ok {
		eng.Stats.CacheHit++
	} else {
		eng.Stats.CacheMiss++
	}
	return entry, ok
}

// updateHash updates GlobalHashTable with current position.
func (eng *Engine) updateHash(alpha, beta, ply int16, move Move, score int16) {
	kind := Exact
	if score <= alpha {
		kind = FailedLow
	} else if score >= beta {
		kind = FailedHigh
	}

	entry := HashEntry{
		Lock:   eng.Position.Zobrist,
		Score:  score,
		Depth:  eng.maxPly - ply,
		Killer: move,
		Kind:   kind,
	}

	GlobalHashTable.Put(entry)
	if ply == 0 {
		eng.root = entry
	}
}

// quiescence searches a quite move.
func (eng *Engine) quiescence(alpha, beta, ply int16) int16 {
	color := eng.Position.ToMove
	score := int16(ColorWeight[color]) * eng.Score()
	if score >= beta {
		return score
	}
	if score > alpha {
		alpha = score
	}

	start := len(eng.moves)
	eng.moves = eng.Position.GenerateViolentMoves(eng.moves)
	sort.Sort(sorterByMvvLva(eng.moves[start:]))
	for start < len(eng.moves) {
		move := eng.popMove()
		eng.DoMove(move)
		if eng.Position.IsChecked(color) {
			eng.UndoMove(move)
			continue
		}
		score := -eng.quiescence(-beta, -alpha, ply+1)
		if score >= beta {
			eng.UndoMove(move)
			eng.moves = eng.moves[:start]
			return score
		}
		if score > alpha {
			alpha = score
		}
		eng.UndoMove(move)
	}
	return alpha
}

func (eng *Engine) tryMove(alpha, beta, ply int16, move Move) int16 {
	color := eng.Position.ToMove
	eng.DoMove(move)
	if eng.Position.IsChecked(color) {
		eng.UndoMove(move)
		return -InfinityScore
	}
	score := -eng.negamax(-beta, -alpha, ply+1)
	if score > KnownWinScore {
		// If the position is a win the score is decreased
		// slightly to the search takes the shortest path.
		score--
	}
	eng.UndoMove(move)
	return score
}

// negamax implements negamax framework.
// http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
//
// negamax fails soft, i.e. the score returned can be outside the bounds.
// https://chessprogramming.wikispaces.com/Fail-Soft
//
// alpha, beta represent lower and upper bounds.
// ply is the move number (increasing).
// Returns the score of the current position up to maxPly - ply depth.
// Returned score is from current player's POV.
//
// Invariants:
//   If score <= alpha then the search failed low
//   else if score >= beta then the search failed high
//   else score is exact.
//
// Assuming this is a maximizing nodes, failing high means that an ancestors
// minimizing nodes already have a better alternative.
func (eng *Engine) negamax(alpha, beta, ply int16) int16 {
	color := eng.Position.ToMove
	if score, done := eng.EndPosition(); done {
		return int16(ColorWeight[color]) * score
	}

	// Check the transposition table.
	entry, has := eng.retrieveHash()
	if has {
		if eng.maxPly-ply > entry.Depth {
			// Wrong depth, so search cannot be pruned.
			goto EndCacheCheck
		}
		if entry.Kind == Exact {
			// Simply return if the score is exact.
			eng.updateHash(alpha, beta, ply, entry.Killer, entry.Score)
			return entry.Score
		}
		if entry.Kind == FailedLow && entry.Score <= alpha {
			// Previously the move failed low so the actual score
			// is at most entry.Score. If that's lower than alpha
			// this will also fail low.
			eng.updateHash(alpha, beta, ply, entry.Killer, entry.Score)
			return entry.Score
		}
		if entry.Kind == FailedHigh && entry.Score >= beta {
			// Previously the move failed high so the actual score
			// is at least entry.Score. If that's higher than beta
			// this will also fail high.
			eng.updateHash(alpha, beta, ply, entry.Killer, entry.Score)
			return entry.Score
		}
	}
EndCacheCheck:

	if ply == eng.maxPly {
		score := eng.quiescence(alpha, beta, 0)
		eng.updateHash(alpha, beta, ply, Move{}, score)
		return score
	}

	localAlpha := alpha
	bestMove, bestScore := Move{}, -InfinityScore

	// Try the killer move first.
	// Entry may not have a killer move for cached quiescence moves.
	if has && entry.Killer.MoveType != NoMove {
		score := eng.tryMove(localAlpha, beta, ply, entry.Killer)
		if score >= beta { // Fail high.
			eng.updateHash(alpha, beta, ply, entry.Killer, score)
			return score
		}
		if score > bestScore {
			bestMove, bestScore = entry.Killer, score
			if score > localAlpha {
				localAlpha = score
			}
		}
	}

	// Try all moves if the killer move failed to produce a cut-off.
	start := len(eng.moves)
	eng.moves = eng.Position.GenerateMoves(eng.moves)
	pivot := pivotByViolent(eng.moves[start:])
	sort.Sort(sorterByMvvLva(eng.moves[start+pivot:]))

	for start < len(eng.moves) {
		move := eng.popMove()
		score := eng.tryMove(localAlpha, beta, ply, move)
		if score >= beta { // Fail high.
			eng.moves = eng.moves[:start]
			eng.updateHash(alpha, beta, ply, move, score)
			return score
		}
		if score > bestScore {
			bestMove, bestScore = move, score
			if score > localAlpha {
				localAlpha = score
			}
		}
	}

	// If no move was found current then the game is over.
	if bestMove.MoveType == NoMove {
		if eng.Position.IsChecked(color) {
			return -MateScore
		} else {
			return 0
		}
	}

	eng.updateHash(alpha, beta, ply, bestMove, bestScore)
	return bestScore
}

// Returned score is from current White's POV.
func (eng *Engine) alphaBeta() int16 {
	score := eng.negamax(-InfinityScore, +InfinityScore, 0)
	score *= int16(ColorWeight[eng.Position.ToMove])
	return score
}

// getPrincipalVariation returns the moves.
func (eng *Engine) getPrincipalVariation() []Move {
	seen := make(map[uint64]bool)
	moves := make([]Move, 0)

	next := eng.root
	for !seen[next.Lock] && next.Kind != NoKind && next.Killer.MoveType != NoMove {
		seen[next.Lock] = true
		moves = append(moves, next.Killer)
		eng.DoMove(next.Killer)
		next, _ = GlobalHashTable.Get(eng.Position.Zobrist)
	}

	// Undo all moves, so we get back to the initial state.
	for i := len(moves) - 1; i >= 0; i-- {
		eng.UndoMove(moves[i])
	}
	return moves
}

// Play finds the next move.
// tc should already be started.
func (eng *Engine) Play(tc TimeControl) (Move, error) {
	eng.Stats = EngineStats{}

	start := time.Now()
	for maxPly := tc.NextDepth(); maxPly != 0; maxPly = tc.NextDepth() {
		eng.maxPly = int16(maxPly)
		score := eng.alphaBeta()
		elapsed := time.Now().Sub(start)

		if eng.Options.AnalyseMode {
			fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d ",
				maxPly, score, eng.Stats.Nodes, elapsed/time.Millisecond,
				eng.Stats.Nodes*uint64(time.Second)/uint64(elapsed+1))

			moves := eng.getPrincipalVariation()
			fmt.Printf("pv")
			for _, move := range moves {
				fmt.Printf(" %v", move)
			}
			fmt.Printf("\n")
		}
	}

	move := eng.root.Killer
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
