// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
	I    int   // index in Weights
}

// Accum accumulates scores.
type Accum struct {
	M, E   int32              // mid game, end game
	Values [len(Weights)]int8 // input values
}

func (a *Accum) add(s Score) {
	a.M += s.M
	a.E += s.E
	a.Values[s.I] += 1
}

func (a *Accum) addN(s Score, n int32) {
	a.M += s.M * n
	a.E += s.E * n
	a.Values[s.I] += int8(n)
}

func (a *Accum) merge(o Accum) {
	a.M += o.M
	a.E += o.E
	for i := range o.Values {
		a.Values[i] += o.Values[i]
	}
}

func (e *Eval) merge() {
	e.Accum.M = e.pad[White].accum.M - e.pad[Black].accum.M
	e.Accum.E = e.pad[White].accum.E - e.pad[Black].accum.E
	for i := range e.Accum.Values {
		e.Accum.Values[i] = e.pad[White].accum.Values[i] - e.pad[Black].accum.Values[i]
	}
}

func initWeights() {
	for i := range Weights {
		Weights[i].I = i
	}
}
