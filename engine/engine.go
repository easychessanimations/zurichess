// Package engine implements board, moves and search.
package engine

import (
	"fmt"
	"time"
)

const (
	DepthMultiplier        = 8
	CheckDepthExtension    = 6
	NullMoveDepthLimit     = DepthMultiplier
	NullMoveDepthReduction = 1 * DepthMultiplier
)

var (
	// scoreMultiplier is used to compute the score from side
	// to move POV from given the score from white POV.
	scoreMultiplier = [ColorArraySize]int16{0, 1, -1}
)

// Options keeps engine's options.
type Options struct {
	AnalyseMode bool // true to display info strings
}

// Stats stores some basic stats on the engine.
//
// Statistics are reset every iteration of the iterative deepening search.
type Stats struct {
	Start     time.Time // when computation was started
	CacheHit  uint64    // number of times the position was found transposition table
	CacheMiss uint64    // number of times the position was not found in the transposition table
	Nodes     uint64    // number of nodes searched
}

// NPS returns nodes per second.
func (s *Stats) NPS(now time.Time) uint64 {
	elapsed := uint64(now.Sub(s.Start) + 1)
	return s.Nodes * uint64(time.Second) / elapsed
}

// Time returns the number of elapsed milliseconds.
func (s *Stats) Time(now time.Time) uint64 {
	elapsed := uint64(now.Sub(s.Start) + 1)
	return elapsed / uint64(time.Millisecond)
}

// CacheHitRatio returns the ration of hits over total number of lookups.
func (s *Stats) CacheHitRatio() float32 {
	return float32(s.CacheHit) / float32(s.CacheHit+s.CacheMiss)
}

// Engine implements the logic to search the best move for a position.
type Engine struct {
	Options  Options
	Position *Position // current Position
	Stats    Stats

	// killer stores a few killer moves per ply.
	// For killer heuristic see https://chessprogramming.wikispaces.com/Killer+Heuristic
	killer [][2]Move

	maxPly  int16     // max ply currently searching at.
	stack   moveStack // stack of moves
	pvTable pvTable   // principal variation table
}

// NewEngine creates a new engine to search for pos.
// If pos is nil then the start position is used.
func NewEngine(pos *Position, options Options) *Engine {
	eng := &Engine{
		Options: options,
		pvTable: newPvTable(),
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
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.Position.DoMove(move)
}

// UndoMove undoes the last move.
func (eng *Engine) UndoMove(move Move) {
	eng.Position.UndoMove(move)
}

// Score evaluates current position from White's POV.
func (eng *Engine) Score() int16 {
	eng.Stats.Nodes++
	return scoreMultiplier[eng.Position.SideToMove] * Evaluate(eng.Position)
}

// endPosition determines whether current position is an end game.
// Returns score and a bool if the game has ended.
func (eng *Engine) endPosition() (int16, bool) {
	pos := eng.Position // shortcut
	if pos.NumPieces[White][King] == 0 {
		return scoreMultiplier[pos.SideToMove] * -MateScore, true
	}
	if pos.NumPieces[Black][King] == 0 {
		return scoreMultiplier[pos.SideToMove] * +MateScore, true
	}
	// K vs K is draw.
	if pos.NumPieces[NoColor][NoPiece] == 2 {
		return 0, true
	}
	// KN vs K and KB vs K are draws
	if pos.NumPieces[NoColor][NoPiece] == 3 {
		if pos.NumPieces[NoColor][Knight]+pos.NumPieces[NoColor][Bishop] == 1 {
			return 0, true
		}
	}
	// Repetition is a draw.
	if pos.IsThreeFoldRepetition() {
		return 0, true
	}
	return 0, false
}

// retrieveHash gets from GlobalHashTable the current position.
func (eng *Engine) retrieveHash() (HashEntry, bool) {
	entry, ok := GlobalHashTable.Get(eng.Position)
	if ok {
		eng.Stats.CacheHit++
	} else {
		eng.Stats.CacheMiss++
	}
	return entry, ok
}

// updateHash updates GlobalHashTable with current position.
func (eng *Engine) updateHash(α, β, ply, score int16, move *Move) {
	kind := Exact
	if score <= α {
		kind = FailedLow
	} else if score >= β {
		kind = FailedHigh
	}

	GlobalHashTable.Put(eng.Position, HashEntry{
		Score:  score,
		Depth:  eng.maxPly - ply,
		Kind:   kind,
		Target: move.Piece(),
		From:   move.From,
		To:     move.To,
	})
}

// quiescence evaluates the position after solving all captures.
//
// See https://chessprogramming.wikispaces.com/Quiescence+Search.
// This is a very limited search which considers only captures.
// Checks are not considered. In fact it assumes that the move
// ordering will always put the king first.
func (eng *Engine) quiescence(α, β, ply int16) int16 {
	if score, done := eng.endPosition(); done {
		return score
	}
	score := eng.Score()
	if score >= β { // stand pat
		return score
	}
	localα := α
	if score > localα {
		localα = score
	}

	eng.stack.Stack(eng.Position.GenerateViolentMoves, mvvlva)
	var move, bestMove Move
	for eng.stack.PopMove(&move) {
		eng.Position.DoMove(move)
		score := -eng.quiescence(-β, -localα, ply+1)
		eng.Position.UndoMove(move)

		if score >= β {
			eng.stack.PopAll()
			return score
		}
		if score > localα {
			localα = score
			bestMove = move
		}
	}

	if α < localα && localα < β && bestMove.MoveType != NoMove {
		eng.pvTable.Put(eng.Position, bestMove)
	}
	return localα
}

func (eng *Engine) tryMove(α, β, ply, depth int16, move Move) int16 {
	side := eng.Position.SideToMove
	eng.Position.DoMove(move)
	if eng.Position.IsChecked(side) {
		eng.Position.UndoMove(move)
		return -InfinityScore
	}

	score := -eng.negamax(-β, -α, ply+1, depth-DepthMultiplier, move.MoveType != NoMove)
	eng.Position.UndoMove(move)

	// If the position is a known win/loss then the score is
	// increased/decreased slightly so the search takes
	// the shortest/longest path.
	if score > KnownWinScore {
		score--
	}
	if score < KnownLossScore {
		score++
	}
	return score
}

// saveKiller saves a killer move.
func (eng *Engine) saveKiller(ply int16, move Move) {
	for len(eng.killer) <= int(ply) {
		eng.killer = append(eng.killer, [2]Move{})
	}
	if move.Capture() == NoPiece && move != eng.killer[ply][0] { // saves only quiet moves.
		eng.killer[ply][1] = eng.killer[ply][0]
		eng.killer[ply][0] = move
	}
}

// generateMoves generates and orders moves.
func (eng *Engine) generateMoves(ply int16, entry *HashEntry) {
	eng.stack.Stack(
		eng.Position.GenerateMoves,
		func(m Move) int16 {
			// Awards bonus for hash and killer moves.
			if m.From == entry.From && m.To == entry.To && m.Target() == entry.Target {
				return HashMoveBonus
			}
			if len(eng.killer) > int(ply) {
				if m == eng.killer[ply][0] || m == eng.killer[ply][1] {
					return KillerMoveBonus
				}
			}
			return mvvlva(m)
		})
}

// negamax implements negamax framework.
// http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
//
// negamax fails soft, i.e. the score returned can be outside the bounds.
// https://chessprogramming.wikispaces.com/Fail-Soft
//
// α, β represent lower and upper bounds.
// ply is the move number (increasing).
// depth is the fractional depth (decreasing)
// nullMoveAllowed is true if null move is allowed, e.g. to avoid two consecutive null moves.
//
// Returns the score of the current position up to maxPly - ply depth.
// Returned score is from current player's POV.
//
// Invariants:
//   If score <= α then the search failed low and the score is an upper bound.
//   else if score >= β then the search failed high and the score is a lower bound.
//   else score is exact.
//
// Assuming this is a maximizing nodes, failing high means that an ancestors
// minimizing nodes already have a better alternative.
func (eng *Engine) negamax(α, β, ply, depth int16, nullMoveAllowed bool) int16 {
	sideToMove := eng.Position.SideToMove
	if score, done := eng.endPosition(); done {
		return score
	}

	// Check the transposition table.
	entry, has := eng.retrieveHash()
	if has && eng.maxPly-ply <= entry.Depth {
		if ply > 0 && entry.Kind == Exact {
			// Simply return if the score is exact.
			return entry.Score
		}
		if entry.Kind == FailedLow && entry.Score <= α {
			// Previously the move failed low so the actual score
			// is at most entry.Score. If that's lower than α
			// this will also fail low.
			return entry.Score
		}
		if entry.Kind == FailedHigh && entry.Score >= β {
			// Previously the move failed high so the actual score
			// is at least entry.Score. If that's higher than β
			// this will also fail high.
			return entry.Score
		}
	}

	sideIsChecked := eng.Position.IsChecked(sideToMove)
	if sideIsChecked {
		// Extend search when the side to move is in check.
		// https://chessprogramming.wikispaces.com/Check+Extensions
		depth += CheckDepthExtension
	}

	if depth <= 0 {
		// Stop searching when maximum depth is reached.
		score := eng.quiescence(α, β, ply)
		eng.updateHash(α, β, ply, score, &Move{})
		return score
	}

	// Do a null move.
	// https://chessprogramming.wikispaces.com/Null+Move+Pruning
	if pos := eng.Position; nullMoveAllowed && // no two consective null moves
		!sideIsChecked && // not illegal move
		depth > NullMoveDepthLimit && // not very close to leafs
		pos.NumPieces[sideToMove][Pawn]+1 < pos.NumPieces[sideToMove][NoPiece] && // at least one minor/major piece.
		KnownLossScore < α && β < KnownWinScore { // disable in lost or won positions

		reduction := int16(NullMoveDepthLimit)
		if pos.NumPieces[sideToMove][Pawn]+3 < pos.NumPieces[sideToMove][NoPiece] {
			// Reduce more when there are three minor/major pieces.
			reduction += DepthMultiplier
		}
		score := eng.tryMove(β-1, β, ply, depth-reduction, Move{})
		if score >= β {
			return score
		}
	}

	localα := α
	bestMove, bestScore := Move{}, -InfinityScore

	eng.generateMoves(ply, &entry)
	var move Move
	for eng.stack.PopMove(&move) {
		score := eng.tryMove(localα, β, ply, depth, move)
		if score >= β { // Fail high.
			eng.saveKiller(ply, move)
			eng.stack.PopAll()
			eng.updateHash(α, β, ply, score, &move)
			return score
		}
		if score > bestScore {
			bestMove, bestScore = move, score
			if score > localα {
				localα = score
			}
		}
	}

	// If no move was found then the game is over.
	if bestMove.MoveType == NoMove {
		if eng.Position.IsChecked(sideToMove) {
			bestScore = -MateScore
		} else {
			bestScore = 0
		}
	} else {
		eng.saveKiller(ply, bestMove)
	}

	eng.updateHash(α, β, ply, bestScore, &bestMove)
	if α < bestScore && bestScore < β && bestMove.MoveType != NoMove {
		eng.pvTable.Put(eng.Position, bestMove)
	}

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
	γ, δ := int(estimated), 15
	α, β := γ-δ, γ+δ

	for {
		// At root a non-null move is required, cannot prune based on null-move.
		score := eng.negamax(int16(α), int16(β), 0, eng.maxPly*DepthMultiplier, true)
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

// printInfo prints a info UCI string.
func (eng *Engine) printInfo(score int16, pv []Move) {
	now := time.Now()
	fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d ",
		eng.maxPly, score, eng.Stats.Nodes, eng.Stats.Time(now), eng.Stats.NPS(now))

	fmt.Printf("pv")
	for _, m := range pv {
		fmt.Printf(" %v", m.UCI())
	}
	fmt.Printf("\n")
}

// Play evaluates current position.
//
// Returns principal variation, i.e. moves[0] is the best move found.
// If no move was found (e.g. position is already a mate) an empty pv is returned.
//
// Time control, tc, should already be started.
func (eng *Engine) Play(tc TimeControl) (moves []Move) {
	score := int16(0)
	eng.Stats = Stats{Start: time.Now()}
	eng.killer = eng.killer[:0]

	for maxPly := tc.NextDepth(); maxPly >= 0; maxPly = tc.NextDepth() {
		eng.maxPly = int16(maxPly)
		score = eng.alphaBeta(score)
		moves = eng.pvTable.Get(eng.Position)
		if eng.Options.AnalyseMode {
			eng.printInfo(score, moves)
		}
	}

	return moves
}
