package engine

import (
	"errors"
	"fmt"
	"log"
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
		350,   // Knight
		350,   // Bishop
		525,   // Rook
		1000,  // Queen
		10000, // King
	}

	knownWinScore   = 20000
	mateScore       = 30000
	infinityScore   = 32000
	bishopPairBonus = 50
	knightPawnBonus = 6
	rookPawnPenalty = 12
)

type Engine struct {
	AnalyseMode bool

	position *Position // current position
	moves    []Move    // moves stack
	nodes    uint64    // number of nodes evaluated

	pieces        [ColorMaxValue][FigureMaxValue]int
	pieceScore    int
	positionScore int
}

// NewEngine returns a new engine for pos.
func NewEngine(pos *Position) *Engine {
	eng := &Engine{
		position: pos,
		moves:    make([]Move, 0, 128),
	}
	eng.countMaterial()
	return eng
}

// ParseMove parses the move from a string.
func (eng *Engine) ParseMove(move string) Move {
	return eng.position.ParseMove(move)
}

// put adjusts score after puting piece on sq.
// mask is which side is to move.
// delta is -1 if the piece is taken (including undo), 1 otherwise.
func (eng *Engine) put(sq Square, piece Piece, delta int) {
	col := piece.Color()
	fig := piece.Figure()
	weight := ColorWeight[col]
	mask := ColorMask[col]

	eng.pieces[col][NoFigure] += delta
	eng.pieces[col][fig] += delta
	eng.pieceScore += delta * weight * figureBonus[fig]
	eng.positionScore += delta * weight * PieceSquareTable[fig][mask^sq]
}

// adjust updates score after making a move.
// delta is -1 if the move is taken back, 1 otherwise.
// position.ToMove must have not been updated already.
// TODO: enpassant.
func (eng *Engine) adjust(move Move, delta int) {
	color := eng.position.ToMove

	if move.MoveType == Promotion {
		eng.put(move.From, ColorFigure(color, Pawn), -delta)
	} else {
		eng.put(move.From, move.Target, -delta)
	}
	eng.put(move.To, move.Target, delta)

	if move.MoveType == Castling {
		rookStart := RookStartSquare(move.To)
		rookEnd := RookEndSquare(move.To)
		rook := CastlingRook(move.To)
		eng.put(rookStart, rook, -delta)
		eng.put(rookEnd, rook, delta)
	}

	if move.Capture != NoPiece {
		eng.put(move.To, move.Capture, -delta)
	}
}

// DoMove executes a move.
func (eng *Engine) DoMove(move Move) {
	eng.adjust(move, 1)
	eng.position.DoMove(move)
}

// UndoMove undoes a move. Must be the last move.
func (eng *Engine) UndoMove(move Move) {
	eng.position.UndoMove(move)
	eng.adjust(move, -1)
}

// countMaterial counts pieces and updates the eng.pieceScore
func (eng *Engine) countMaterial() {
	eng.pieceScore = 0
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			bb := eng.position.ByPiece(col, fig)
			for bb > 0 {
				eng.put(bb.Pop(), ColorFigure(col, fig), 1)
			}
		}
	}
}

// Evaluate current position from white's POV.
// Figure values and bonuses are taken from:
// http://home.comcast.net/~danheisman/Articles/evaluation_of_material_imbalance.htm
func (eng *Engine) Score() int {
	score := eng.pieceScore + eng.positionScore

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

// EndPosition determines whether this is and end game
// position based on the number of pieces.
// Returns score and a bool if the game has ended.
func (eng *Engine) EndPosition() (int, bool) {
	if eng.pieces[White][King] == 0 {
		return -mateScore, true
	}
	if eng.pieces[Black][King] == 0 {
		return +mateScore, true
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

	return eng.Score(), false
}

// negamax implements negamax framework with fail-soft.
// http://chessprogramming.wikispaces.com/Alpha-Beta#Implementation-Negamax%20Framework
func (eng *Engine) negamax(alpha, beta int, color Color, depth int) (Move, int) {
	eng.nodes++
	if score, done := eng.EndPosition(); done {
		return Move{}, ColorWeight[color] * score
	}
	if depth == 0 {
		return Move{}, ColorWeight[color] * eng.Score()
	}

	bestMove, bestScore := Move{}, -infinityScore
	start := len(eng.moves)
	moveGen := NewMoveGenerator(eng.position)

	for piece := WhitePawn; piece != NoPiece; {
		if len(eng.moves) == start {
			piece, eng.moves = moveGen.Next(eng.moves)
			continue
		}

		// Pop & try last move.
		last := len(eng.moves) - 1
		move := eng.moves[last]
		eng.moves = eng.moves[:last]

		eng.DoMove(move)
		if !eng.position.IsChecked(color) {
			_, score := eng.negamax(-beta, -alpha, color.Other(), depth-1)
			score = -score
			if score > knownWinScore {
				score--
			}
			if score >= beta {
				eng.UndoMove(move)
				eng.moves = eng.moves[:start]
				return Move{}, beta
			}
			if score > bestScore {
				bestMove, bestScore = move, score
				if score > alpha {
					alpha = score
				}
			}
		}
		eng.UndoMove(move)
	}

	if bestMove.MoveType == NoMove {
		if eng.position.IsChecked(color) {
			bestMove, bestScore = Move{}, -mateScore
		} else {
			bestMove, bestScore = Move{}, 0
		}
	}

	return bestMove, bestScore
}

func (eng *Engine) alphaBeta(depth int) (Move, int) {
	move, score := eng.negamax(-infinityScore, +infinityScore, eng.position.ToMove, depth)

	score *= ColorWeight[eng.position.ToMove]
	return move, score
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
		if eng.AnalyseMode {
			fmt.Printf("info depth %d score cp %d nodes %d time %d nps %d pv %v\n",
				depth, score, eng.nodes, elapsed/time.Millisecond,
				eng.nodes*uint64(time.Second)/uint64(elapsed+1),
				move)
		}
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
