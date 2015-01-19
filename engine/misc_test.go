package engine

import "testing"

func TestLSB(t *testing.T) {
	data := []struct {
		n   uint64
		lsb uint64
	}{
		{1, 1},
		{13, 1},
		{24, 8},
		{16, 16},
		{17 << 20, 1 << 20},
	}

	for _, d := range data {
		actual := LSB(d.n)
		if d.lsb != actual {
			t.Errorf("expected LSB(%d) == %d, got %d", d.n, d.lsb, actual)
		}
	}
}

func TestLogN(t *testing.T) {
	for e := uint(0); e < 64; e++ {
		n := uint64(1) << e
		actual := LogN(n)
		if actual != e {
			t.Errorf("expected LogN(%d) == %d, got %d", n, e, actual)
		}
	}
}
