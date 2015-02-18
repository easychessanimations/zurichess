// zobrist.go contains magic numbers used for Zobrist hashing.
//
// More information on Zobrist hashing can be found in the paper:
// http://research.cs.wisc.edu/techreports/1970/TR88.pdf

package engine

import (
	"math/rand"
)

var (
	// The zobrist* arrays contain magic numbers used for Zobrist hashing.
	zobristPiece     [PieceArraySize][SquareArraySize]uint64
	zobristEnpassant [SquareArraySize]uint64
	zobristCastle    [CastleArraySize]uint64
	zobristColor     [ColorArraySize]uint64
)

func rand64(r *rand.Rand) uint64 {
	return uint64(r.Int63())<<32 ^ uint64(r.Int63())
}

func initzobristPiece(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
				zobristPiece[ColorFigure(col, fig)][sq] = rand64(r)
			}
		}
	}
}

func initzobristEnpassant(r *rand.Rand) {
	for sq := SquareA3; sq <= SquareH3; sq++ {
		zobristEnpassant[sq] = rand64(r)
	}
	for sq := SquareA6; sq <= SquareH6; sq++ {
		zobristEnpassant[sq] = rand64(r)
	}
}

func initzobristCastle(r *rand.Rand) {
	for i := CastleMinValue; i < CastleMaxValue; i++ {
		zobristCastle[i] = rand64(r)
	}
}

func initzobristColor(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		zobristColor[col] = rand64(r)
	}
}

func init() {
	r := rand.New(rand.NewSource(1))
	initzobristPiece(r)
	initzobristEnpassant(r)
	initzobristCastle(r)
	initzobristColor(r)
}
