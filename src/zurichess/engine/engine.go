package engine

import (
	"errors"
	"log"
	"math/rand"
)

var _ = log.Println

type Engine struct {
	Position *Position
}

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")

	figureScore = []int{
		0,     // NoFigure
		100,   // Pawn
		300,   // Knight
		315,   // Bishop
		500,   // Rook
		900,   // Queen
		10000, // King
	}

	checkScore = -50
	mateScore  = 200000
)

// evaluate evaluates the score of a position from current color POV.
// Simplest implementation adapted from:
// https://chessprogramming.wikispaces.com/Evaluation
func (eng *Engine) evaluate() int {
	score := 0
	toMove := eng.Position.ToMove

	if eng.Position.IsChecked(toMove) {
		score += checkScore
	}
	if eng.Position.IsChecked(toMove.Other()) {
		score -= checkScore
	}

	// Compute piece values.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		colorScore := 0
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			bb := eng.Position.ByColor[col] & eng.Position.ByFigure[fig]
			colorScore += int(bb.Popcnt()) * figureScore[fig]
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
	bestScore := -mateScore

	found := false
	moves := eng.Position.GenerateMoves(nil)
	for _, move := range moves {
		eng.Position.DoMove(move)
		if !eng.Position.IsChecked(toMove) {
			found = true
			_, score := eng.minMax(depth - 1)
			score *= weight
			if (score == bestScore && rand.Intn(2) == 0) || score > bestScore {
				bestScore = score
				bestMove = move
			}
		}
		eng.Position.UndoMove(move)
	}

	// If there is no valid move, then it's a stalement or a checkmate.
	if !found {
		if eng.Position.IsChecked(toMove) {
			return Move{}, -mateScore
		} else {
			return Move{}, 0
		}
	}

	// log.Printf("at %d got %v %d", depth, bestMove, bestScore)
	return bestMove, bestScore * weight
}

func (eng *Engine) Play() (Move, error) {
	move, _ := eng.minMax(3)
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
