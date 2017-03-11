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
	M, E   int32  // mid game, end game
	Values []int8 // input values
}

func resize(v []int8) []int8 {
	for len(v) < len(Weights) {
		v = append(v, 0)
	}
	return v
}

func (a *Accum) add(s Score) {
	a.M += s.M
	a.E += s.E
	a.Values = resize(a.Values)
	a.Values[s.I] += 1
}

func (a *Accum) addN(s Score, n int32) {
	a.M += s.M * n
	a.E += s.E * n
	a.Values = resize(a.Values)
	a.Values[s.I] += int8(n)
}

func (a *Accum) merge(o Accum) {
	a.M += o.M
	a.E += o.E
	a.Values = resize(a.Values)
	for i := range o.Values {
		a.Values[i] += o.Values[i]
	}
}

func (a *Accum) deduct(o Accum) {
	a.M -= o.M
	a.E -= o.E
	a.Values = resize(a.Values)
	for i := range o.Values {
		a.Values[i] -= o.Values[i]
	}
}
