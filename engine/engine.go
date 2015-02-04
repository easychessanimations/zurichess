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
	return MvvLva(m.Piece().Figure(), m.Capture.Figure()) +
		MvvLva(NoFigure, m.Promotion().Figure())
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

// EngineOptions keeps engine's optins.
type EngineOptions struct {
	AnalyseMode bool // true to display info strings
}

// EngineStats stores some basic stats on the engine.
// Statistics are per each depth search.
type EngineStats struct {
	CacheHit  uint64 // number of times the position was found transposition table
	CacheMiss uint64 // number of times the position was not found in the transposition table
	Nodes     uint64 // number of nodes searched
}

// CacheHitRatio returns the ration of hits over total number of lookups.
func (es *EngineStats) CacheHitRatio() float32 {
	return float32(es.CacheHit) / float32(es.CacheHit+es.CacheMiss)
}

// Engine implements the logic to search the best move for a position.
type Engine struct {
	Options  EngineOptions
	Position *Position // current Position
	Stats    EngineStats

	moves []Move // moves stack
	root  Move   // move at root, set only if score is exact.

	pieces       [ColorArraySize][FigureArraySize]int
	scoreMidGame int
	scoreEndGame int
	maxPly       int16 // max ply currently searching at.
}

// NewEngine creates a new engine.
// If pos is nil, the start position is used.
func NewEngine(pos *Position, opt EngineOptions) *Engine {
	eng := &Engine{
		Options: opt,
		moves:   make([]Move, 0, 1024),
	}
	eng.SetPosition(pos)
	return eng
}

// SetPosition sets current position.
// If pos is nil, the starting position is set.
func (eng *Engine) SetPosition(pos *Position) {
	if pos != nil {
		eng.Position = pos
	} else {
		eng.Position, _ = PositionFromFEN(FENStartPos)
	}
	eng.countMaterial()
}

// put adjusts score after putting piece on sq.
// delta is -1 if the piece is taken (including undo), 1 otherwise.
func (eng *Engine) put(col Color, fig Figure, delta int) {
	eng.pieces[NoColor][NoFigure] += delta
	eng.pieces[col][NoFigure] += delta
	eng.pieces[NoColor][fig] += delta
	eng.pieces[col][fig] += delta
}

// adjust updates score after making a move.
// delta is -1 if the move is taken back, 1 otherwise.
// Position.ToMove must have not been updated already.
// TODO: enpassant.
func (eng *Engine) adjust(move Move, delta int) {
	color := eng.Position.ToMove
	if move.MoveType == Promotion {
		eng.put(color, Pawn, -delta)
		eng.put(move.Target.Color(), move.Target.Figure(), delta)
	}
	if move.Capture != NoPiece {
		eng.put(move.Capture.Color(), move.Capture.Figure(), -delta)
	}
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.scoreMidGame += MidGameMaterial.EvaluateMove(move)
	eng.scoreEndGame += EndGameMaterial.EvaluateMove(move)
	eng.adjust(move, 1)
	eng.Position.DoMove(move)
}

// UndoMove undoes the last move.
func (eng *Engine) UndoMove(move Move) {
	eng.Position.UndoMove(move)
	eng.adjust(move, -1)
	eng.scoreMidGame -= MidGameMaterial.EvaluateMove(move)
	eng.scoreEndGame -= EndGameMaterial.EvaluateMove(move)
}

// countMaterial updates score for current position.
func (eng *Engine) countMaterial() {
	eng.scoreMidGame = MidGameMaterial.EvaluatePosition(eng.Position)
	eng.scoreEndGame = EndGameMaterial.EvaluatePosition(eng.Position)

	for col := NoColor; col <= ColorMaxValue; col++ {
		for fig := NoFigure; fig <= FigureMaxValue; fig++ {
			eng.pieces[col][fig] = 0
		}
	}
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			eng.put(col, fig, eng.Position.ByPiece(col, fig).Popcnt())
		}
	}
}

// phase returns current phase and total phase.
//
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

// Score evaluates current position from white's POV.
//
// For figure values, bonuses, etc see:
//      - https://chessprogramming.wikispaces.com/Simplified+evaluation+function
//      - http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) Score() int16 {
	eng.Stats.Nodes++

	// Piece score is something between MidGame and EndGame
	// depending on the pieces on the table.
	currPhase, totalPhase := eng.phase()
	score := (eng.scoreMidGame*(totalPhase-currPhase) + eng.scoreEndGame*currPhase) / totalPhase

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

// EndPosition determines whether current position is an end game.
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

// popMove pops last move from eng.moves.
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
		Lock:     eng.Position.Zobrist,
		Score:    score,
		Depth:    eng.maxPly - ply,
		Favorite: move,
		Kind:     kind,
	}

	GlobalHashTable.Put(entry)
	if ply == 0 && kind == Exact {
		if entry.Favorite.MoveType == NoMove {
			panic("expected valid move at root")
		}
		eng.root = move
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

func (eng *Engine) tryMove(α, β, ply int16, move Move) int16 {
	color := eng.Position.ToMove
	eng.DoMove(move)
	if eng.Position.IsChecked(color) {
		eng.UndoMove(move)
		return -InfinityScore
	}
	score := -eng.negamax(-β, -α, ply+1)
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
//
// At ply 0 negamax sets eng.root.
func (eng *Engine) negamax(alpha, beta, ply int16) int16 {
	color := eng.Position.ToMove
	if score, done := eng.EndPosition(); done {
		return int16(ColorWeight[color]) * score
	}

	// Check the transposition table.
	entry, has := eng.retrieveHash()
	if has && eng.maxPly-ply <= entry.Depth {
		if ply > 0 && entry.Kind == Exact {
			// Simply return if the score is exact.
			return entry.Score
		}
		if entry.Kind == FailedLow && entry.Score <= alpha {
			// Previously the move failed low so the actual score
			// is at most entry.Score. If that's lower than alpha
			// this will also fail low.
			return entry.Score
		}
		if entry.Kind == FailedHigh && entry.Score >= beta {
			// Previously the move failed high so the actual score
			// is at least entry.Score. If that's higher than beta
			// this will also fail high.
			return entry.Score
		}
	}

	if ply == eng.maxPly {
		// Stop searching when maximum depth is reached.
		score := eng.quiescence(alpha, beta, 0)
		eng.updateHash(alpha, beta, ply, Move{}, score)
		return score
	}

	localAlpha := alpha
	bestMove, bestScore := Move{}, -InfinityScore

	// Try the killer move first.
	// Entry may not have a killer move for cached quiescence moves.
	if has && entry.Favorite.MoveType != NoMove {
		score := eng.tryMove(localAlpha, beta, ply, entry.Favorite)
		if score >= beta { // Fail high.
			eng.updateHash(alpha, beta, ply, entry.Favorite, score)
			return score
		}
		if score > bestScore {
			bestMove, bestScore = entry.Favorite, score
			if score > localAlpha {
				localAlpha = score
			}
		}
	}

	// Try all moves if the killer move failed to produce a cut-off.
	start := len(eng.moves)
	eng.moves = eng.Position.GenerateMoves(eng.moves)
	sort.Sort(sorterByMvvLva(eng.moves[start:]))

	for start < len(eng.moves) {
		move := eng.popMove()
		if has && move == entry.Favorite {
			continue
		}

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

	// If no move was found then the game is over.
	if bestMove.MoveType == NoMove {
		if eng.Position.IsChecked(color) {
			bestScore = -MateScore
		} else {
			bestScore = 0
		}
	}

	eng.updateHash(alpha, beta, ply, bestMove, bestScore)
	return bestScore
}

func inf(a int) int {
	if a <= int(-InfinityScore) {
		return int(-InfinityScore)
	}
	return int(a)
}

func sup(b int) int {
	if b >= int(InfinityScore) {
		return int(InfinityScore)
	}
	return int(b)
}

// alphaBeta starts the search up to depth eng.maxPly.
// The returned score is from current side to move POV.
// estimated is the score from previous depths.
func (eng *Engine) alphaBeta(estimated int16) int16 {
	// This method only implement aspiration windows.
	// (see https://chessprogramming.wikispaces.com/Aspiration+Windows).
	//
	// The gradual widening algorithm is the one used by RobboLito
	// and Stockfish and it is explained here:
	// http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=499768&t=46624
	γ, δ := int(estimated), 10
	α, β := γ-δ, γ+δ

	for {
		score := eng.negamax(int16(α), int16(β), 0)
		if int(score) <= α {
			α = inf(α - δ)
			δ += δ / 2
		} else if int(score) >= β {
			β = sup(β + δ)
			δ += δ / 2
		} else {
			return score
		}
	}
}

// getPrincipalVariation returns the moves.
func (eng *Engine) getPrincipalVariation() []Move {
	seen := make(map[uint64]bool)
	var moves []Move

	next := eng.root
	for next.MoveType != NoMove && !seen[eng.Position.Zobrist] {
		seen[eng.Position.Zobrist] = true
		moves = append(moves, next)
		eng.DoMove(next)
		entry, _ := GlobalHashTable.Get(eng.Position.Zobrist)
		next = entry.Favorite
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
	score := int16(0)

	start := time.Now()
	for maxPly := tc.NextDepth(); maxPly != 0; maxPly = tc.NextDepth() {
		eng.maxPly = int16(maxPly)
		score = eng.alphaBeta(score)
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

	move := eng.root
	if move.MoveType == NoMove {
		// If there is no valid move, then it's a stalemate or a checkmate.
		if eng.Position.IsChecked(eng.Position.ToMove) {
			return Move{}, ErrorCheckMate
		}
		return Move{}, ErrorStaleMate
	}

	return move, nil
}
