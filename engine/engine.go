// Package engine implements board, move generation and position searching.
//
// The package can be used as a general library for chess tool writing and
// provides the core functionality for the zurichess chess engine.
//
// Position (basic.go, position.go) uses:
//
//   * Bitboards for representation - https://chessprogramming.wikispaces.com/Bitboards
//   * Magic bitboards for sliding move generation - https://chessprogramming.wikispaces.com/Magic+Bitboards
//
// Search (engine.go) features implemented are:
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
//   * Mate distance pruning - https://chessprogramming.wikispaces.com/Mate+Distance+Pruning
//   * Static Single Evaluation - https://chessprogramming.wikispaces.com/Static+Exchange+Evaluation
//   * Zobrist hashing - https://chessprogramming.wikispaces.com/Zobrist+Hashing
//
// Move ordering (move_ordering.go) consists of:
//
//   * Hash move heuristic
//   * Captures sorted by MVVLVA - https://chessprogramming.wikispaces.com/MVV-LVA
//   * Killer moves - https://chessprogramming.wikispaces.com/Killer+Move
//
// Evaluation (material.go) function is quite basic and consists of:
//
//   * Material and mobility
//   * Piece square tables for pawns and king. Other figures did not improve the eval.
//   * King shelter (only in mid game)
//   * Pawn structure: connected, isolated, double, passed. Evaluation is cached (see pawn_table.go).
//   * Phased eval between mid game and end game.
//
package engine

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

const (
	CheckDepthExtension    int32 = 1 // how much to extend search in case of checks
	NullMoveDepthLimit     int32 = 1 // disable null-move below this limit
	NullMoveDepthReduction int32 = 1 // default null-move depth reduction. Can reduce more in some situations.
	PVSDepthLimit          int32 = 0 // do not do PVS below and including this limit
	LMRDepthLimit          int32 = 3 // do not do LMR below and including this limit
	FutilityDepthLimit     int32 = 2 // maximum depth to do futility pruning.
)

var (
	// scoreMultiplier is used to compute the score from side
	// to move POV from given the score from white POV.
	scoreMultiplier = [ColorArraySize]int32{0, -1, 1}
)

// Options keeps engine's options.
type Options struct {
	AnalyseMode bool // true to display info strings
}

// Stats stores some basic stats of the search.
//
// Statistics are reset every iteration of the iterative deepening search.
type Stats struct {
	Start     time.Time // when the computation was started
	CacheHit  uint64    // number of times the position was found transposition table
	CacheMiss uint64    // number of times the position was not found in the transposition table
	Nodes     uint64    // number of nodes searched
	Depth     int32     // depth search
	SelDepth  int32     // maximum depth reached on PV (doesn't include the hash moves)
}

// maxDuration returns maximum of a and b.
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// NPS returns nodes per second.
func (s *Stats) NPS(now time.Time) uint64 {
	elapsed := uint64(maxDuration(now.Sub(s.Start), time.Microsecond))
	return s.Nodes * uint64(time.Second) / elapsed
}

// Time returns the number of elapsed milliseconds.
func (s *Stats) Time(now time.Time) uint64 {
	elapsed := uint64(maxDuration(now.Sub(s.Start), time.Microsecond))
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

	timeControl *TimeControl
	stopped     bool

	// A buffer to write pv lines.
	// TODO: move UCI output logic out of Engine.
	buffer bytes.Buffer
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
	eng.evaluation = MakeEvaluation(eng.Position)
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
func (eng *Engine) Score() int32 {
	return scoreMultiplier[eng.Position.SideToMove] * eng.evaluation.Evaluate()
}

// endPosition determines whether the current position is an end game.
// Returns score and a bool if the game has ended.
func (eng *Engine) endPosition() (int32, bool) {
	pos := eng.Position // shortcut
	// Trivial cases when kings are missing.
	if pos.ByPiece(White, King) == 0 && pos.ByPiece(Black, King) == 0 {
		return 0, true
	}
	if pos.ByPiece(White, King) == 0 {
		return scoreMultiplier[pos.SideToMove] * (MatedScore + eng.ply()), true
	}
	if pos.ByPiece(Black, King) == 0 {
		return scoreMultiplier[pos.SideToMove] * (MateScore - eng.ply()), true
	}
	// Neither side cannot mate.
	if pos.InsufficientMaterial() {
		return 0, true
	}
	// Fifty full moves without a capture or a pawn move.
	if pos.FiftyMoveRule() {
		return 0, true
	}
	// Repetition is a draw.
	// At root we need to continue searching even if we saw two repetitions already,
	// however we can prune deeper search only at two repetitions.
	if r := pos.ThreeFoldRepetition(); eng.ply() > 0 && r >= 2 || r >= 3 {
		return 0, true
	}
	// TODO: Handle 50 moves rule.
	return 0, false
}

// retrieveHash gets from GlobalHashTable the current position.
func (eng *Engine) retrieveHash() hashEntry {
	entry := GlobalHashTable.get(eng.Position)

	if entry.kind == noEntry {
		eng.Stats.CacheMiss++
		return hashEntry{}
	}
	if entry.move != NullMove && !eng.Position.IsValid(entry.move) {
		eng.Stats.CacheMiss++
		return hashEntry{}
	}

	// Return mate score relative to root.
	// The score was adjusted relative to position before the
	// hash table was updated.
	if entry.score < KnownLossScore {
		if entry.kind == exact {
			entry.score += eng.ply()
		}
	} else if entry.score > KnownWinScore {
		if entry.kind == exact {
			entry.score -= eng.ply()
		}
	}

	eng.Stats.CacheHit++
	return entry
}

// updateHash updates GlobalHashTable with the current position.
func (eng *Engine) updateHash(α, β, depth, score int32, move Move) {
	kind := exact
	if score <= α {
		kind = failedLow
	} else if score >= β {
		kind = failedHigh
	}

	// Save the mate score relative to the current position.
	// When retrieving from hash the score will be adjusted relative to root.
	if score < KnownLossScore {
		if kind == exact {
			score -= eng.ply()
		} else if kind == failedLow {
			score = KnownLossScore
		} else {
			return
		}
	} else if score > KnownWinScore {
		if kind == exact {
			score += eng.ply()
		} else if kind == failedHigh {
			score = KnownWinScore
		} else {
			return
		}
	}

	GlobalHashTable.put(eng.Position, hashEntry{
		kind:  kind,
		score: score,
		depth: int8(depth),
		move:  move,
	})
}

// searchQuiescence evaluates the position after solving all captures.
//
// This is a very limited search which considers only violent moves.
// Checks are not considered. In fact it assumes that the move
// ordering will always put the king capture first.
func (eng *Engine) searchQuiescence(α, β, depth int32) int32 {
	eng.Stats.Nodes++
	if score, done := eng.endPosition(); done {
		return score
	}

	// Stand pat.
	// TODO: Some suggest to not stand pat when in check.
	// However, I did several tests and handling checks in quiescence
	// doesn't help at all.
	score := eng.Score()
	if score >= β {
		return score
	}
	localα := α
	if score > localα {
		localα = score
	}

	var bestMove Move
	eng.stack.GenerateMoves(Violent, NullMove)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		if move.MoveType() == Normal && seeSign(eng.Position, move) {
			continue // Discard losing captures.
		}

		eng.DoMove(move)
		score := -eng.searchQuiescence(-β, -localα, depth-1)
		eng.UndoMove(move)

		if score >= β {
			return score
		}
		if score > localα {
			localα = score
			bestMove = move
		}
	}

	if α < localα && localα < β {
		eng.pvTable.Put(eng.Position, bestMove)
	}
	return localα
}

// tryMove makes a move a descends on the search tree.
//
// α, β represent lower and upper bounds.
// ply is the move number (increasing).
// depth is the remaining depth (decreasing)
// lmr is how much to reduce a late move. Implies non-null move.
// nullWindow indicates whether to scout first. Implies non-null move.
// move is the move to execute. Can be NullMove.
//
// Returns the score from the deeper search.
func (eng *Engine) tryMove(α, β, depth, lmr int32, nullWindow bool, move Move) int32 {
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
		lmr = 0 // tactical, dangerous

		// Extend the search when our move gives check.
		// However do not extend if we can just take the undefended piece.
		// See discussion: http://www.talkchess.com/forum/viewtopic.php?t=56361
		if pos.GetAttacker(move.To(), them) == NoFigure || pos.GetAttacker(move.To(), us) != NoFigure {
			depth += CheckDepthExtension
		}
	}

	score := α + 1
	if lmr > 0 { // reduce late moves
		score = -eng.searchTree(-α-1, -α, depth-lmr, true)
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
func (eng *Engine) ply() int32 {
	return int32(eng.Position.Ply - eng.rootPly)
}

func min(a, b int32) int32 {
	if a <= b {
		return a
	}
	return b
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
func (eng *Engine) searchTree(α, β, depth int32, nullMoveAllowed bool) int32 {
	ply := eng.ply()
	pvNode := α+1 < β

	// Update statistics.
	eng.Stats.Nodes++
	if eng.stopped || eng.Stats.Nodes%32768 == 0 && eng.timeControl.Stopped() {
		eng.stopped = true
		return α
	}
	if pvNode && ply > eng.Stats.SelDepth {
		eng.Stats.SelDepth = eng.ply()
	}

	// Verify that this is not already an endgame.
	sideToMove := eng.Position.SideToMove
	if score, done := eng.endPosition(); done {
		return score
	}

	// Mate pruning: If an ancestor already has a mate in ply moves then
	// the search will always fail low so we return the lowest wining score.
	if MateScore-ply <= α {
		return KnownWinScore
	}

	// Check the transposition table.
	entry := eng.retrieveHash()
	hash := entry.move
	if entry.kind != noEntry && depth <= int32(entry.depth) {
		if ply > 0 && entry.kind == exact {
			// Simply return if the score is exact.
			// Update principal variation table if possible.
			if α < entry.score && entry.score < β {
				eng.pvTable.Put(eng.Position, hash)
			}
			return entry.score
		}
		if entry.kind == failedLow && entry.score <= α {
			// Previously the move failed low so the actual score
			// is at most entry.score. If that's lower than α
			// this will also fail low.
			return entry.score
		}
		if entry.kind == failedHigh && entry.score >= β {
			// Previously the move failed high so the actual score
			// is at least entry.score. If that's higher than β
			// this will also fail high.
			return entry.score
		}
	}

	// Stop searching when the maximum search depth is reached.
	if depth <= 0 {
		// Depth can be < 0 due to aggressive LMR.
		score := eng.searchQuiescence(α, β, 0)
		eng.updateHash(α, β, depth, score, NullMove)
		return score
	}

	// Do a null move. If the null move fails high then the current
	// position is too good, so opponent will not play it.
	// Verification that we are not in check is done by tryMove
	// which bails out if after the null move we are still in check.
	if pos := eng.Position; nullMoveAllowed && // no two consective null moves
		depth > NullMoveDepthLimit && // not very close to leafs
		pos.HasNonPawns(sideToMove) && // at least one minor/major piece.
		KnownLossScore < α && β < KnownWinScore { // disable in lost or won positions

		reduction := NullMoveDepthReduction
		if pos.NumNonPawns(sideToMove) >= 3 {
			// Reduce more when there are three minor/major pieces.
			reduction++
		}
		score := eng.tryMove(β-1, β, depth-reduction, 0, false, NullMove)
		if score >= β {
			return score
		}
	}

	sideIsChecked := eng.Position.IsChecked(sideToMove)

	// Futility pruning at frontier nodes.
	// Disable when in check or when searching for a mate.
	if !sideIsChecked && depth <= FutilityDepthLimit && !pvNode &&
		KnownLossScore < α && β < KnownWinScore {
		if futility := eng.Score() - depth*1500; futility >= β {
			return futility
		}
	}

	hasGoodMoves := hash != NullMove || eng.stack.HasKiller()
	// Principal variation search: search with a null window if there is already a good move.
	nullWindow := false // updated once alpha is improved
	allowNullWindow := pvNode && hasGoodMoves && depth > PVSDepthLimit
	// Late move reduction: search best moves with full depth, reduce remaining moves.
	allowLateMove := !sideIsChecked && depth > LMRDepthLimit

	numQuiet := int32(0)
	localα := α
	bestMove, bestScore := NullMove, -InfinityScore

	eng.stack.GenerateMoves(All, hash)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		// Reduce most quiet moves and bad captures.
		lmr := int32(0)
		if allowLateMove && move != hash && !eng.stack.IsKiller(move) {
			if move.IsQuiet() {
				// Reduce quiet moves more at high depths and after many quiet moves.
				// Large numQuiet means it's likely not a CUT node.
				// Large depth means reductions are less risky.
				numQuiet++
				tmp := 1 + min(depth, numQuiet)/5
				if tmp != 1 && seeSign(eng.Position, move) {
					lmr = tmp
				} else {
					lmr = 1
				}
			} else if seeSign(eng.Position, move) {
				// Bad captures (SEE<0) can be reduced, too.
				lmr = 1
			}
		}

		score := eng.tryMove(localα, β, depth, lmr, nullWindow, move)
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
			bestScore = MatedScore + ply
		} else {
			bestScore = 0
		}
	}

	// Update hash and principal variation tables.
	eng.updateHash(α, β, depth, bestScore, bestMove)
	if α < bestScore && bestScore < β {
		eng.pvTable.Put(eng.Position, bestMove)
	}

	return bestScore
}

func inf(a int32) int32 {
	if a <= -InfinityScore {
		return -InfinityScore
	}
	return a
}

func sup(b int32) int32 {
	if b >= InfinityScore {
		return InfinityScore
	}
	return b
}

// search starts the search up to depth depth.
// The returned score is from current side to move POV.
// estimated is the score from previous depths.
func (eng *Engine) search(depth, estimated int32) int32 {
	// This method only implements aspiration windows.
	//
	// The gradual widening algorithm is the one used by RobboLito
	// and Stockfish and it is explained here:
	// http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=499768&t=46624
	γ, δ := estimated, int32(210)
	α, β := inf(γ-δ), sup(γ+δ)
	score := estimated

	if depth < 4 {
		// Disable aspiration window for very low search depths.
		// This wastes lots of time especially for depth == 0 which is
		// used for tunning.
		α = -InfinityScore
		β = +InfinityScore
	}

	for !eng.stopped {
		// At root a non-null move is required, cannot prune based on null-move.
		score = eng.searchTree(α, β, depth, true)

		if score <= α {
			α = inf(α - δ)
			δ += δ / 2
		} else if score >= β {
			β = sup(β + δ)
			δ += δ / 2
		} else {
			return score
		}
	}

	return score
}

// printInfo prints a info UCI string.
//
// TODO: Engine shouldn't know about the protocol used.
func (eng *Engine) printInfo(score int32, pv []Move) {
	buf := &eng.buffer // shortcut

	// Write depth.
	now := time.Now()
	fmt.Fprintf(buf, "info depth %d seldepth %d ", eng.Stats.Depth, eng.Stats.SelDepth)

	// Write score.
	if score > KnownWinScore {
		fmt.Fprintf(buf, "score mate %d ", (MateScore-score+1)/2)
	} else if score < KnownLossScore {
		fmt.Fprintf(buf, "score mate %d ", (MatedScore-score)/2)
	} else {
		fmt.Fprintf(buf, "score cp %d ", score/10)
	}

	// Write stats.
	fmt.Fprintf(buf, "nodes %d time %d nps %d ", eng.Stats.Nodes, eng.Stats.Time(now), eng.Stats.NPS(now))

	// Write principal variation.
	fmt.Fprintf(buf, "pv")
	for _, m := range pv {
		fmt.Fprintf(buf, " %v", m.UCI())
	}
	fmt.Fprintf(buf, "\n")
}

// Play evaluates current position.
//
// Returns the principal variation, that is
//	moves[0] is the best move found and
//	moves[1] is the pondering move.
//
// If no move was found because the game has finished
// then an empty pv is returned.
//
// Time control, tc, should already be started.
func (eng *Engine) Play(tc *TimeControl) (moves []Move) {
	now := time.Now()
	eng.Stats = Stats{Start: now, Depth: -1}
	eng.rootPly = eng.Position.Ply
	eng.timeControl = tc
	eng.stopped = false
	eng.stack.Reset(eng.Position)
	eng.buffer.Reset()

	silent := true
	score := int32(0)
	for depth := int32(0); depth < 64; depth++ {
		if !tc.NextDepth(depth) {
			// Stop if tc control says we are done.
			// Search at least one depth, otherwise a move cannot be returned.
			break
		}

		eng.Stats.Depth = depth
		score = eng.search(depth, score)

		if !eng.stopped {
			moves = eng.pvTable.Get(eng.Position)
			if eng.Options.AnalyseMode {
				eng.printInfo(score, moves)
			}
		}

		if !silent || now.Add(2*time.Second).After(time.Now()) {
			// Delay first output because first plies produce a lot of noise.
			silent = false
			eng.flush()
		}
	}

	eng.flush()
	return moves
}

// Flush writes the buffer to stdout.
func (eng *Engine) flush() {
	os.Stdout.Write(eng.buffer.Bytes())
	os.Stdout.Sync()
	eng.buffer.Reset()
}
