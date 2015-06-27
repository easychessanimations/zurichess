// Tool bench benchmarks zurichess.
//
// The benchmark runs the engine on several games and outputs
// the number of nodes and the number of nodes per second.
// The test tests that the number of nodes stays constant
// for non-functional changes.
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
	// Several games downloaded from http://www.chessgames.com/.
	games = []gameInfo{
		{
			"Garry Kasparov - Veselin Topalov Hoogovens A Tournament Wijk aan Zee NED 1999.01.20",
			strings.Fields("e2e4 d7d6 d2d4 g8f6 b1c3 g7g6 c1e3 f8g7 d1d2 c7c6 f2f3 b7b5 g1e2 b8d7 e3h6 g7h6 d2h6 c8b7 a2a3 e7e5 e1c1 d8e7 c1b1 a7a6 e2c1 e8c8 c1b3 e5d4 d1d4 c6c5 d4d1 d7b6 g2g3 c8b8 b3a5 b7a8 f1h3 d6d5 h6f4 b8a7 h1e1 d5d4 c3d5 b6d5 e4d5 e7d6 d1d4 c5d4 e1e7 a7b6 f4d4 b6a5 b2b4 a5a4 d4c3 d6d5 e7a7 a8b7 a7b7 d5c4 c3f6 a4a3 f6a6 a3b4 c2c3 b4c3 a6a1 c3d2 a1b2 d2d1 h3f1 d8d2 b7d7 d2d7 f1c4 b5c4 b2h8 d7d3 h8a8 c4c3 a8a4 d1e1 f3f4 f7f5 b1c1 d3d2 a4a7"),
		},
		{
			"Vladimir Kramnik - Alexey Shirov Linares Linares, ESP 1994.??.??",
			strings.Fields("g1f3 d7d5 d2d4 c8f5 c2c4 e7e6 b1c3 c7c6 d1b3 d8b6 c4c5 b6c7 c1f4 c7c8 e2e3 g8f6 b3a4 b8d7 b2b4 a7a6 h2h3 f8e7 a4b3 e8g8 f1e2 f5e4 e1g1 e4f3 e2f3 e7d8 a2a4 d8c7 f4g5 h7h6 g5f6 d7f6 b4b5 e6e5 b5b6 c7b8 a4a5 e5d4 e3d4 b8f4 b3c2 c8d7 g2g3 d7h3 f3g2 h3h5 g3f4 f6g4 f1d1 a8e8 d1d3 h5h2 g1f1 f7f5 c2d2 f8f6 f2f3 e8e4 c3d5 c6d5 c5c6 e4f4 c6b7 f4e4 a1c1 g8h7 b7b8Q h2b8 f3g4 b8h2 d3f3 e4g4 b6b7 f6g6 c1c2 g4g2 d2g2 g6g2 c2g2 h2h1 f1f2 h1b1"),
		},
		{
			"Mikhail Tal - Boris Spassky Leningrad tt Leningrad tt 1954.??.??",
			strings.Fields("c2c4 g8f6 b1c3 e7e6 d2d4 c7c5 d4d5 e6d5 c4d5 g7g6 g1f3 f8g7 c1f4 d7d6 h2h3 e8g8 e2e3 f6e8 f1e2 b8d7 e1g1 d7e5 f4e5 d6e5 f3d2 f7f5 d1b3 e8d6 d2c4 e5e4 c3b5 d6b5 b3b5 b7b6 d5d6 c8d7 b5b3 b6b5 c4b6 c5c4 e2c4 b5c4 b3c4 f8f7 b6a8 d8a8 c4b3 g7e5 a1c1 g8g7 f1d1 a7a5 c1c7 a8e8 b3d5 a5a4 b2b4 a4b3 a2b3 e5f6 c7b7 e8e5 d5c4 f5f4 e3f4 e5f4 g2g3 f4f3 c4d5 f6c3 d1f1 g7h6 d5d1 f3f6 d1e2 c3d4 b7b4 e4e3 e2d3 f6f2 f1f2 e3f2 g1h2 f2f1N d3f1 f7f1 b4d4 f1f2 h2g1 f2f3"),
		},
	}

	depth = flag.Int("depth", 5, "depth to search to")
)

type gameInfo struct {
	description string   // description of the game
	moves       []string // moves played
}

// eval returns the number of nodes needed to go through all moves
func (g *gameInfo) eval(depth int) uint64 {
	engine.GlobalHashTable = engine.NewHashTable(2)
	pos, _ := engine.PositionFromFEN(engine.FENStartPos)
	eng := engine.NewEngine(pos, engine.Options{})

	var nodes uint64
	for _, mstr := range g.moves {
		tc := engine.NewFixedDepthTimeControl(pos, depth)
		tc.Start(false)
		eng.Play(tc)
		nodes += eng.Stats.Nodes
		eng.DoMove(pos.UCIToMove(mstr))
	}
	return nodes
}

// evalAll evaluates all games playing each position up to ply depth.
func evalAll(depth int) (uint64, float64) {
	start := time.Now()
	var nodes uint64
	for i := range games {
		n := games[i].eval(depth)
		nodes += n
		log.Printf("#%d %d %s\n", i, n, games[i].description)
	}
	elapsed := time.Now().Sub(start)
	return nodes, float64(nodes) / elapsed.Seconds()
}

func main() {
	flag.Parse()
	nodes, nps := evalAll(*depth)
	fmt.Printf("nodes %d\n", nodes)
	fmt.Printf("  nps %.0f\n", nps)
}
