// Package engine implements board, moves and search.
package engine

import (
	"fmt"
	"time"
)

const (
	DepthMultiplier     = 4
	CheckDepthExtension = 3
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

// Nodes returns nodes per second.
func (s *Stats) NPS(now time.Time) uint64 {
	elapsed := uint64(time.Now().Sub(s.Start) + 1)
	return s.Nodes * uint64(time.Second) / elapsed
}

func (s *Stats) Time(now time.Time) uint64 {
	elapsed := uint64(time.Now().Sub(s.Start) + 1)
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

	maxPly       int16 // max ply currently searching at.
	scoreMidGame int
	scoreEndGame int

	// killer stores a few killer moves per ply.
	// For killer heuristic see https://chessprogramming.wikispaces.com/Killer+Heuristic
	killer [][2]Move

	stack   moveStack
	pvTable pvTable                              // principal variation table
	pieces  [ColorArraySize][FigureArraySize]int // number of pieces
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
// mg and eg are move's midgame and endgame scores adjust accordingly
// whether the move is made or taken back.
func (eng *Engine) adjust(move Move, delta, mg, eg int) {
	eng.scoreMidGame += mg
	eng.scoreEndGame += eg
	if move.MoveType == Promotion {
		eng.put(move.Target.Color(), Pawn, -delta)
		eng.put(move.Target.Color(), move.Target.Figure(), delta)
	}
	if move.Capture != NoPiece {
		eng.put(move.Capture.Color(), move.Capture.Figure(), -delta)
	}
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	mg, eg := eng.evaluateMove(move)
	eng.adjust(move, +1, +mg, +eg)
	eng.Position.DoMove(move)
}

// UndoMove undoes the last move.
func (eng *Engine) UndoMove(move Move) {
	eng.Position.UndoMove(move)
	mg, eg := eng.evaluateMove(move)
	eng.adjust(move, -1, -mg, -eg)
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

// pawns computes the pawn structure score of side.
// pawns awards chains and penalizes double pawns.
func (eng *Engine) pawns(side Color) int {
	pawns := eng.Position.ByPiece(side, Pawn)
	forward := pawns
	if side == White {
		forward >>= 8
	} else {
		forward <<= 8
	}

	cs := (pawns & ((BbPawnLeftAttack & forward) >> 1)).Popcnt()
	cs += (pawns & ((BbPawnRightAttack & forward) << 1)).Popcnt()
	ds := (pawns & forward).Popcnt()
	return cs*PawnChainBonus - ds*DoublePawnPenalty
}

// phase returns current phase and total phase.
//
// phase is determined by the number of pieces left in the game where
// pawn has score 0, knight and bishop 1, rook 2, queen 2.
// See tapered eval: // https://chessprogramming.wikispaces.com/Tapered+Eval
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

// Score evaluates current position from White's POV.
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

	score += eng.pawns(White)
	score -= eng.pawns(Black)
	return int16(score)
}

// endPosition determines whether current position is an end game.
// Returns score and a bool if the game has ended.
func (eng *Engine) endPosition() (int16, bool) {
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
		Target: move.Target,
		From:   move.From,
		To:     move.To,
	})
}

// quiescence searches a quite move.
func (eng *Engine) quiescence(α, β, ply int16) int16 {
	color := eng.Position.SideToMove
	score := int16(colorWeight[color]) * eng.Score()
	if score >= β {
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
		if eng.Position.IsChecked(color) {
			eng.Position.UndoMove(move)
			continue
		}

		mg, eg := eng.evaluateMove(move)
		eng.adjust(move, +1, +mg, +eg)
		score := -eng.quiescence(-β, -localα, ply+1)
		eng.adjust(move, -1, -mg, -eg)
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

// evaluateMove returns mid game and eng game score of the move.
// scores returned to not depend on the current position.
func (eng *Engine) evaluateMove(move Move) (int, int) {
	return MidGameMaterial.EvaluateMove(move), EndGameMaterial.EvaluateMove(move)
}

func (eng *Engine) tryMove(α, β, ply, depth int16, move Move) int16 {
	side := eng.Position.SideToMove
	eng.Position.DoMove(move)
	if eng.Position.IsChecked(side) {
		eng.Position.UndoMove(move)
		return -InfinityScore
	}

	mg, eg := eng.evaluateMove(move)
	eng.adjust(move, +1, +mg, +eg)
	score := -eng.negamax(-β, -α, ply+1, depth-DepthMultiplier)
	eng.adjust(move, -1, -mg, -eg)
	eng.Position.UndoMove(move)

	if score > KnownWinScore {
		// If the position is a win the score is decreased
		// slightly so the search takes the shortest path.
		score--
	}
	return score
}

// saveKiller saves a killer move.
func (eng *Engine) saveKiller(ply int16, move Move) {
	for len(eng.killer) <= int(ply) {
		eng.killer = append(eng.killer, [2]Move{})
	}
	if move.Capture == NoPiece && move != eng.killer[ply][0] { // saves only quiet moves.
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
			if m.Target == entry.Target && m.From == entry.From && m.To == entry.To {
				return HashMoveBonus
			}
			if len(eng.killer) > int(ply) {
				for _, k := range eng.killer[ply] {
					if m.Target == k.Target && m.From == k.From && m.To == k.To {
						return KillerMoveBonus
					}
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
func (eng *Engine) negamax(α, β, ply, depth int16) int16 {
	sideToMove := eng.Position.SideToMove
	if score, done := eng.endPosition(); done {
		return int16(colorWeight[sideToMove]) * score
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

	if eng.Position.IsChecked(sideToMove) {
		// Extend search when the side to move is in check.
		// https://chessprogramming.wikispaces.com/Check+Extensions
		depth += CheckDepthExtension
	}

	if depth <= 0 {
		// Stop searching when maximum depth is reached.
		score := eng.quiescence(α, β, 0)
		eng.updateHash(α, β, ply, score, &Move{})
		return score
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
		score := eng.negamax(int16(α), int16(β), 0, eng.maxPly*DepthMultiplier)
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

func (eng *Engine) printInfo(score int16) {
	now := time.Now()
	fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d ",
		eng.maxPly, score, eng.Stats.Nodes, eng.Stats.Time(now), eng.Stats.NPS(now))

	fmt.Printf("pv")
	for _, move := range eng.pvTable.Get(eng.Position) {
		fmt.Printf(" %v", move.UCI())
	}
	fmt.Printf("\n")
}

// Play evaluates current position.
// Returns principal variation, moves[0] is the next move.
//
// tc should already be started.
func (eng *Engine) Play(tc TimeControl) (moves []Move) {
	score := int16(0)
	eng.Stats = Stats{Start: time.Now()}

	for maxPly := tc.NextDepth(); maxPly >= 0; maxPly = tc.NextDepth() {
		eng.maxPly = int16(maxPly)
		score = eng.alphaBeta(score)
		moves = eng.pvTable.Get(eng.Position)
		if eng.Options.AnalyseMode {
			eng.printInfo(score)
		}
	}

	if !eng.Options.AnalyseMode {
		// eng.printInfo(score)
	}
	return moves
}
