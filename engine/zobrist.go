// zobrist.go contains magic numbers used for Zobrist hashing.
//
// More information on Zobrist hashing can be found in the paper:
// http://research.cs.wisc.edu/techreports/1970/TR88.pdf

package engine

import (
	"math/rand"
)

var (
	// The Zobrist* arrays contain magic numbers used for Zobrist hashing.

	ZobristPiece     [PieceArraySize][SquareArraySize]uint64
	ZobristEnpassant [SquareArraySize]uint64
	ZobristCastle    [CastleArraySize]uint64
	ZobristColor     [ColorArraySize]uint64
)

func rand64(r *rand.Rand) uint64 {
	return uint64(r.Int63())<<32 ^ uint64(r.Int63())
}

func initZobristPiece(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
				ZobristPiece[ColorFigure(col, fig)][sq] = rand64(r)
			}
		}
	}
}

func initZobristEnpassant(r *rand.Rand) {
	for sq := SquareA3; sq <= SquareH3; sq++ {
		ZobristEnpassant[sq] = rand64(r)
	}
	for sq := SquareA6; sq <= SquareH6; sq++ {
		ZobristEnpassant[sq] = rand64(r)
	}
}

func initZobristCastle(r *rand.Rand) {
	for i := CastleMinValue; i < CastleMaxValue; i++ {
		ZobristCastle[i] = rand64(r)
	}
}

func initZobristColor(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		ZobristColor[col] = rand64(r)
	}
}

func init() {
	r := rand.New(rand.NewSource(1))
	initZobristPiece(r)
	initZobristEnpassant(r)
	initZobristCastle(r)
	initZobristColor(r)
}
