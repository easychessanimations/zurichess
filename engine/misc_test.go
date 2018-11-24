// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"testing"

	. "bitbucket.org/zurichess/board"
)

func TestDistance(t *testing.T) {
	data := []struct {
		i, j Square
		d    int32
	}{
		{SquareA1, SquareA8, 7},
		{SquareA1, SquareH8, 7},
		{SquareB2, SquareB2, 0},
		{SquareB2, SquareC3, 1},
		{SquareE5, SquareD4, 1},
		{SquareE5, SquareD4, 1},
		{SquareE1, SquareG5, 4},
	}

	for i, d := range data {
		if got, want := distance[d.i][d.j], d.d; got != want {
			t.Errorf("#%d wanted distance[%v][%v] == %d, got %d", i, d.i, d.j, want, got)
		}
	}
}

func TestMurmurMixSwap(t *testing.T) {
	c1 := uint64(3080512559332270987)
	c2 := uint64(1670079002898303149)

	h1 := murmurSeed[NoFigure]
	h1 = murmurMix(h1, c1)
	h1 = murmurMix(h1, c2)

	h2 := murmurSeed[NoFigure]
	h2 = murmurMix(h2, c2)
	h2 = murmurMix(h2, c1)

	if h1 == h2 {
		t.Errorf("murmurMix(c1, c2) == murmurMix(c2, c1) (%d, %d), wanted different", h1, h2)
	}
}
