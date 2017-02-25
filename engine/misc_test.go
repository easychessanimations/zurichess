// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"testing"

	. "bitbucket.org/zurichess/zurichess/board"
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
