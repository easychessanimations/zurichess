package engine

import (
	"errors"
	"log"
	"math"
	"math/rand"
)

var _ = log.Println

type Engine struct {
	Position *Position
	moves    []Move
}

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")

	figureBonus = []int{
		0,     // NoFigure
		100,   // Pawn
		300,   // Knight
		350,   // Bishop
		500,   // Rook
		900,   // Queen
		10000, // King
	}

	mateScore = 200000
)

// evaluate evaluates the score of a position from white's color POV.
// Simplest implementation adapted from:
// https://chessprogramming.wikispaces.com/Evaluation
func (eng *Engine) evaluate() int {
	pos := eng.Position
	score := 0

	// Compute piece values.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		colorScore := 0
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			bb := pos.ByColor[col] & pos.ByFigure[fig]
			colorScore += Popcnt(uint64(bb)) * figureBonus[fig]
		}
		score += colorScore * ColorWeight[col]
	}

	return score
}

func (eng *Engine) minMax(depth int) (Move, int) {
	if depth == 0 {
		return Move{}, eng.evaluate()
	}

	toMove := eng.Position.ToMove
	weight := ColorWeight[toMove]
	bestMove := Move{}
	bestScore := -weight * math.MaxInt32

	found := false
	start := len(eng.moves)
	eng.moves = eng.Position.GenerateMoves(eng.moves)
	for len(eng.moves) > start {
		// Pops last move.
		last := len(eng.moves) - 1
		move := eng.moves[last]
		eng.moves = eng.moves[:last]

		eng.Position.DoMove(move)
		if !eng.Position.IsChecked(toMove) {
			found = true
			_, score := eng.minMax(depth - 1)
			score += rand.Intn(11) - 5
			if score*weight > bestScore*weight {
				bestScore = score
				bestMove = move
			}
		}
		eng.Position.UndoMove(move)
	}

	// If there is no valid move, then it's a stalement or a checkmate.
	if !found {
		if eng.Position.IsChecked(toMove) {
			return Move{}, -weight * mateScore
		} else {
			return Move{}, 0
		}
	}

	// log.Printf("at %d got %v %d", depth, bestMove, bestScore)
	return bestMove, bestScore
}

func (eng *Engine) Play() (Move, error) {
	move, _ := eng.minMax(4)
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
