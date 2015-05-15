package engine

import "testing"

func TestPutGet(t *testing.T) {
	white := Bitboard(0x123)
	black := Bitboard(0xff03312)
	score := Score{123, 456}

	var pt pawnTable
	if _, has := pt.get(white, black); has {
		t.Errorf("entry was not expected")
	}

	pt.put(white, black, score)
	if actual, has := pt.get(white, black); !has {
		t.Errorf("entry not cached")
	} else if score != actual {
		t.Errorf("expected score %d, got %d", score, actual)
	}
}
