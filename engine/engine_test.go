// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"strings"
	"testing"
)

func TestGame(t *testing.T) {
	pos, _ := PositionFromFEN(FENStartPos)
	eng := NewEngine(pos, nil, Options{})
	for i := 0; i < 1; i++ {
		tc := NewFixedDepthTimeControl(pos, 3)
		tc.Start(false)
		_, pv := eng.Play(tc)
		eng.DoMove(pv[0])
	}
}

func TestMateIn1(t *testing.T) {
	for i, d := range mateIn1 {
		pos, _ := PositionFromFEN(d.fen)
		bm, err := pos.UCIToMove(d.bm)
		if err != nil {
			t.Errorf("#%d cannot parse move %s", i, d.bm)
			continue
		}

		tc := NewFixedDepthTimeControl(pos, 2)
		tc.Start(false)
		eng := NewEngine(pos, nil, Options{})
		_, pv := eng.Play(tc)

		if len(pv) != 1 {
			t.Errorf("#%d Expected at most one move, got %d", i, len(pv))
			t.Errorf("position is %v", pos)
			continue
		}

		if pv[0] != bm {
			t.Errorf("#%d expected move %v, got %v", i, bm, pv[0])
			t.Errorf("position is %v", pos)
			continue
		}
	}
}

// Test score is the same if we start with the position or move.
func TestScore(t *testing.T) {
	for _, game := range testGames {
		pos, _ := PositionFromFEN(FENStartPos)
		dynamic := NewEngine(pos, nil, Options{})
		static := NewEngine(pos, nil, Options{})

		moves := strings.Fields(game)
		for _, move := range moves {
			m, _ := pos.UCIToMove(move)
			if !pos.IsPseudoLegal(m) {
				// t.Fatalf("bad bad bad")
			}

			dynamic.DoMove(m)
			static.SetPosition(pos)
			if dynamic.Score() != static.Score() {
				t.Fatalf("expected static score %v, got dynamic score %v", static.Score(), dynamic.Score())
			}
		}
	}
}

func TestEndGamePosition(t *testing.T) {
	pos, _ := PositionFromFEN("6k1/5p1p/4p1p1/3p4/5P1P/8/3r2q1/6K1 w - - 2 55")
	tc := NewFixedDepthTimeControl(pos, 3)
	tc.Start(false)
	eng := NewEngine(pos, nil, Options{})
	_, pv := eng.Play(tc)
	if pv != nil {
		t.Errorf("got %d moves (nonil, pv), expected nil pv", len(pv))
	}
}

// pvLogger logs the PV.
// It will panic if pvs are not in order.
type pvLog struct {
	depth   int32
	multiPV int
	score   int32
	moves   []Move
}

type pvLogger []pvLog

func (l *pvLogger) BeginSearch()                           {}
func (l *pvLogger) EndSearch()                             {}
func (l *pvLogger) CurrMove(depth int, move Move, num int) {}

func (l *pvLogger) PrintPV(stats Stats, multiPV int, score int32, moves []Move) {
	*l = append(*l, pvLog{
		depth:   stats.Depth,
		multiPV: multiPV,
		score:   score,
		moves:   moves,
	})
}

func TestMultiPV(t *testing.T) {
	for f, fen := range testFENs {
		pos, _ := PositionFromFEN(fen)
		tc := NewFixedDepthTimeControl(pos, 4)
		tc.Start(false)
		pvl := pvLogger{}
		eng := NewEngine(pos, &pvl, Options{MultiPV: 3})
		eng.Play(tc)

		// Check the number of iterations.
		numIterations := 0
		for i := range pvl {
			if pvl[i].multiPV == 1 {
				numIterations++
			}
		}
		if numIterations != 4+1 {
			t.Errorf("#%d %s: expected 4+1 iterations, got %d", f, fen, numIterations)
		}

		// Check score and depth order.
		for i := 1; i < len(pvl); i++ {
			if pvl[i-1].depth > pvl[i].depth {
				// TODO: this is not really correct if we repeat the PVS lines
				t.Errorf("#%d %s: wrong depth order", f, fen)
			}
			if pvl[i-1].depth == pvl[i].depth && pvl[i-1].score < pvl[i].score {
				t.Errorf("#%d %s: wrong score order", f, fen)
			}
		}

		// Check different moves for the same iterations.
		for i := range pvl {
			for j := range pvl {
				if i <= j {
					continue
				}
				if pvl[i].depth != pvl[j].depth || pvl[i].multiPV == pvl[j].multiPV {
					continue
				}
				if len(pvl[i].moves) == 0 || len(pvl[j].moves) == 0 {
					continue
				}
				if pvl[i].moves[0] == pvl[j].moves[0] {
					t.Errorf("#%d %s: got identical moves", f, fen)
				}
			}
		}
	}
}

func BenchmarkGame(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pos, _ := PositionFromFEN(FENStartPos)
		eng := NewEngine(pos, nil, Options{})
		for j := 0; j < 20; j++ {
			tc := NewFixedDepthTimeControl(pos, 4)
			tc.Start(false)
			_, pv := eng.Play(tc)
			eng.DoMove(pv[0])
		}
	}
}
