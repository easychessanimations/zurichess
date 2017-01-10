// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !coach

package engine

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
}

// Accum accumulates scores.
type Accum struct {
	M, E int32 // mid game, end game
}

func (a *Accum) add(s Score) {
	a.M += s.M
	a.E += s.E
}

func (a *Accum) addN(s Score, n int32) {
	a.M += s.M * n
	a.E += s.E * n
}

func (a *Accum) merge(o Accum) {
	a.M += o.M
	a.E += o.E
}

func (e *Eval) merge() {
	e.Accum.M = e.pad[White].accum.M - e.pad[Black].accum.M
	e.Accum.E = e.pad[White].accum.E - e.pad[Black].accum.E
}

func initWeights() {
}
