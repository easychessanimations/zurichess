package engine

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

var _ = log.Println
var _ = fmt.Println

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")

	figureBonus = [FigureMaxValue]int{
		0,     // NoFigure
		100,   // Pawn
		400,   // Knight
		400,   // Bishop
		600,   // Rook
		1200,  // Queen
		10000, // King
	}

	knownWinScore   = 20000
	mateScore       = 30000
	bishopPairBonus = 50
	knightPawnBonus = 6
	rookPawnPenalty = 12
)

type Engine struct {
	position *Position // current position
	moves    []Move    // moves stack
	nodes    uint64    // number of nodes evaluated

	pieces     [ColorMaxValue][FigureMaxValue]int
	pieceScore int
}

// NewEngine returns a new engine for pos.
func NewEngine(pos *Position) *Engine {
	eng := &Engine{
		position: pos,
		moves:    make([]Move, 0, 64),
	}
	eng.countMaterial()
	return eng
}

// ParseMove parses the move from a string.
func (eng *Engine) ParseMove(move string) Move {
	return eng.position.ParseMove(move)
}

func (eng *Engine) put(col Color, pi Figure) {
	eng.pieces[col][pi]++
	eng.pieceScore += ColorWeight[col] * figureBonus[pi]
}

func (eng *Engine) remove(col Color, pi Figure) {
	eng.pieces[col][pi]--
	eng.pieceScore -= ColorWeight[col] * figureBonus[pi]
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	capt := move.Capture
	if capt != NoPiece {
		eng.remove(capt.Color(), capt.Figure())
	}
	if move.MoveType == Promotion {
		eng.remove(eng.position.ToMove, Pawn)
		eng.put(eng.position.ToMove, move.Promotion.Figure())
	}
	eng.position.DoMove(move)
}

// UndoMove takes back a move.
func (eng *Engine) UndoMove(move Move) {
	eng.position.UndoMove(move)
	capt := move.Capture
	if capt != NoPiece {
		eng.put(capt.Color(), capt.Figure())
	}
	if move.MoveType == Promotion {
		eng.put(eng.position.ToMove, Pawn)
		eng.remove(eng.position.ToMove, move.Promotion.Figure())
	}
}

// countMaterial counts pieces and updates the eng.pieceScore
func (eng *Engine) countMaterial() {
	eng.pieceScore = 0
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			cnt := Popcnt(uint64(eng.position.ByPiece(col, fig)))
			eng.pieces[col][fig] = cnt
			eng.pieceScore += ColorWeight[col] * figureBonus[fig]
		}
	}
}

// Evaluate current position from white's POV.
// Figure values and bonuses are taken from:
// http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) Score() int {
	score := eng.pieceScore

	// Give bonus for connected bishops.
	if eng.pieces[White][Bishop] >= 2 {
		score += bishopPairBonus
	}
	if eng.pieces[Black][Bishop] >= 2 {
		score -= bishopPairBonus
	}

	// Give bonuses based on number of pawns.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		numPawns := eng.pieces[col][Pawn]
		if numPawns > 5 {
			adjust := knightPawnBonus * eng.pieces[col][Knight]
			adjust -= rookPawnPenalty * eng.pieces[col][Rook]
			score += adjust * ColorWeight[col] * (numPawns - 5)
		}
	}

	return score
}

func (eng *Engine) alphaBetaMax(alpha, beta int, depth int) (Move, int) {
	eng.nodes++
	if depth == 0 {
		return Move{}, eng.Score() + rand.Intn(11) - 5
	}

	bestMove := Move{}
	start := len(eng.moves)
	eng.moves = eng.position.GenerateMoves(eng.moves)
	for len(eng.moves) > start {
		// Pops last move.
		last := len(eng.moves) - 1
		move := eng.moves[last]
		eng.moves = eng.moves[:last]

		eng.DoMove(move)
		if !eng.position.IsChecked(White) {
			_, score := eng.alphaBetaMin(alpha, beta, depth-1)
			if score > knownWinScore {
				score--
			}
			if score >= beta {
				eng.UndoMove(move)
				eng.moves = eng.moves[:start]
				return Move{}, beta
			}
			if score > alpha {
				bestMove = move
				alpha = score
			}
			if bestMove.MoveType == NoMove {
				bestMove = move
			}
		}
		eng.UndoMove(move)
	}

	if bestMove.MoveType == NoMove {
		if eng.position.IsChecked(White) {
			return Move{}, -mateScore
		} else {
			return Move{}, 0
		}
	}

	return bestMove, alpha
}

func (eng *Engine) alphaBetaMin(alpha, beta int, depth int) (Move, int) {
	eng.nodes++
	if depth == 0 {
		return Move{}, eng.Score() + rand.Intn(11) - 5
	}

	bestMove := Move{}
	start := len(eng.moves)
	eng.moves = eng.position.GenerateMoves(eng.moves)
	for len(eng.moves) > start {
		// Pops last move.
		last := len(eng.moves) - 1
		move := eng.moves[last]
		eng.moves = eng.moves[:last]

		eng.DoMove(move)
		if !eng.position.IsChecked(Black) {
			_, score := eng.alphaBetaMax(alpha, beta, depth-1)
			if score < -knownWinScore {
				score++
			}
			if score <= alpha {
				eng.UndoMove(move)
				eng.moves = eng.moves[:start]
				return Move{}, alpha
			}
			if score < beta {
				bestMove = move
				beta = score
			}
			if bestMove.MoveType == NoMove {
				bestMove = move
			}
		}
		eng.UndoMove(move)
	}

	if bestMove.MoveType == NoMove {
		if eng.position.IsChecked(Black) {
			return Move{}, mateScore
		} else {
			return Move{}, 0
		}
	}

	return bestMove, beta
}

func (eng *Engine) alphaBeta(depth int) (Move, int) {
	if eng.position.ToMove == White {
		return eng.alphaBetaMax(math.MinInt32, math.MaxInt32, depth)
	} else {
		return eng.alphaBetaMin(math.MinInt32, math.MaxInt32, depth)
	}
}

// Play find the next move.
// tc should already be started.
func (eng *Engine) Play(tc TimeControl) (Move, error) {
	var move Move
	var score int

	start := time.Now()
	for depth := tc.NextDepth(); depth != 0; depth = tc.NextDepth() {
		move, score = eng.alphaBeta(depth)
		elapsed := time.Now().Sub(start)
		_, _ = score, elapsed
		/*
			fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d pv %v\n",
				depth, score, eng.nodes, elapsed/time.Millisecond,
				eng.nodes*uint64(time.Second)/uint64(elapsed+1),
				move)
		*/
	}

	if move.MoveType == NoMove {
		// If there is no valid move, then it's a stalement or a checkmate.
		if eng.position.IsChecked(eng.position.ToMove) {
			return Move{}, ErrorCheckMate
		} else {
			return Move{}, ErrorStaleMate
		}
	}

	return move, nil
}
