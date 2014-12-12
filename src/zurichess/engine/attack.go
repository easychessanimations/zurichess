package engine

import (
	"log"
	"math"
	"math/rand"
)

var (
	BbKnightAttack [64]Bitboard
	BbKingAttack   [64]Bitboard
	RookMagic      [64]magicInfo
	BishopMagic    [64]magicInfo
)

func init() {
	rand.Seed(5)
	initBbKnightAttack()
	initBbKingAttack()
	initRookMagic()
	initBishopMagic()
}

func initJumpAttack(jump [][2]int, attack []Bitboard) {
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			bb := Bitboard(0)
			for _, d := range jump {
				r_, f_ := r+d[0], f+d[1]
				if 0 > r_ || r_ >= 8 || 0 > f_ || f_ >= 8 {
					continue
				}
				bb |= RankFile(r_, f_).Bitboard()
			}
			attack[RankFile(r, f)] = bb
		}
	}
}

func initBbKnightAttack() {
	knightJump := [][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
	initJumpAttack(knightJump, BbKnightAttack[:])
	log.Println("BbKnightAttack initialized")
}

func initBbKingAttack() {
	kingJump := [][2]int{
		{-1, -1}, {-1, +0}, {-1, +1}, {+0, +1},
		{+1, +1}, {+1, +0}, {+1, -1}, {+0, -1},
	}
	initJumpAttack(kingJump, BbKingAttack[:])
	log.Println("BbKingAttack initialized")
}

func slidingAttack(sq Square, deltas [][2]int, occupancy Bitboard) Bitboard {
	r, f := sq.Rank(), sq.File()
	bb := Bitboard(0)
	for _, d := range deltas {
		r_, f_ := r, f
		for {
			r_, f_ = r_+d[0], f_+d[1]
			if 0 > r_ || r_ >= 8 || 0 > f_ || f_ >= 8 {
				// Stop when outside of the board.
				break
			}
			sq_ := RankFile(r_, f_)
			bb |= sq_.Bitboard()
			if occupancy&sq_.Bitboard() != 0 {
				// Stop when a piece was hit.
				break
			}
		}
	}
	return bb
}

// randMagic returns a random magic number
func randMagic() uint64 {
	r := uint64(rand.Int63())
	r &= uint64(rand.Int63())
	r &= uint64(rand.Int63())
	return (r << 6) + 1
}

func spell(magic uint64, shift uint, bb Bitboard) uint {
	bb = bb ^ (bb >> 23) // from fast-hash
	return uint(magic * uint64(bb) >> shift)
}

type magicInfo struct {
	store []Bitboard
	mask  Bitboard
	magic uint64
	shift uint
}

func (mi *magicInfo) Attack(ref Bitboard) Bitboard {
	return mi.store[spell(mi.magic, mi.shift, ref&mi.mask)]
}

type wizard struct {
	// Sliding deltas.
	Deltas        [][2]int
	MinShift      uint // Which shifts to search.
	MaxShift      uint
	MaxNumEntries uint // How much to search.

	numMagicTests uint
	magics        [64]uint64
	shifts        [64]uint // Number of bits for indexes.

	store     []Bitboard // Temporary store to check hash collisions.
	reference []Bitboard
	occupancy []Bitboard
}

func (wiz *wizard) tryMagicNumber(magic uint64, shift uint) bool {
	// Clear store.
	if len(wiz.store) < 1<<shift {
		wiz.store = make([]Bitboard, 1<<shift)
	}
	for j := range wiz.store[:1<<shift] {
		wiz.store[j] = 0
	}
	// Verify that magic gives a perfect hash.
	for i, bb := range wiz.reference {
		index := spell(magic, 64-shift, bb)
		if wiz.store[index] != 0 && wiz.store[index] != wiz.occupancy[i] {
			return false
		}
		wiz.store[index] = wiz.occupancy[i]
	}
	return true
}

func (wiz *wizard) searchMagic(sq Square, mi *magicInfo) {
	if wiz.shifts[sq] != 0 && wiz.shifts[sq] <= wiz.MinShift {
		// Don't search if shift is low enough already.
		return
	}

	// mask is the attack set on empty board minus the border.
	mask := slidingAttack(sq, wiz.Deltas, BbEmpty)
	// Compute border. Trick source: stockfish.
	border := (BbRank1 | BbRank8) & ^RankBb(sq.Rank())
	border |= (BbFileA | BbFileH) & ^FileBb(sq.File())
	mask &= ^border

	wiz.reference = wiz.reference[:0]
	wiz.occupancy = wiz.occupancy[:0]

	// Carry-Rippler trick to enumerate all subsets of mask.
	for subset := Bitboard(0); ; {
		attack := slidingAttack(sq, wiz.Deltas, subset)
		wiz.reference = append(wiz.reference, subset)
		wiz.occupancy = append(wiz.occupancy, attack)
		subset = (subset - mask) & mask
		if subset == 0 {
			break
		}
	}

	// Try magic numbers with small shifts.
	for i := 0; i < 1000 || wiz.shifts[sq] == 0; i++ {
		wiz.numMagicTests++

		// Pick a smaller shift than current best.
		var shift uint
		if wiz.shifts[sq] == 0 {
			shift = wiz.MaxShift
		} else {
			shift = wiz.shifts[sq] - 1
		}

		// Pick a good magic and test whether it gives a perfect hash.
		var magic uint64
		for Popcnt(uint64(mask)*magic) < 6 {
			magic = randMagic()
		}
		if wiz.tryMagicNumber(magic, shift) {
			wiz.magics[sq] = magic
			wiz.shifts[sq] = shift

			mi.store = make([]Bitboard, 1<<shift)
			copy(mi.store, wiz.store)
			mi.mask = mask
			mi.magic = magic
			mi.shift = 64 - shift
		}
	}
}

func (wiz *wizard) SearchMagics(mi []magicInfo) {
	numEntries := uint(math.MaxUint32)
	minShift := uint(math.MaxUint32)
	for numEntries > wiz.MaxNumEntries {
		numEntries = 0
		for sq := SquareMinValue; sq < SquareMaxValue; sq++ {
			wiz.searchMagic(sq, &mi[sq])
			numEntries += 1 << wiz.shifts[sq]
			if minShift > wiz.shifts[sq] {
				minShift = wiz.shifts[sq]
			}
		}
	}
	/*
		log.Println("numMagicTests =", wiz.numMagicTests,
			"; numEntries =", numEntries,
			"; minShift =", minShift)
	*/
}

func initRookMagic() {
	wiz := &wizard{
		Deltas:        [][2]int{{-1, +0}, {+1, +0}, {+0, -1}, {+0, +1}},
		MinShift:      10,
		MaxShift:      13,
		MaxNumEntries: 160000,
	}
	wiz.SearchMagics(RookMagic[:])
	log.Println("RookMagic initialized")
}

func initBishopMagic() {
	wiz := &wizard{
		Deltas:        [][2]int{{-1, +1}, {+1, +1}, {+1, -1}, {-1, -1}},
		MinShift:      5,
		MaxShift:      9,
		MaxNumEntries: 7000,
	}
	wiz.SearchMagics(BishopMagic[:])
	log.Println("BishopMagic initialized")
}
