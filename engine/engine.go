// Package engine implements board, moves and search.
//
// Search features implemented are:
//
//   * Aspiration window - https://chessprogramming.wikispaces.com/Aspiration+Windows
//   * Check extension - https://chessprogramming.wikispaces.com/Check+Extensions
//   * Fail soft - https://chessprogramming.wikispaces.com/Fail-Soft
//   * Killer move heuristic - /https://chessprogramming.wikispaces.com/Killer+Heuristic
//   * Late move redution (LMR) - https://chessprogramming.wikispaces.com/Late+Move+Reductions
//   * Negamax framework - http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
//   * Null move prunning (NMP) - https://chessprogramming.wikispaces.com/Null+Move+Pruning
//   * Principal variation search (PVS) - https://chessprogramming.wikispaces.com/Principal+Variation+Search
//   * Quiescence search - https://chessprogramming.wikispaces.com/Quiescence+Search.
//   * Hash move heuristic
package engine

import (
	"fmt"
	"time"
)

const (
	CheckDepthExtension    = 1 // how much to extend search in case of checks
	NullMoveDepthLimit     = 1 // disable null-move below this limit
	NullMoveDepthReduction = 1 // default null-move depth reduction. Can reduce more in some situations.
	PVSDepthLimit          = 0 // do not do PVS below and including this limit
	LMRDepthLimit          = 3 // do not do LMR below and including this limit
)

var (
	// scoreMultiplier is used to compute the score from side
	// to move POV from given the score from white POV.
	scoreMultiplier = [ColorArraySize]int16{0, -1, 1}
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
	Depth     int
	SelDepth  int
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
	Options  Options   // engine options
	Position *Position // current Position
	Stats    Stats     // search statistics

	evaluation Evaluation // position evaluator
	rootPly    int        // position's ply at the start of the search
	stack      stack      // stack of moves
	pvTable    pvTable    // principal variation table
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
	eng.evaluation = MakeEvaluation(eng.Position, &GlobalMaterial)
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
	return scoreMultiplier[eng.Position.SideToMove] * eng.evaluation.Evaluate()
}

// endPosition determines whether the current position is an end game.
// Returns score and a bool if the game has ended.
func (eng *Engine) endPosition() (int16, bool) {
	pos := eng.Position // shortcut
	ply := int16(eng.ply())
	if pos.NumPieces[White][King] == 0 {
		return scoreMultiplier[pos.SideToMove] * (MatedScore + ply), true
	}
	if pos.NumPieces[Black][King] == 0 {
		return scoreMultiplier[pos.SideToMove] * (MateScore - ply), true
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
		// Return mate score relative to root.
		// The score was adjusted relative to position before the
		// hash table was updated.
		if entry.Score < KnownLossScore {
			if entry.Kind == Exact {
				entry.Score += int16(eng.ply())
			}
		} else if entry.Score > KnownWinScore {
			if entry.Kind == Exact {
				entry.Score -= int16(eng.ply())
			}
		}
	} else {
		eng.Stats.CacheMiss++
	}
	return entry, ok
}

// updateHash updates GlobalHashTable with the current position.
func (eng *Engine) updateHash(α, β, depth, score int16, move Move) {
	kind := Exact
	if score <= α {
		kind = FailedLow
	} else if score >= β {
		kind = FailedHigh
	}

	// Save the mate score relative to the current position.
	// When retrieving from hash the score will be adjusted relative to root.
	if score < KnownLossScore {
		if kind == Exact {
			score -= int16(eng.ply())
		} else if kind == FailedLow {
			score = KnownLossScore
		} else {
			return
		}
	} else if score > KnownWinScore {
		if kind == Exact {
			score += int16(eng.ply())
		} else if kind == FailedHigh {
			score = KnownWinScore
		} else {
			return
		}
	}

	GlobalHashTable.Put(eng.Position, HashEntry{
		Score: score,
		Depth: depth,
		Kind:  kind,
		Move:  move,
	})
}

// searchQuiescence evaluates the position after solving all captures.
//
// This is a very limited search which considers only violent moves.
// Checks are not considered. In fact it assumes that the move
// ordering will always put the king capture first.
func (eng *Engine) searchQuiescence(α, β int16) int16 {
	eng.Stats.Nodes++
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

	var bestMove Move
	eng.stack.GenerateMoves(true, NullMove)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		eng.DoMove(move)
		score := -eng.searchQuiescence(-β, -localα)
		eng.UndoMove(move)

		if score >= β {
			return score
		}
		if score > localα {
			localα = score
			bestMove = move
		}
	}

	if α < localα && localα < β && bestMove.MoveType() != NoMove {
		eng.pvTable.Put(eng.Position, bestMove)
	}
	return localα
}

// tryMove makes a move a descends on the search tree.
//
// α, β represent lower and upper bounds.
// ply is the move number (increasing).
// depth is the fractional depth (decreasing)
// nullWindow indicates whether to scout first. Implies non-null move.
// lateMove indicates this move is late and should be reduce. Implies non-null move.
// move is the move to execute
//
// Returns the score from the deeper search.
func (eng *Engine) tryMove(α, β, depth int16, nullWindow bool, lateMove bool, move Move) int16 {
	depth--
	pos := eng.Position // shortcut
	us := pos.SideToMove
	them := us.Opposite()

	eng.DoMove(move)
	if pos.IsChecked(us) {
		// Exit early if we throw the king in check.
		eng.UndoMove(move)
		return -InfinityScore
	}
	if pos.IsChecked(them) {
		lateMove = false // tactical, dangerous

		// Extend the search when our move gives check.
		// However do not extend if we can just take the undefended piece.
		// See discussion: http://www.talkchess.com/forum/viewtopic.php?t=56361
		// TODO: This is a very crude form of SEE.
		if !pos.IsAttackedBy(move.To(), them) || pos.IsAttackedBy(move.To(), us) {
			depth += CheckDepthExtension
		}
	}

	score := α + 1
	if lateMove { // reduce late moves
		score = -eng.searchTree(-α-1, -α, depth-1, true)
	}

	if score > α { // if late move reduction is disabled or has failed
		if nullWindow {
			score = -eng.searchTree(-α-1, -α, depth, true)
			if α < score && score < β {
				score = -eng.searchTree(-β, -α, depth, true)
			}
		} else {
			score = -eng.searchTree(-β, -α, depth, move != NullMove)
		}
	}

	eng.UndoMove(move)
	return score
}

// ply returns the ply from the beginning of the search.
func (eng *Engine) ply() int {
	return eng.Position.Ply - eng.rootPly
}

// searchTree implements searchTree framework.
//
// searchTree fails soft, i.e. the score returned can be outside the bounds.
//
// α, β represent lower and upper bounds.
// ply is the move number (increasing).
// depth is the fractional depth (decreasing)
// nullMoveAllowed is true if null move is allowed, e.g. to avoid two consecutive null moves.
//
// Returns the score of the current position up to depth (modulo reductions/extensions).
// The returned score is from current player's POV.
//
// Invariants:
//   If score <= α then the search failed low and the score is an upper bound.
//   else if score >= β then the search failed high and the score is a lower bound.
//   else score is exact.
//
// Assuming this is a maximizing nodes, failing high means that an ancestors
// minimizing nodes already have a better alternative.
func (eng *Engine) searchTree(α, β, depth int16, nullMoveAllowed bool) int16 {
	// Update statistics.
	eng.Stats.Nodes++
	ply := eng.ply()
	if ply > eng.Stats.SelDepth {
		eng.Stats.SelDepth = eng.ply()
	}

	// Verify that this is not already an endgame.
	sideToMove := eng.Position.SideToMove
	if score, done := eng.endPosition(); done {
		return score
	}

	// Mate pruning: If an ancestor already has a mate in ply moves then
	// the search will always fail low so we return the lowest wining score.
	if int16(MateScore-ply) <= α {
		return KnownWinScore
	}

	// Check the transposition table.
	entry, has := eng.retrieveHash()
	hash := entry.Move
	if has && depth <= entry.Depth {
		hash = entry.Move
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

	// Stop searching when the maximum search depth is reached.
	if depth <= 0 {
		score := eng.searchQuiescence(α, β)
		eng.updateHash(α, β, depth, score, NullMove)
		return score
	}

	// Do a null move. If the null move fails high then the current
	// position is too good, so opponent will not play it.
	// Verification that we are not in check is done by tryMove
	// which bails out if after the null move we are still in check.
	if pos := eng.Position; nullMoveAllowed && // no two consective null moves
		depth > NullMoveDepthLimit && // not very close to leafs
		pos.NumPieces[sideToMove][Pawn]+1 < pos.NumPieces[sideToMove][NoPiece] && // at least one minor/major piece.
		KnownLossScore < α && β < KnownWinScore { // disable in lost or won positions

		reduction := int16(NullMoveDepthReduction)
		if pos.NumPieces[sideToMove][Pawn]+3 < pos.NumPieces[sideToMove][NoPiece] {
			// Reduce more when there are three minor/major pieces.
			reduction++
		}
		score := eng.tryMove(β-1, β, depth-reduction, false, false, NullMove)
		if score >= β {
			return score
		}
	}

	pvNode := α+1 < β
	sideIsChecked := eng.Position.IsChecked(sideToMove)
	hasGoodMoves := hash != NullMove || eng.stack.HasKiller()
	// Principal variation search: search with a null window if there is already a good move.
	nullWindow := false // updated once alpha is improved
	allowNullWindow := pvNode && hasGoodMoves && depth > PVSDepthLimit
	// Late move reduction: search best moves with full depth, reduce remaining moves.
	allowLateMove := !sideIsChecked && depth > LMRDepthLimit

	localα := α
	bestMove, bestScore := NullMove, int16(-InfinityScore)
	eng.stack.GenerateMoves(false, hash)

	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		quiet := move.IsQuiet() && move != hash && !eng.stack.IsKiller(move)

		// We reduce most quiet moves. If we already have killers or a hash move and
		// then all quiet moves are unlikely to raise α.
		lateMove := allowLateMove && quiet
		score := eng.tryMove(localα, β, depth, nullWindow, lateMove, move)

		if score >= β { // Fail high, cut node.
			eng.stack.SaveKiller(move)
			eng.updateHash(α, β, depth, score, move)
			return score
		}
		if score > bestScore {
			nullWindow = allowNullWindow
			bestMove, bestScore = move, score
			if score > localα {
				localα = score
			}
		}
	}

	// If no move was found then the game is over.
	if bestMove == NullMove {
		if sideIsChecked {
			bestScore = int16(MatedScore + ply)
		} else {
			bestScore = 0
		}
	}

	// Update hash and principal variation tables.
	eng.updateHash(α, β, depth, bestScore, bestMove)
	if α < bestScore && bestScore < β && bestMove != NullMove {
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

// search starts the search up to depth depth.
// The returned score is from current side to move POV.
// estimated is the score from previous depths.
func (eng *Engine) search(depth, estimated int16) int16 {
	// This method only implements aspiration windows.
	//
	// The gradual widening algorithm is the one used by RobboLito
	// and Stockfish and it is explained here:
	// http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=499768&t=46624
	γ, δ := int(estimated), 21
	α, β := inf(γ-δ), sup(γ+δ)
	score := estimated

	if depth < 4 {
		// Disable aspiration window for very low search depths.
		// This wastes lots of time especially for depth == 0 which is
		// used for tunning.
		α = int(-InfinityScore)
		β = int(+InfinityScore)
	}

	for {
		// At root a non-null move is required, cannot prune based on null-move.
		score = eng.searchTree(int16(α), int16(β), depth, true)

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

	fmt.Printf("info depth %d seldepth %d ", eng.Stats.Depth, eng.Stats.SelDepth)

	if score > KnownWinScore {
		fmt.Printf("score mate %d ", (MateScore-score+1)/2)
	} else if score < KnownLossScore {
		fmt.Printf("score mate %d ", (MatedScore-score)/2)
	} else {
		fmt.Printf("score cp %d ", score)
	}

	fmt.Printf("nodes %d time %d nps %d ", eng.Stats.Nodes, eng.Stats.Time(now), eng.Stats.NPS(now))

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
	eng.rootPly = eng.Position.Ply
	eng.stack.Reset(eng.Position)

	for depth := tc.NextDepth(); depth >= 0; depth = tc.NextDepth() {
		eng.Stats.Depth = depth
		score = eng.search(int16(depth), score)
		moves = eng.pvTable.Get(eng.Position)
		if eng.Options.AnalyseMode {
			eng.printInfo(score, moves)
		}
	}

	return moves
}
