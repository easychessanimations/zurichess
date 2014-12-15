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

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")

	figureBonus = [FigureMaxValue]int{
		0,     // NoFigure
		100,   // Pawn
		325,   // Knight
		325,   // Bishop
		500,   // Rook
		975,   // Queen
		10000, // King
	}

	mateScore       = 200000
	bishopPairBonus = 50
	knightPawnBonus = 6
	rookPawnPenalty = 12
)

type TimeControl struct {
	// Remaining time.
	Time time.Duration
	// Time increment after each move.
	Inc time.Duration
	// Number of moves left. Recommended values
	// 0 when there is no time refresh.
	// 1 when solving puzzels.
	// n when there is a time refresh.
	MovesToGo int
}

type Engine struct {
	position *Position
	moves    []Move
	nodes    uint64
}

// NewEngine returns a new engine for pos.
func NewEngine(pos *Position) *Engine {
	return &Engine{
		position: pos,
		moves:    make([]Move, 0, 64),
	}
}

// ParseMove parses the move from a string.
func (eng *Engine) ParseMove(move string) Move {
	return eng.position.ParseMove(move)
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.position.DoMove(move)
}

// UndoMove takes back a move.
func (eng *Engine) UndoMove(move Move) {
	eng.position.UndoMove(move)
}

// Evaluate current position from white's POV.
// Figure values and bonuses are taken from:
// http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) Evaluate() int {
	pos := eng.position
	score := 0

	// Compute piece values.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		fb := figureBonus
		numPawns := Popcnt(uint64(pos.ByPiece(col, Pawn)))
		if numPawns > 5 {
			fb[Knight] += (numPawns - 5) * knightPawnBonus
			fb[Rook] -= (numPawns - 5) * rookPawnPenalty
		}

		colorScore := 0
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			bb := pos.ByPiece(col, fig)
			colorScore += Popcnt(uint64(bb)) * fb[fig]
		}
		score += colorScore * ColorWeight[col]
	}

	// Award bishop pair.
	{
		if Popcnt(uint64(pos.ByPiece(White, Bishop))) >= 2 {
			score += bishopPairBonus
		}
		if Popcnt(uint64(pos.ByPiece(Black, Bishop))) >= 2 {
			score -= bishopPairBonus
		}
	}

	return score
}

// Score returns a cached result of Evaluate.
func (eng *Engine) Score() int {
	return eng.Evaluate()
}

func (eng *Engine) minMax(depth int) (Move, int) {
	eng.nodes++
	if depth == 0 {
		return Move{}, eng.Evaluate()
	}

	toMove := eng.position.ToMove
	weight := ColorWeight[toMove]
	bestMove := Move{}
	bestScore := -weight * math.MaxInt32

	found := false
	start := len(eng.moves)
	eng.moves = eng.position.GenerateMoves(eng.moves)
	for len(eng.moves) > start {
		// Pops last move.
		last := len(eng.moves) - 1
		move := eng.moves[last]
		eng.moves = eng.moves[:last]

		eng.DoMove(move)
		if !eng.position.IsChecked(toMove) {
			found = true
			_, score := eng.minMax(depth - 1)
			score += rand.Intn(11) - 5
			if score*weight > bestScore*weight {
				bestScore = score
				bestMove = move
			}
		}
		eng.UndoMove(move)
	}

	// If there is no valid move, then it's a stalement or a checkmate.
	if !found {
		if eng.position.IsChecked(toMove) {
			return Move{}, -weight * mateScore
		} else {
			return Move{}, 0
		}
	}

	// log.Printf("at %d got %v %d", depth, bestMove, bestScore)
	return bestMove, bestScore
}

const (
	defaultMovesToGo = 30 // default number of more moves expected to play
	branchFactor     = 21 // default branching factor
)

func (eng *Engine) Play(tc TimeControl) (Move, error) {
	start := time.Now()

	// Compute how much time to think according to the formula below.
	// The formula allows engine to use more of time.Left in the begining
	// and rely more on the inc time later.
	movesToGo := time.Duration(defaultMovesToGo)
	if tc.MovesToGo != 0 {
		movesToGo = time.Duration(tc.MovesToGo)
	}
	thinkTime := (tc.Time + (movesToGo-1)*tc.Inc) / movesToGo
	timeLimit := thinkTime / branchFactor

	var move Move
	elapsed := time.Now().Sub(start)
	for depth := 3; depth < 64 && elapsed <= timeLimit; depth++ {
		var score int
		move, score = eng.minMax(depth)
		elapsed = time.Now().Sub(start)
		fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d pv %v\n",
			depth, score, eng.nodes, elapsed/time.Millisecond,
			eng.nodes*uint64(time.Second)/uint64(elapsed+1),
			move)
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
