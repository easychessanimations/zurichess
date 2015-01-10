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
	"time"

	"zurichess/engine"
)

var (
	input      = flag.String("input", "", "file with EPD lines")
	deadline   = flag.Duration("deadline", 10*time.Second, "how much time to spend for each move")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	quiet      = flag.Bool("quiet", false, "don't print individual tests")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	// Enable cpuprofile.
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Opens input.
	if *input == "" {
		log.Fatal("--input not specified")
	}
	f, err := os.Open(*input)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stats := engine.EngineStats{}

	// Read file line by line.
	solvedTests, numTests := 0, 0
	buf := bufio.NewReader(f)
	for i, o := 0, 0; ; i++ {
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
		epd, err := engine.ParseEPD(line)
		if err != nil {
			log.Println("error:", err)
			log.Println("skipping", line)
			continue
		}

		if len(epd.BestMove) == 0 {
			// Ignore positions without best move.
			// TODO: Handle AvoidMove, etc.
			continue
		}

		timeControl := &engine.OnClockTimeControl{
			Time:      *deadline,
			Inc:       0,
			MovesToGo: 1,
		}
		timeControl.Start()

		// Evaluate position.
		ai := engine.NewEngine(nil, engine.EngineOptions{})
		ai.SetPosition(epd.Position)
		actual, _ := ai.Play(timeControl)

		// Update stats.
		stats.CacheHit += ai.Stats.CacheHit
		stats.CacheMiss += ai.Stats.CacheMiss
		stats.Nodes += ai.Stats.Nodes

		// Update number of solved games.
		numTests++
		var expected engine.Move
		for _, expected := range epd.BestMove {
			if expected == actual {
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
				i+1, expected.String(), actual.String(),
				float32(ai.Stats.CacheHit)/float32(ai.Stats.CacheHit+ai.Stats.CacheMiss)*100,
				ai.Stats.Nodes/1000, solvedTests, numTests, line)
			o++
		}
	}

	fmt.Printf("%s solved %d out of %d ; nodes %d ; cachehit %d out of %d (%.2f%%) ;\n",
		*input, solvedTests, numTests, stats.Nodes,
		stats.CacheHit, stats.CacheHit+stats.CacheMiss,
		stats.CacheHitRatio()*100)
}
