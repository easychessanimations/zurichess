// zobrist.go contains magic number for zobrist hashing.
package engine

import (
	"math/rand"
)

var (
	ZobriestPiece     [PieceMaxValue][SquareMaxValue]uint64
	ZobriestEnpassant [SquareMaxValue]uint64
	ZobriestCastle    [CastleMaxValue]uint64
	ZobriestColor     [ColorMaxValue]uint64
)

func rand64(r *rand.Rand) uint64 {
	return uint64(r.Int63())<<32 ^ uint64(r.Int63())
}

func initZobriestPiece() {
	r := rand.New(rand.NewSource(1))
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		for fig := FigureMinValue; fig < FigureMaxValue; fig++ {
			for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
				ZobriestPiece[ColorFigure(col, fig)][sq] = rand64(r)
			}
		}
	}
}

func initZobriestEnpassant() {
	r := rand.New(rand.NewSource(2))
	for sq := SquareA3; sq <= SquareH3; sq++ {
		ZobriestEnpassant[sq] = rand64(r)
	}
	for sq := SquareA6; sq <= SquareH6; sq++ {
		ZobriestEnpassant[sq] = rand64(r)
	}
}

func initZobriestCastle() {
	r := rand.New(rand.NewSource(3))
	for i := CastleMinValue; i < CastleMaxValue; i++ {
		ZobriestCastle[i] = rand64(r)
	}
}

func initZobriestColor() {
	r := rand.New(rand.NewSource(4))
	for col := ColorMinValue; col < ColorMaxValue; col++ {
		ZobriestColor[col] = rand64(r)
	}
}

func init() {
	initZobriestPiece()
	initZobriestEnpassant()
	initZobriestCastle()
	initZobriestColor()
}
