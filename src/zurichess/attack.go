package main

import (
	"log"
)

var (
	knightJump = [8][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}

	BbKnightAttack [64]Bitboard
)

func initBbKnightAttack() {
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			bb := Bitboard(0)
			for _, d := range knightJump {
				r_, f_ := r+d[0], f+d[1]
				if 0 > r_ || r_ >= 8 || 0 > f_ || f_ >= 8 {
					continue
				}
				bb |= RankFile(r_, f_).Bitboard()
			}
			BbKnightAttack[sq] = bb
		}
	}
	log.Println("BbKnightAttack initialized")
}
