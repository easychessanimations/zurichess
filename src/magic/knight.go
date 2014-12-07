package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	output = flag.String("o", "/dev/stdout", "output file")

	knightJump = [8][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
)

func main() {
	flag.Parse()

	out, err := os.Create(*output)
	if err != nil {
		log.Fatalf("failed to open %s: %v", *output, err)
	}
	defer out.Close()

	fmt.Fprintf(out, "package main\n")
	fmt.Fprintf(out, "\n")

	fmt.Fprintf(out, "var knightAttack = [64]Bitboard{\n")
	for r := 0; r < 8; r++ {
		fmt.Fprintf(out, "\t")
		for f := 0; f < 8; f++ {
			bb := uint64(0)
			for _, d := range knightJump {
				r_, f_ := r+d[0], f+d[1]
				if 0 > r_ || r_ >= 8 || 0 > f_ || f_ >= 8 {
					continue
				}
				bb |= 1 << uint(r_*8+f_)
			}
			fmt.Fprintf(out, "0x%016x, ", bb)
		}
		fmt.Fprintf(out, "\n")
	}
	fmt.Fprintf(out, "}\n")
}
