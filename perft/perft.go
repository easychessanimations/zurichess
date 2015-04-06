// Perft is a perft tool.
//
// Perft's main purpose is to test, debug and benchmark move generation.
// To do this we count number of nodes, captures, en passant, castles and
// promotions for given depths (usually small 4-7) from specific position.
// In order to aid debugging debugging pertf can do split up to any level.
//
// For more results and test description see:
//      https://chessprogramming.wikispaces.com/Perft
//      https://chessprogramming.wikispaces.com/Perft+Results
//      http://www.10x8.net/chess/PerfT.html
//
// Examples:
//
// Simple fast integration test:
//      $ go test bitbucket.org/brtzsnr/zurichess/perft
//
// startpos:
//	$ ./perft --fen startpos --max_depth 7
//	Searching FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
//	depth        nodes   captures enpassant castles   promotions eval  KNps   elapsed
//	-----+------------+----------+---------+---------+----------+-----+------+-------
//	    1           20          0         0         0          0 good    154 129.948µs
//	    2          400          0         0         0          0 good    158 2.531444ms
//	    3         8902         34         0         0          0 good    266 33.494604ms
//	    4       197281       1576         0         0          0 good   3454 57.114844ms
//	    5      4865609      82719       258         0          0 good  12141 400.762477ms
//	    6    119060324    2812008      5248         0          0 good  24027 4.955285846s
//	    7   3195901860  108329926    319617    883453          0 good  40040 1m19.817376124s
//
// kiwipete:
//	$ ./perft --fen kiwipete --max_depth 6
//	Searching FEN "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
//	depth        nodes   captures enpassant castles   promotions eval  KNps   elapsed
//	-----+------------+----------+---------+---------+----------+-----+------+-------
//	    1           48          8         0         2          0 good    311 154.233µs
//	    2         2039        351         1        91          0 good    267 7.648547ms
//	    3        97862      17102        45      3162          0 good   1259 77.709405ms
//	    4      4085603     757163      1929    128013      15172 good  11099 368.115794ms
//	    5    193690690   35043416     73365   4993637       8392 good  20369 9.508896954s
//	    6   8031647685 1558445089   3577504 184513607   56627920 good  27478 4m52.289206499s
//
// duplain:
//	$ ./perft --fen duplain --max_depth 7
//	Searching FEN "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"
//	depth        nodes   captures enpassant castles   promotions eval  KNps   elapsed
//	-----+------------+----------+---------+---------+----------+-----+------+-------
//	    1           14          1         0         0          0 good    100 140.46µs
//	    2          191         14         0         0          0 good    115 1.662379ms
//	    3         2812        209         2         0          0 good    137 20.457747ms
//	    4        43238       3348       123         0          0 good    937 46.124066ms
//	    5       674624      52051      1165         0          0 good  14036 48.065305ms
//	    6     11030083     940350     33325         0       7552 good  28502 386.996491ms
//	    7    178633661   14519036    294874         0     140024 good  72956 2.448514893s
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"bitbucket.org/brtzsnr/zurichess/engine"
)

var (
	fen        = flag.String("fen", "startpos", "position to search")
	minDepth   = flag.Int("min_depth", 1, "minimum depth to search (inclusive)")
	maxDepth   = flag.Int("max_depth", 5, "maximum depth to search (inclusive)")
	depth      = flag.Int("depth", 0, "if non zero, searches only this depth")
	splitDepth = flag.Int("split", 0, "split depth")

	splitMoves []string
	perftMoves []engine.Move
)

// counters counts leafs after backtracking on a position up to certain depth.
type counters struct {
	nodes      uint64
	captures   uint64
	enpassant  uint64
	castles    uint64
	promotions uint64
}

// Add adds ot to co.
func (co *counters) Add(ot counters) {
	co.nodes += ot.nodes
	co.captures += ot.captures
	co.enpassant += ot.enpassant
	co.castles += ot.castles
	co.promotions += ot.promotions
}

type hashEntry struct {
	zobriest uint64
	counters counters
	depth    int
}

var (
	startpos = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	kiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
	duplain  = "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"

	known = map[string]string{
		"startpos": startpos,
		"kiwipete": kiwipete,
		"duplain":  duplain,
	}

	data = map[string][]counters{
		startpos: []counters{
			{1, 0, 0, 0, 0},
			{20, 0, 0, 0, 0},
			{400, 0, 0, 0, 0},
			{8902, 34, 0, 0, 0},
			{197281, 1576, 0, 0, 0},
			{4865609, 82719, 258, 0, 0},
			{119060324, 2812008, 5248, 0, 0},
			{3195901860, 108329926, 319617, 883453, 0},
			{84998978956, 3523740106, 7187977, 23605205, 0},
		},
		kiwipete: []counters{
			{1, 0, 0, 0, 0},
			{48, 8, 0, 2, 0},
			{2039, 351, 1, 91, 0},
			{97862, 17102, 45, 3162, 0},
			{4085603, 757163, 1929, 128013, 15172},
			{193690690, 35043416, 73365, 4993637, 8392},
			{8031647685, 1558445089, 3577504, 184513607, 56627920},
		},
		duplain: []counters{
			{1, 0, 0, 0, 0},
			{14, 1, 0, 0, 0},
			{191, 14, 0, 0, 0},
			{2812, 209, 2, 0, 0},
			{43238, 3348, 123, 0, 0},
			{674624, 52051, 1165, 0, 0},
			{11030083, 940350, 33325, 0, 7552},
			{178633661, 14519036, 294874, 0, 140024},
		},
	}

	// A hash table to speedup perft.
	hashSize  = 1 << 20
	hashTable = make([]hashEntry, hashSize)
)

func perft(pos *engine.Position, depth int, hashTable []hashEntry, moves *[]engine.Move) counters {
	if depth == 0 {
		return counters{1, 0, 0, 0, 0}
	}

	if hashTable != nil {
		index := pos.Zobrist() % uint64(len(hashTable))
		if hashTable[index].depth == depth && hashTable[index].zobriest == pos.Zobrist() {
			return hashTable[index].counters
		}
	}

	r := counters{}
	start := len(*moves)
	pos.GenerateMoves(moves)
	for start < len(*moves) {
		last := len(*moves) - 1
		move := (*moves)[last]
		*moves = (*moves)[:last]

		pos.DoMove(move)
		if pos.IsChecked(pos.SideToMove.Opposite()) {
			pos.UndoMove(move)
			continue
		}

		if depth == 1 { // count only leaf nodes
			if move.Capture() != engine.NoPiece {
				r.captures++
			}
			if move.MoveType == engine.Enpassant {
				r.enpassant++
			}
			if move.MoveType == engine.Castling {
				r.castles++
			}
			if move.MoveType == engine.Promotion {
				r.promotions++
			}
		}

		r.Add(perft(pos, depth-1, hashTable, moves))
		pos.UndoMove(move)
	}

	if hashTable != nil {
		index := pos.Zobrist() % uint64(len(hashTable))
		hashTable[index] = hashEntry{
			zobriest: pos.Zobrist(),
			counters: r,
			depth:    depth,
		}
	}
	return r
}

func split(pos *engine.Position, depth, splitDepth int) counters {
	r := counters{}
	if depth == 0 || splitDepth == 0 {
		moves := make([]engine.Move, 0, 256)
		r = perft(pos, depth, hashTable, &moves)
	} else {
		var moves []engine.Move
		pos.GenerateMoves(&moves)
		for _, move := range moves {
			pos.DoMove(move)
			if !pos.IsChecked(pos.SideToMove.Opposite()) {
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
		*minDepth = *depth
		*maxDepth = *depth
	}

	fmt.Printf("Searching FEN \"%s\"\n", *fen)
	pos, err := engine.PositionFromFEN(*fen)
	if err != nil {
		log.Fatalln("Cannot parse --fen:", err)
	}

	fmt.Printf("depth        nodes   captures enpassant castles   promotions eval  KNps   elapsed\n")
	fmt.Printf("-----+------------+----------+---------+---------+----------+-----+------+-------\n")

	for d := *minDepth; d <= *maxDepth; d++ {
		start := time.Now()
		c := split(pos, d, *splitDepth)
		duration := time.Since(start)

		ok := ""
		if d < len(expected) {
			if c == expected[d] {
				ok = "good"
			} else {
				ok = "bad"
			}
		}

		fmt.Printf("   %2d %12d %10d %9d %9d %10d %-4s %6.f %v\n",
			d, c.nodes, c.captures, c.enpassant, c.castles, c.promotions,
			ok, float64(c.nodes)/duration.Seconds()/1e3, duration)

		if ok == "bad" {
			e := expected[d]
			fmt.Printf("   %2d %12d %10d %9d %9d %10d %s\n",
				d, e.nodes, e.captures, e.enpassant, e.castles, e.promotions,
				"expected")
			break
		}
	}
}
