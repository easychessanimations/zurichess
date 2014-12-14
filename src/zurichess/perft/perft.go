// Perft test.
// https://chessprogramming.wikispaces.com/Perft
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"zurichess/engine"
)

var (
	fen        = flag.String("fen", "startpos", "position to search")
	min        = flag.Int("min", 1, "min depth to search (inclusive)")
	max        = flag.Int("max", 5, "max depth to search (inclusive)")
	depth      = flag.Int("depth", 0, "if non zero, searches only this depth")
	splitDepth = flag.Int("split", 0, "split depth")

	splitMoves []string
	perftMoves []engine.Move
)

type counters struct {
	nodes     uint64
	captures  uint64
	enpassant uint64
	castles   uint64
}

func (co *counters) Add(ot counters) {
	co.nodes += ot.nodes
	co.captures += ot.captures
	co.enpassant += ot.enpassant
	co.castles += ot.castles
}

func (co counters) Equals(ot counters) bool {
	return co.nodes == ot.nodes &&
		co.captures == ot.captures &&
		co.enpassant == ot.enpassant &&
		co.castles == ot.castles
}

var (
	startpos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	kiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -"
	duplain  = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -"

	known = map[string]string{
		"startpos": startpos,
		"kiwipete": kiwipete,
		"duplain":  duplain,
	}

	data = map[string][]counters{
		startpos: []counters{
			{1, 0, 0, 0},
			{20, 0, 0, 0},
			{400, 0, 0, 0},
			{8902, 34, 0, 0},
			{197281, 1576, 0, 0},
			{4865609, 82719, 258, 0},
			{119060324, 2812008, 5248, 0},
		},
		kiwipete: []counters{
			{1, 0, 0, 0},
			{48, 8, 0, 2},
			{2039, 351, 1, 91},
			{97862, 17102, 45, 3162},
			{4085603, 757163, 1929, 128013},
			{193690690, 35043416, 73365, 4993637},
		},
		duplain: []counters{
			{1, 0, 0, 0},
			{14, 1, 0, 0},
			{191, 14, 0, 0},
			{2812, 209, 2, 0},
			{43238, 3348, 123, 0},
			{674624, 52051, 1165, 0},
			{11030083, 940350, 33325, 0},
			{178633661, 14519036, 294874, 0},
		},
	}
)

func perft(pos *engine.Position, depth int, moves *[]engine.Move) counters {
	if depth == 0 {
		return counters{1, 0, 0, 0}
	}

	r := counters{}
	start := len(*moves)
	*moves = pos.GenerateMoves(*moves)

	for len(*moves) > start {
		last := len(*moves) - 1
		move := (*moves)[last]
		*moves = (*moves)[:last]

		pos.DoMove(move)
		if pos.IsChecked(pos.ToMove.Other()) {
			pos.UndoMove(move)
			continue
		}

		if depth == 1 { // count only leaf nodes
			if move.Capture != engine.NoPiece {
				r.captures++
			}
			if move.MoveType == engine.Enpassant {
				r.enpassant++
			}
			if move.MoveType == engine.Castling {
				r.castles++
			}
		}

		r.Add(perft(pos, depth-1, moves))
		pos.UndoMove(move)
	}
	return r
}

func split(pos *engine.Position, depth, splitDepth int) counters {
	r := counters{}
	if depth == 0 || splitDepth == 0 {
		r = perft(pos, depth, new([]engine.Move))
	} else {
		moves := pos.GenerateMoves(nil)
		for _, move := range moves {
			pos.DoMove(move)
			if !pos.IsChecked(pos.ToMove.Other()) {
				splitMoves = append(splitMoves, move.String())
				r.Add(split(pos, depth-1, splitDepth-1))
				splitMoves = splitMoves[:len(splitMoves)-1]
			}
			pos.UndoMove(move)
		}
	}

	if len(splitMoves) != 0 {
		fmt.Printf("   %2d %12d %8d %9d %7d split %s\n",
			depth, r.nodes, r.captures, r.enpassant, r.castles, strings.Join(splitMoves, " "))
	}
	return r
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	var expected []counters
	if s, has := known[*fen]; has {
		*fen = s
		expected = data[*fen]
	}
	if *depth != 0 {
		*min = *depth
		*max = *depth
	}

	fmt.Printf("Searching FEN \"%s\"\n", *fen)
	pos, err := engine.PositionFromFEN(*fen)
	if err != nil {
		log.Fatalln("Cannot parse --fen:", err)
	}

	fmt.Printf("depth        nodes   captures enpassant castles eval   KNps elapsed\n")
	fmt.Printf("-----+------------+----------+---------+-------+----+------+-------\n")

	for d := *min; d <= *max; d++ {
		start := time.Now()
		c := split(pos, d, *splitDepth)
		duration := time.Since(start)

		ok := ""
		if d < len(expected) {
			if c.Equals(expected[d]) {
				ok = "good"
			} else {
				ok = "bad"
			}
		}

		fmt.Printf("   %2d %12d %10d %9d %7d %4s %6.f %v\n",
			d, c.nodes, c.captures, c.enpassant, c.castles, ok,
			float64(c.nodes)/duration.Seconds()/1e3, duration)

		if ok == "bad" {
			e := expected[d]
			fmt.Printf("   %2d %12d %8d %9d %7d %v\n",
				d, e.nodes, e.captures, e.enpassant, e.castles, "expected")
			break
		}
	}
}
