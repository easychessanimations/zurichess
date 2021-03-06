// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	"testing"

	. "bitbucket.org/zurichess/board"
)

func TestCachePutGet(t *testing.T) {
	c1 := uint64(3080512559332270987)
	c2 := uint64(1670079002898303149)

	h1 := murmurSeed[NoFigure]
	h1 = murmurMix(h1, c1)
	h1 = murmurMix(h1, c2)

	ew, eb := Accum{1, 2}, Accum{3, 5}
	c := new(pawnsTable)
	c.put(h1, ew, eb)
	if gw, gb, ok := c.get(h1); !ok {
		t.Errorf("entry not in the cache, expecting a git")
	} else if ew != gw || eb != gb {
		t.Errorf("got get(%d) == %v, %v; wanted %v. %v", h1, gw, gb, ew, eb)
	}

	h2 := murmurSeed[NoFigure]
	h2 = murmurMix(h2, c2)
	h2 = murmurMix(h2, c1)
	if _, _, ok := c.get(h2); ok {
		t.Errorf("entry in the cache, expecting a miss")
	}
}
