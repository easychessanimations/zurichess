package mates

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"bitbucket.org/brtzsnr/zurichess/engine"
	"bitbucket.org/brtzsnr/zurichess/notation"
)

func helper(t *testing.T, path string, depth, failures int) {
	fin, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot open %s for reading: %v", path, err)
	}
	defer fin.Close()

	failed, total := 0, 0
	buf := bufio.NewReader(fin)
	for {
		// Read EPD line.
		line, err := buf.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
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
			t.Fatal(err)
			continue
		}

		// Starts engine to play up to depth.
		tc := engine.NewFixedDepthTimeControl(epd.Position, depth)
		tc.Start(false)
		eng := engine.NewEngine(nil, engine.Options{})
		eng.SetPosition(epd.Position)
		pv := eng.Play(tc)

		// Check returned move.
		solved := false
		for _, expected := range epd.BestMove {
			if len(pv) > 0 && expected == pv[0] {
				solved = true
				break
			}
		}

		total++
		if !solved {
			failed++
			t.Logf("failed %s", epd.Position)
			t.Logf("expected one of %v, got pv %v", epd.BestMove, pv)
		}
	}

	if failed != failures {
		t.Errorf("failed %d out of %d", failed, total)
	}
}

func TestMateIn1(t *testing.T) {
	helper(t, "testdata/mateIn1.epd", 2, 0)
}

func TestMateIn2(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	helper(t, "testdata/mateIn2.epd", 6, 1)
}

/*
// Too long.
func TestMateIn3(t *testing.T) {
	helper(t, "testdata/mateIn3.epd", 6)
}
*/
