// zobrist.go contains magic number for zobrist hashing.
package engine

import (
	"math/rand"
)

var (
	ZobriestPiece     [PieceArraySize][SquareArraySize]uint64
	ZobriestEnpassant [SquareArraySize]uint64
	ZobriestCastle    [CastleArraySize]uint64
	ZobriestColor     [ColorArraySize]uint64
)

func rand64(r *rand.Rand) uint64 {
	return uint64(r.Int63())<<32 ^ uint64(r.Int63())
}

func initZobriestPiece(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		for fig := FigureMinValue; fig <= FigureMaxValue; fig++ {
			for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
				ZobriestPiece[ColorFigure(col, fig)][sq] = rand64(r)
			}
		}
	}
}

func initZobriestEnpassant(r *rand.Rand) {
	for sq := SquareA3; sq <= SquareH3; sq++ {
		ZobriestEnpassant[sq] = rand64(r)
	}
	for sq := SquareA6; sq <= SquareH6; sq++ {
		ZobriestEnpassant[sq] = rand64(r)
	}
}

func initZobriestCastle(r *rand.Rand) {
	for i := CastleMinValue; i < CastleMaxValue; i++ {
		ZobriestCastle[i] = rand64(r)
	}
}

func initZobriestColor(r *rand.Rand) {
	for col := ColorMinValue; col <= ColorMaxValue; col++ {
		ZobriestColor[col] = rand64(r)
	}
}

func init() {
	r := rand.New(rand.NewSource(1))
	initZobriestPiece(r)
	initZobriestEnpassant(r)
	initZobriestCastle(r)
	initZobriestColor(r)
}
