// puzzle tries to solve puzzles from files.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"bitbucket.org/brtzsnr/zurichess/engine"
	"bitbucket.org/brtzsnr/zurichess/notation"
)

var (
	input      = flag.String("input", "", "file with EPD lines")
	output     = flag.String("output", "", "file to write EPD with solutions")
	deadline   = flag.Duration("deadline", 0, "how much time to spend for each move")
	maxDepth   = flag.Int("max_depth", 0, "search up to max_depth plies")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	quiet      = flag.Bool("quiet", false, "don't print individual tests")

	mvvLva   = flag.String("mvv_lva", "", "set MvvLva table")
	maxNodes = flag.Uint64("max_nodes", 0, "maximum nodes to search")
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Validate falgs.
	flag.Parse()
	if *input == "" {
		log.Fatal("--input not specified")
	}
	if *mvvLva != "" {
		engine.SetMvvLva(*mvvLva)
	}
	if *cpuprofile != "" { // Enable cpuprofile.
		fin, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(fin)
		defer pprof.StopCPUProfile()
	}

	var err error
	var fin *os.File
	if fin, err = os.Open(*input); err != nil {
		log.Fatalf("cannot open %s for reading: %v", *input, err)
	}
	defer fin.Close()

	var fout *os.File
	if *output != "" {
		if fout, err = os.Create(*output); err != nil {
			log.Fatalf("cannot open %s for writing: %v", *output, err)
		}
		defer fout.Close()
	}

	stats := engine.Stats{}

	// Read file line by line.
	solvedTests, numTests := 0, 0
	buf := bufio.NewReader(fin)
	for i, o := 0, 0; ; i++ {
		// Builds time control.
		var timeControl engine.TimeControl
		if *deadline != 0 {
			timeControl = &engine.OnClockTimeControl{
				Time:      *deadline,
				Inc:       0,
				MovesToGo: 1,
			}
		} else if *maxDepth != 0 {
			timeControl = &engine.FixedDepthTimeControl{
				MinDepth: 1,
				MaxDepth: *maxDepth,
			}
		} else {
			log.Fatal("--deadline or --max_depth must be specified")
		}

		// Read EPD line.
		line, err := buf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}

		// Trim comments and spaces.
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Reads position from file.
		epd, err := notation.ParseEPD(line)
		if err != nil {
			log.Println("error:", err)
			log.Println("skipping", line)
			continue
		}

		// Evaluate position.
		timeControl.Start()
		ai := engine.NewEngine(nil, engine.Options{})
		ai.SetPosition(epd.Position)
		actual := ai.Play(timeControl)

		// Update number of solved games.
		numTests++
		var expected engine.Move
		for _, expected = range epd.BestMove {
			if expected == actual[0] {
				solvedTests++
				break
			}
		}

		if !*quiet {
			// Print a header from time to time.
			if o%25 == 0 {
				fmt.Println()
				fmt.Println("line     bm actual  cache  nodes  correct epd")
				fmt.Println("----+------+------+------+------+--------+---")
			}

			// Print results.
			fmt.Printf("%4d %6s %6s %5.2f%% %5dK %4d/%4d %s\n",
				i+1, expected.String(), actual[0].String(),
				float32(ai.Stats.CacheHit)/float32(ai.Stats.CacheHit+ai.Stats.CacheMiss)*100,
				ai.Stats.Nodes/1000, solvedTests, numTests, line)
			o++
		}

		if fout != nil {
			epd.BestMove = []engine.Move{actual[0]}
			fmt.Fprintln(fout, epd.String())
		}

		// Update stats.
		stats.CacheHit += ai.Stats.CacheHit
		stats.CacheMiss += ai.Stats.CacheMiss
		stats.Nodes += ai.Stats.Nodes
		if *maxNodes != 0 && stats.Nodes > *maxNodes {
			break
		}
	}

	fmt.Printf("%s solved %d out of %d ; nodes %d ; cachehit %d out of %d (%.2f%%) ;\n",
		*input, solvedTests, numTests, stats.Nodes,
		stats.CacheHit, stats.CacheHit+stats.CacheMiss,
		stats.CacheHitRatio()*100)
}
