// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
//   * Futility Pruning - https://chessprogramming.wikispaces.com/Futility+pruning
//   * History leaf pruning - https://chessprogramming.wikispaces.com/History+Leaf+Pruning
//   * Killer move heuristic - https://chessprogramming.wikispaces.com/Killer+Heuristic
//   * Late move redution (LMR) - https://chessprogramming.wikispaces.com/Late+Move+Reductions
//   * Mate distance pruning - https://chessprogramming.wikispaces.com/Mate+Distance+Pruning
//   * Negamax framework - http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
//   * Null move prunning (NMP) - https://chessprogramming.wikispaces.com/Null+Move+Pruning
//   * Principal variation search (PVS) - https://chessprogramming.wikispaces.com/Principal+Variation+Search
//   * Quiescence search - https://chessprogramming.wikispaces.com/Quiescence+Search
//   * Razoring - https://chessprogramming.wikispaces.com/Razoring
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
//   * King safery ala Toga style - https://chessprogramming.wikispaces.com/King+Safety#Attacking%20King%20Zone
//   * Pawn structure: connected, isolated, double, passed. Evaluation is cached (see cache.go).
//   * Phased eval between mid game and end game.
//
package engine

import (
	"math/rand"

	. "bitbucket.org/zurichess/zurichess/board"
)

const (
	checkDepthExtension int32 = 1 // how much to extend search in case of checks
	lmrDepthLimit       int32 = 4 // do not do LMR below and including this limit
	futilityDepthLimit  int32 = 4 // maximum depth to do futility pruning.

	initialAspirationWindow = 13 // ~a quarter of a pawn
	futilityMargin          = 75 // ~one and a halfpawn
	checkpointStep          = 10000
)

// Options keeps engine's options.
type Options struct {
	AnalyseMode   bool // true to display info strings
	MultiPV       int
	HandicapLevel int
}

// Stats stores statistics about the search.
type Stats struct {
	CacheHit  uint64 // number of times the position was found transposition table
	CacheMiss uint64 // number of times the position was not found in the transposition table
	Nodes     uint64 // number of nodes searched
	Depth     int32  // depth search
	SelDepth  int32  // maximum depth reached on PV (doesn't include the hash moves)
}

// CacheHitRatio returns the ratio of transposition table hits over total number of lookups.
func (s *Stats) CacheHitRatio() float32 {
	return float32(s.CacheHit) / float32(s.CacheHit+s.CacheMiss)
}

// Logger logs search progress.
type Logger interface {
	// BeginSearch signals a new search is started.
	BeginSearch()
	// EndSearch signals end of search.
	EndSearch()
	// PrintPV logs the principal variation after iterative deepening completed one depth.
	PrintPV(stats Stats, multiPV int, score int32, pv []Move)
	// CurrMove logs the current move. Current move index is 1-based.
	CurrMove(depth int, move Move, num int)
}

// NulLogger is a logger that does nothing.
type NulLogger struct{}

func (nl *NulLogger) BeginSearch()                                             {}
func (nl *NulLogger) EndSearch()                                               {}
func (nl *NulLogger) PrintPV(stats Stats, multiPV int, score int32, pv []Move) {}
func (nl *NulLogger) CurrMove(depth int, move Move, num int)                   {}

// historyEntry keeps counts of how well move performed in the past.
type historyEntry struct {
	stat int32
	move Move
}

// historyTable is a hash table that contains history of moves.
//
// old moves are automatically evicted when new moves are inserted
// so this cache is approx. LRU.
type historyTable [1 << 12]historyEntry

// historyHash hashes the move and returns an index into the history table.
func historyHash(m Move) uint32 {
	// This is a murmur inspired hash so upper bits are better
	// mixed than the lower bits. The hash multiplier was chosen
	// to minimize the number of misses.
	h := uint32(m) * 438650727
	return (h + (h << 17)) >> 20
}

// newSearch updates the stats before a new search.
func (ht *historyTable) newSearch() {
	for i := range ht {
		ht[i].stat /= 8
	}
}

// get returns the stats for a move, m.
// If the move is not in the table, returns 0.
func (ht *historyTable) get(m Move) int32 {
	h := historyHash(m)
	if ht[h].move != m {
		return 0
	}
	return ht[h].stat
}

// inc increments the counters for m.
// Evicts an old move if necessary.
func (ht *historyTable) add(m Move, delta int32) {
	h := historyHash(m)
	if ht[h].move != m {
		ht[h] = historyEntry{stat: delta, move: m}
	} else {
		ht[h].stat += delta
	}
}

// Engine implements the logic to search for the best move for a position.
type Engine struct {
	Options  Options   // engine options
	Log      Logger    // logger
	Stats    Stats     // search statistics
	Position *Position // current Position

	rootPly         int           // position's ply at the start of the search
	stack           stack         // stack of moves
	pvTable         pvTable       // principal variation table
	history         *historyTable // keeps history of moves
	ignoreRootMoves []Move        // moves to ignore at root

	timeControl *TimeControl
	stopped     bool
	checkpoint  uint64
}

// NewEngine creates a new engine to search for pos.
// If pos is nil then the start position is used.
func NewEngine(pos *Position, log Logger, options Options) *Engine {
	if options.MultiPV == 0 {
		options.MultiPV = 1
	}

	if log == nil {
		log = &NulLogger{}
	}
	history := &historyTable{}
	eng := &Engine{
		Options: options,
		Log:     log,
		pvTable: newPvTable(),
		history: history,
		stack:   stack{history: history},
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
	GlobalHashTable.prefetch(eng.Position)
}

// UndoMove undoes the last move.
func (eng *Engine) UndoMove() {
	eng.Position.UndoMove()
}

// Score evaluates current position from current player's POV.
func (eng *Engine) Score() int32 {
	return Evaluate(eng.Position).GetCentipawnsScore() * eng.Position.Us().Multiplier()
}

// cachedScore implements a cache on top of Score.
// The cached static evaluation is stored in the hashEntry.
func (eng *Engine) cachedScore(e *hashEntry) int32 {
	if e.kind&hasStatic == 0 {
		e.kind |= hasStatic
		e.static = int16(eng.Score())
	}
	return int32(e.static)
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
		return pos.Us().Multiplier() * (MatedScore + eng.ply()), true
	}
	if pos.ByPiece(Black, King) == 0 {
		return pos.Us().Multiplier() * (MateScore - eng.ply()), true
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
	return 0, false
}

// retrieveHash gets from GlobalHashTable the current position.
func (eng *Engine) retrieveHash() hashEntry {
	entry := GlobalHashTable.get(eng.Position)
	if entry.kind == 0 || entry.move != NullMove && !eng.Position.IsPseudoLegal(entry.move) {
		eng.Stats.CacheMiss++
		return hashEntry{}
	}

	// Return mate score relative to root.
	// The score was adjusted relative to position before the hash table was updated.
	if entry.score < KnownLossScore {
		entry.score += int16(eng.ply())
	} else if entry.score > KnownWinScore {
		entry.score -= int16(eng.ply())
	}

	eng.Stats.CacheHit++
	return entry
}

// updateHash updates GlobalHashTable with the current position.
func (eng *Engine) updateHash(flags hashFlags, depth, score int32, move Move, static int32) {
	// If search is stopped then score cannot be trusted.
	if eng.stopped {
		return
	}
	// Update principal variation table in exact nodes.
	if flags&exact != 0 {
		eng.pvTable.Put(eng.Position, move)
	}
	if eng.ply() == 0 && len(eng.ignoreRootMoves) != 0 {
		// At root if there are moves to ignore (e.g. because of multipv)
		// then this is an incomplete search, so don't update the hash.
		return
	}

	// Save the mate score relative to the current position.
	// When retrieving from hash the score will be adjusted relative to root.
	if score < KnownLossScore {
		score -= eng.ply()
	} else if score > KnownWinScore {
		score += eng.ply()
	}

	GlobalHashTable.put(eng.Position, hashEntry{
		kind:   flags,
		score:  int16(score),
		depth:  int8(depth),
		move:   move,
		static: int16(static),
	})
}

// searchQuiescence evaluates the position after solving all captures.
//
// This is a very limited search which considers only some violent moves.
// Depth is ignored, so hash uses depth 0. Search continues until
// stand pat or no capture can improve the score.
func (eng *Engine) searchQuiescence(α, β int32) int32 {
	eng.Stats.Nodes++

	entry := eng.retrieveHash()
	if score := int32(entry.score); isInBounds(entry.kind, α, β, score) {
		return score
	}

	static := eng.cachedScore(&entry)
	if static >= β {
		// Stand pat if static score is already a cut-off.
		eng.updateHash(failedHigh|hasStatic, 0, static, NullMove, static)
		return static
	}

	pos := eng.Position
	us := pos.Us()
	inCheck := pos.IsChecked(us)
	localα := max(α, static)

	var bestMove Move
	eng.stack.GenerateMoves(Violent, NullMove)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		// Prune futile moves that would anyway result in a stand-pat at that next depth.
		if !inCheck && isFutile(pos, static, α, futilityMargin, move) ||
			!inCheck && seeSign(pos, move) {
			continue
		}

		// Discard illegal or losing captures.
		eng.DoMove(move)
		if eng.Position.IsChecked(us) {
			eng.UndoMove()
			continue
		}
		score := -eng.searchQuiescence(-β, -localα)
		eng.UndoMove()

		if score >= β {
			eng.updateHash(failedHigh|hasStatic, 0, score, move, static)
			return score
		}
		if score > localα {
			localα = score
			bestMove = move
		}
	}

	eng.updateHash(getBound(α, β, localα)|hasStatic, 0, localα, bestMove, static)
	return localα
}

// tryMove descends on the search tree. This function
// is called from searchTree after the move is executed
// and it will undo the move.
//
// α, β represent lower and upper bounds.
// depth is the remaining depth (decreasing)
// lmr is how much to reduce a late move. Implies non-null move.
// nullWindow indicates whether to scout first. Implies non-null move.
//
// Returns the score from the deeper search.
func (eng *Engine) tryMove(α, β, depth, lmr int32, nullWindow bool) int32 {
	depth--

	score := α + 1
	if lmr > 0 { // reduce late moves
		score = -eng.searchTree(-α-1, -α, depth-lmr)
	}

	if score > α { // if late move reduction is disabled or has failed
		if nullWindow {
			score = -eng.searchTree(-α-1, -α, depth)
			if α < score && score < β {
				score = -eng.searchTree(-β, -α, depth)
			}
		} else {
			score = -eng.searchTree(-β, -α, depth)
		}
	}

	eng.UndoMove()
	return score
}

// ply returns the ply from the beginning of the search.
func (eng *Engine) ply() int32 {
	return int32(eng.Position.Ply - eng.rootPly)
}

// isIgnoredRootMove returns true if move should be ignored at root.
func (eng *Engine) isIgnoredRootMove(move Move) bool {
	if eng.ply() != 0 {
		return false
	}
	for _, m := range eng.ignoreRootMoves {
		if m == move {
			return true
		}
	}
	return false
}

// searchTree implements searchTree framework.
//
// searchTree fails soft, i.e. the score returned can be outside the bounds.
//
// α, β represent lower and upper bounds.
// depth is the search depth (decreasing)
//
// Returns the score of the current position up to depth (modulo reductions/extensions).
// The returned score is from current player's POV.
//
// Invariants:
//   If score <= α then the search failed low and the score is an upper bound.
//   else if score >= β then the search failed high and the score is a lower bound.
//   else score is exact.
//
// Assuming this is a maximizing nodes, failing high means that a
// minimizing ancestor node already has a better alternative.
func (eng *Engine) searchTree(α, β, depth int32) int32 {
	ply := eng.ply()
	pvNode := α+1 < β
	pos := eng.Position
	us := pos.Us()

	// Update statistics.
	eng.Stats.Nodes++
	if !eng.stopped && eng.Stats.Nodes >= eng.checkpoint {
		eng.checkpoint = eng.Stats.Nodes + checkpointStep
		if eng.timeControl.Stopped() {
			eng.stopped = true
		}
	}
	if eng.stopped {
		return α
	}
	if pvNode && ply > eng.Stats.SelDepth {
		eng.Stats.SelDepth = ply
	}

	// Verify that this is not already an endgame.
	if score, done := eng.endPosition(); done && (ply != 0 || score != 0) {
		// At root we ignore draws because some GUIs don't properly detect
		// theoretical draws. E.g. cutechess doesn't detect that kings and
		// bishops when all bishops are on the same color. If the position
		// is a theoretical draw, keep searching for a move.
		return score
	}

	// Mate pruning: If an ancestor already has a mate in ply moves then
	// the search will always fail low so we return the lowest wining score.
	if MateScore-ply <= α {
		return KnownWinScore
	}

	// Stop searching when the maximum search depth is reached.
	// Depth can be < 0 due to aggressive LMR.
	if depth <= 0 {
		return eng.searchQuiescence(α, β)
	}

	// Check the transposition table.
	// Entry will store the cached static evaluation which may be computed later.
	entry := eng.retrieveHash()
	hash := entry.move
	if eng.isIgnoredRootMove(hash) {
		entry = hashEntry{}
		hash = NullMove
	}
	if score := int32(entry.score); depth <= int32(entry.depth) && isInBounds(entry.kind, α, β, score) {
		if pvNode {
			// Update the pv table, otherwise we risk not having a node at root
			// if the pv entry was overwritten.
			eng.pvTable.Put(pos, hash)
		}
		if score >= β && hash != NullMove {
			// If this is CUT node, update the killer like in the regular move loop.
			eng.stack.SaveKiller(hash)
		}
		return score
	}

	sideIsChecked := pos.IsChecked(us)

	// Do a null move. If the null move fails high then the current
	// position is too good, so opponent will not play it.
	// Verification that we are not in check is done by tryMove
	// which bails out if after the null move we are still in check.
	if !sideIsChecked && // nullmove is illegal when in check
		MinorsAndMajors(pos, us) != 0 && // at least one minor/major piece.
		KnownLossScore < α && β < KnownWinScore && // disable in lost or won positions
		(entry.kind&hasStatic == 0 || int32(entry.static) >= β) {
		eng.DoMove(NullMove)
		reduction := 1 + depth/3
		score := eng.tryMove(β-1, β, depth-reduction, 0, false)
		if score >= β && score < KnownWinScore {
			return score
		}
	}

	// Razoring at very low depth: if QS is under a considerable margin
	// we return that score.
	if depth == 1 &&
		!sideIsChecked && // disable in check
		!pvNode && // disable in pv nodes
		KnownLossScore < α && β < KnownWinScore { // disable when searching for a mate
		rα := α - futilityMargin
		if score := eng.searchQuiescence(rα, rα+1); score <= rα {
			return score
		}
	}

	// Futility and history pruning at frontier nodes.
	// Based on Deep Futility Pruning http://home.hccnet.nl/h.g.muller/deepfut.html
	// Based on History Leaf Pruning https://chessprogramming.wikispaces.com/History+Leaf+Pruning
	// Statically evaluates the position. Use static evaluation from hash if available.
	static := int32(0)
	allowLeafsPruning := false
	if depth <= futilityDepthLimit && // enable when close to the frontier
		!sideIsChecked && // disable in check
		!pvNode && // disable in pv nodes
		KnownLossScore < α && β < KnownWinScore { // disable when searching for a mate
		allowLeafsPruning = true
		static = eng.cachedScore(&entry)
	}

	// Principal variation search: search with a null window if there is already a good move.
	bestMove, localα := NullMove, int32(-InfinityScore)
	// dropped true if not all moves were searched.
	// Mate cannot be declared unless all moves were tested.
	dropped := false
	numMoves := int32(0)

	eng.stack.GenerateMoves(Violent|Quiet, hash)
	for move := eng.stack.PopMove(); move != NullMove; move = eng.stack.PopMove() {
		if ply == 0 {
			eng.Log.CurrMove(int(depth), move, int(numMoves+1))
		}
		if eng.isIgnoredRootMove(move) {
			continue
		}

		givesCheck := pos.GivesCheck(move)
		critical := move == hash || eng.stack.IsKiller(move)
		history := eng.history.get(move)
		newDepth := depth
		numMoves++

		if allowLeafsPruning && !critical && !givesCheck && localα > KnownLossScore {
			// Prune moves that do not raise alphas and moves that performed bad historically.
			// Prune bad captures moves that performed bad historically.
			if isFutile(pos, static, α, depth*futilityMargin, move) ||
				history < -10 && move.IsQuiet() ||
				see(pos, move) < -futilityMargin {
				dropped = true
				continue
			}
		}

		// Extend good moves that also gives check.
		// See discussion: http://www.talkchess.com/forum/viewtopic.php?t=56361
		// When the move gives check, history pruning and futility pruning are also disabled.
		if givesCheck && !seeSign(pos, move) {
			newDepth += checkDepthExtension
			critical = true
		}

		// Late move reduction: search best moves with full depth, reduce remaining moves.
		lmr := int32(0)
		if !sideIsChecked && depth > lmrDepthLimit && !critical {
			// Reduce quiet moves and bad captures more at high depths and after many quiet moves.
			// Large numMoves means it's likely not a CUT node.  Large depth means reductions are less risky.
			if move.IsQuiet() {
				lmr = 2 + min(depth, numMoves)/6
			} else if see := see(pos, move); see < -futilityMargin {
				lmr = 2 + min(depth, numMoves)/6
			} else if see < 0 {
				lmr = 1 + min(depth, numMoves)/6
			}
		}

		// Skip illegal moves that leave the king in check.
		eng.DoMove(move)
		if pos.IsChecked(us) {
			eng.UndoMove()
			continue
		}

		score := eng.tryMove(max(α, localα), β, newDepth, lmr, numMoves > 1)

		if score >= β {
			// Fail high, cut node.
			eng.history.add(move, 5+5*depth)
			eng.stack.SaveKiller(move)
			eng.updateHash(failedHigh|(entry.kind&hasStatic), depth, score, move, int32(entry.static))
			return score
		}
		if score > localα {
			bestMove, localα = move, score
		}
		eng.history.add(move, -1)
	}

	bound := getBound(α, β, localα)
	if !dropped && bestMove == NullMove {
		// If no move was found then the game is over.
		bound = exact
		if sideIsChecked {
			localα = MatedScore + ply
		} else {
			localα = 0
		}
	}

	eng.updateHash(bound|(entry.kind&hasStatic), depth, localα, bestMove, int32(entry.static))
	return localα
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
	γ, δ := estimated, int32(initialAspirationWindow)
	α, β := max(γ-δ, -InfinityScore), min(γ+δ, InfinityScore)
	score := estimated

	if depth < 4 {
		// Disable aspiration window for very low search depths.
		α, β = -InfinityScore, +InfinityScore
	}

	for !eng.stopped {
		// At root a non-null move is required, cannot prune based on null-move.
		score = eng.searchTree(α, β, depth)
		if score <= α {
			α = max(α-δ, -InfinityScore)
			δ += δ / 2
		} else if score >= β {
			β = min(β+δ, InfinityScore)
			δ += δ / 2
		} else {
			return score
		}
	}

	return score
}

// searchMultiPV searches eng.options.MultiPV principal variations from current position.
// Returns score and the moves of the highest scoring pv line (possible empty).
// If a pv is not found (e.g. search is stopped during the first ply), return 0, nil.
func (eng *Engine) searchMultiPV(depth, estimated int32) (int32, []Move) {
	type pv struct {
		score int32
		moves []Move
	}

	multiPV := eng.Options.MultiPV
	searchMultiPV := (eng.Options.HandicapLevel+4)/5 + 1
	if multiPV < searchMultiPV {
		multiPV = searchMultiPV
	}

	pvs := make([]pv, 0, multiPV)
	eng.ignoreRootMoves = eng.ignoreRootMoves[:0]
	for p := 0; p < multiPV; p++ {
		estimated = eng.search(depth, estimated)
		if eng.stopped {
			break // if eng has been stopped then this is not a legit pv.
		}

		moves := eng.pvTable.Get(eng.Position)
		hasPV := len(moves) != 0 && !eng.isIgnoredRootMove(moves[0])
		if p == 0 || hasPV { // at depth 0 we might not get a PV
			pvs = append(pvs, pv{estimated, moves})
		}
		if !hasPV {
			break
		}
		// if there is PV ignore the first move for the next PVs
		eng.ignoreRootMoves = append(eng.ignoreRootMoves, moves[0])
	}

	// Sort PVs by score.
	if len(pvs) == 0 {
		return 0, nil
	}
	for i := range pvs {
		for j := i; j >= 0; j-- {
			if j == 0 || pvs[j-1].score > pvs[i].score {
				tmp := pvs[i]
				copy(pvs[j+1:i+1], pvs[j:i])
				pvs[j] = tmp
				break
			}
		}
	}

	for i := range pvs {
		eng.Log.PrintPV(eng.Stats, i+1, pvs[i].score, pvs[i].moves)
	}

	// For best play return the PV with highest score.
	if eng.Options.HandicapLevel == 0 || len(pvs) <= 1 {
		return pvs[0].score, pvs[0].moves
	}

	// PVs are sorted by score. Pick one PV at random
	// and if the score is not too far off, return it.
	s := int32(eng.Options.HandicapLevel)
	d := s*s/2 + s*10 + 5
	n := rand.Intn(len(pvs))
	if rand.Intn(eng.Options.HandicapLevel) == 0 {
		if n1 := rand.Intn(len(pvs)); n1 > n {
			n1 = n
		}
	}
	for pvs[n].score+d < pvs[0].score {
		n--
	}
	return pvs[n].score, pvs[n].moves
}

// Play evaluates current position.
//
// Returns the principal variation, that is
//      moves[0] is the best move found and
//      moves[1] is the pondering move.
//
// Returns a nil pv if no move was found because the game is already finished.
// Returns empty pv array if it's valid position, but no pv was found (e.g. search depth is 0).
//
// Time control, tc, should already be started.
func (eng *Engine) Play(tc *TimeControl) (score int32, moves []Move) {
	eng.Log.BeginSearch()
	eng.Stats = Stats{Depth: -1}

	eng.rootPly = eng.Position.Ply
	eng.timeControl = tc
	eng.stopped = false
	eng.checkpoint = checkpointStep
	eng.stack.Reset(eng.Position)
	eng.history.newSearch()

	for depth := int32(0); depth < 64; depth++ {
		if !tc.NextDepth(depth) {
			// Stop if tc control says we are done.
			// Search at least one depth, otherwise a move cannot be returned.
			break
		}

		eng.Stats.Depth = depth
		if s, m := eng.searchMultiPV(depth, score); len(moves) == 0 || len(m) != 0 {
			score, moves = s, m
		}
	}

	eng.Log.EndSearch()
	if len(moves) == 0 && !eng.Position.HasLegalMoves() {
		return 0, nil
	} else if moves == nil {
		return score, []Move{}
	}
	return score, moves
}

// isFutile return true if m cannot raise current static
// evaluation above α. This is just an heuristic and mistakes
// can happen.
func isFutile(pos *Position, static, α, margin int32, m Move) bool {
	if m.MoveType() == Promotion || m.Piece().Figure() == Pawn && BbPawnStartRank.Has(m.To()) {
		// Promotion and passed pawns can increase the static evaluation
		// by more than futilityMargin.
		return false
	}
	δ := futilityFigureBonus[m.Capture().Figure()]
	return static+δ+margin < α
}
