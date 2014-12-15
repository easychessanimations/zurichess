package engine

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"time"
)

var _ = log.Println

type Engine struct {
	Position *Position
	moves    []Move
}

var (
	ErrorCheckMate = errors.New("current position is checkmate")
	ErrorStaleMate = errors.New("current position is stalemate")

	figureBonus = [8]int{
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

// Figure values and bonuses are taken from:
// http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) evaluate() int {
	pos := eng.Position
	score := 0

	// Compute piece values.
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		fb := figureBonus
		numPawns := Popcnt(uint64(pos.ByPiece(White, Pawn)))
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
	minDuration := 250 * time.Millisecond
	start := time.Now()

	var move Move
	depth := 3
	for depth < 32 && time.Now().Sub(start) < minDuration {
		depth++
		move, _ = eng.minMax(depth)
	}

	log.Println("reached depth", depth, "in", time.Now().Sub(start))
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
