package engine

import (
	"log"
	"math"
	"math/rand"
)

var (
	BbPawnAttack   [64]Bitboard // pawn's attack tables
	BbKnightAttack [64]Bitboard // knight's attack tables
	BbKingAttack   [64]Bitboard // king's attack tables (excluding castling)
	BbSuperAttack  [64]Bitboard // super piece's attack tables

	RookMagic   [64]magicInfo
	BishopMagic [64]magicInfo

	rookDeltas   = [][2]int{{-1, +0}, {+1, +0}, {+0, -1}, {+0, +1}}
	bishopDeltas = [][2]int{{-1, +1}, {+1, +1}, {+1, -1}, {-1, -1}}
)

func init() {
	initBbPawnAttack()
	initBbKnightAttack()
	initBbKingAttack()
	initBbSuperAttack()
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

func initBbPawnAttack() {
	pawnJump := [][2]int{
		{-1, -1}, {-1, +1}, {+1, +1}, {+1, -1},
	}
	initJumpAttack(pawnJump, BbPawnAttack[:])
}

func initBbKnightAttack() {
	knightJump := [][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
	initJumpAttack(knightJump, BbKnightAttack[:])
}

func initBbKingAttack() {
	kingJump := [][2]int{
		{-1, -1}, {-1, +0}, {-1, +1}, {+0, +1},
		{+1, +1}, {+1, +0}, {+1, -1}, {+0, -1},
	}
	initJumpAttack(kingJump, BbKingAttack[:])
}

func initBbSuperAttack() {
	for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
		BbSuperAttack[sq] = slidingAttack(sq, rookDeltas, BbEmpty) | slidingAttack(sq, bishopDeltas, BbEmpty)
	}
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

func spell(magic uint64, shift uint, bb Bitboard) uint {
	hi := uint32(bb>>32) * uint32(magic)
	lo := uint32(magic>>32) * uint32(bb)
	return uint((hi ^ lo) >> shift)
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
	Rand          *rand.Rand

	numMagicTests uint
	magics        [64]uint64
	shifts        [64]uint // Number of bits for indexes.

	store     []Bitboard // Temporary store to check hash collisions.
	reference []Bitboard
	occupancy []Bitboard
}

func (wiz *wizard) tryMagicNumber(mi *magicInfo, sq Square, magic uint64, shift uint) bool {
	wiz.numMagicTests++

	// Clear store.
	if len(wiz.store) < 1<<shift {
		wiz.store = make([]Bitboard, 1<<shift)
	}
	for j := range wiz.store[:1<<shift] {
		wiz.store[j] = 0
	}

	// Verify that magic gives a perfect hash.
	for i, bb := range wiz.reference {
		index := spell(magic, 32-shift, bb)
		if wiz.store[index] != 0 && wiz.store[index] != wiz.occupancy[i] {
			return false
		}
		wiz.store[index] = wiz.occupancy[i]
	}

	// Perfect hash, store it.
	wiz.magics[sq] = magic
	wiz.shifts[sq] = shift

	mi.store = make([]Bitboard, 1<<shift)
	copy(mi.store, wiz.store)
	mi.mask = wiz.mask(sq)
	mi.magic = magic
	mi.shift = 32 - shift
	return true
}

// randMagic returns a random magic number
func (wiz *wizard) randMagic() uint64 {
	r := uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	return (r << 6) + 1
}

// mask is the attack set on empty board minus the border.
func (wiz *wizard) mask(sq Square) Bitboard {
	// Compute border. Trick source: stockfish.
	border := (RankBb(0) | RankBb(7)) & ^RankBb(sq.Rank())
	border |= (FileBb(0) | FileBb(7)) & ^FileBb(sq.File())
	return ^border & slidingAttack(sq, wiz.Deltas, BbEmpty)
}

// prepare computes reference and occupancy tables for a square.
func (wiz *wizard) prepare(sq Square) {
	wiz.reference = wiz.reference[:0]
	wiz.occupancy = wiz.occupancy[:0]

	// Carry-Rippler trick to enumerate all subsets of mask.
	for mask, subset := wiz.mask(sq), Bitboard(0); ; {
		attack := slidingAttack(sq, wiz.Deltas, subset)
		wiz.reference = append(wiz.reference, subset)
		wiz.occupancy = append(wiz.occupancy, attack)
		subset = (subset - mask) & mask
		if subset == 0 {
			break
		}
	}
}

func (wiz *wizard) searchMagic(sq Square, mi *magicInfo) {
	if wiz.shifts[sq] != 0 && wiz.shifts[sq] <= wiz.MinShift {
		// Don't search if shift is low enough.
		return
	}

	// Try magic numbers with small shifts.
	wiz.prepare(sq)
	mask := wiz.mask(sq)
	for i := 0; i < 100 || wiz.shifts[sq] == 0; i++ {
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
			magic = wiz.randMagic()
		}
		wiz.tryMagicNumber(mi, sq, magic, shift)
	}
}

func (wiz *wizard) SearchMagics(mi []magicInfo) {
	numEntries := uint(math.MaxUint32)
	minShift := uint(math.MaxUint32)
	for numEntries > wiz.MaxNumEntries {
		numEntries = 0
		for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
			wiz.searchMagic(sq, &mi[sq])
			numEntries += 1 << wiz.shifts[sq]
			if minShift > wiz.shifts[sq] {
				minShift = wiz.shifts[sq]
			}
		}
	}
}

func (wiz *wizard) SetMagic(mi []magicInfo, sq Square, magic uint64, shift uint) {
	wiz.prepare(sq)
	if ok := wiz.tryMagicNumber(&mi[sq], sq, magic, shift); !ok {
		log.Printf("invalid magic: sq=%v magic=%d shift=%d",
			sq, magic, shift)
	}
}

func initRookMagic() {
	wiz := &wizard{
		Deltas:        rookDeltas,
		MinShift:      10,
		MaxShift:      13,
		MaxNumEntries: 150000,
		Rand:          rand.New(rand.NewSource(1)),
	}

	// A set of known good magics.
	// For readability reasons, do not make an array.
	wiz.SetMagic(RookMagic[:], SquareA1, 9295500549393752449, 12)
	wiz.SetMagic(RookMagic[:], SquareA2, 18049720605753345, 11)
	wiz.SetMagic(RookMagic[:], SquareA3, 10376328864348180673, 11)
	wiz.SetMagic(RookMagic[:], SquareA4, 291045271964516929, 11)
	wiz.SetMagic(RookMagic[:], SquareA5, 2307005820086321217, 11)
	wiz.SetMagic(RookMagic[:], SquareA6, 281750458860033, 11)
	wiz.SetMagic(RookMagic[:], SquareA7, 18305237051834625, 11)
	wiz.SetMagic(RookMagic[:], SquareA8, 1297054834895110401, 12)
	wiz.SetMagic(RookMagic[:], SquareB1, 13889171757164658753, 11)
	wiz.SetMagic(RookMagic[:], SquareB2, 4647785459341074433, 10)
	wiz.SetMagic(RookMagic[:], SquareB3, 9308940704799531073, 10)
	wiz.SetMagic(RookMagic[:], SquareB4, 1153836319825219969, 10)
	wiz.SetMagic(RookMagic[:], SquareB5, 9233507609921728513, 10)
	wiz.SetMagic(RookMagic[:], SquareB6, 2255993851355201, 10)
	wiz.SetMagic(RookMagic[:], SquareB7, 5136922622066945, 10)
	wiz.SetMagic(RookMagic[:], SquareB8, 9259418567802208257, 11)
	wiz.SetMagic(RookMagic[:], SquareC1, 36169603228895233, 11)
	wiz.SetMagic(RookMagic[:], SquareC2, 2378182305882579073, 10)
	wiz.SetMagic(RookMagic[:], SquareC3, 76563401821651073, 10)
	wiz.SetMagic(RookMagic[:], SquareC4, 603554055084056705, 10)
	wiz.SetMagic(RookMagic[:], SquareC5, 281612432966145, 10)
	wiz.SetMagic(RookMagic[:], SquareC6, 1612307392792895489, 10)
	wiz.SetMagic(RookMagic[:], SquareC7, 73271459176251649, 10)
	wiz.SetMagic(RookMagic[:], SquareC8, 9223380923163090945, 11)
	wiz.SetMagic(RookMagic[:], SquareD1, 4620702031007450625, 11)
	wiz.SetMagic(RookMagic[:], SquareD2, 9147971120693377, 10)
	wiz.SetMagic(RookMagic[:], SquareD3, 144994816713835009, 10)
	wiz.SetMagic(RookMagic[:], SquareD4, 16928122915000321, 10)
	wiz.SetMagic(RookMagic[:], SquareD5, 304085102237697, 10)
	wiz.SetMagic(RookMagic[:], SquareD6, 2594144870807388673, 10)
	wiz.SetMagic(RookMagic[:], SquareD7, 2306969090851356737, 10)
	wiz.SetMagic(RookMagic[:], SquareD8, 2450107799592568065, 11)
	wiz.SetMagic(RookMagic[:], SquareE1, 9013814070622209, 11)
	wiz.SetMagic(RookMagic[:], SquareE2, 288371130820461185, 10)
	wiz.SetMagic(RookMagic[:], SquareE3, 2884555715983188353, 10)
	wiz.SetMagic(RookMagic[:], SquareE4, 8800389571841, 10)
	wiz.SetMagic(RookMagic[:], SquareE5, 396334668968887297, 10)
	wiz.SetMagic(RookMagic[:], SquareE6, 9223380839395557889, 10)
	wiz.SetMagic(RookMagic[:], SquareE7, 149606662934785, 10)
	wiz.SetMagic(RookMagic[:], SquareE8, 4692757757756180737, 11)
	wiz.SetMagic(RookMagic[:], SquareF1, 4755811110967395329, 11)
	wiz.SetMagic(RookMagic[:], SquareF2, 1441222275323134017, 10)
	wiz.SetMagic(RookMagic[:], SquareF3, 72097194711195713, 10)
	wiz.SetMagic(RookMagic[:], SquareF4, 288511881193724929, 10)
	wiz.SetMagic(RookMagic[:], SquareF5, 342506678958162561, 10)
	wiz.SetMagic(RookMagic[:], SquareF6, 18023555379888641, 10)
	wiz.SetMagic(RookMagic[:], SquareF7, 175939041364993, 10)
	wiz.SetMagic(RookMagic[:], SquareF8, 297527932510274049, 11)
	wiz.SetMagic(RookMagic[:], SquareG1, 18084840452688385, 11)
	wiz.SetMagic(RookMagic[:], SquareG2, 9147941038985345, 10)
	wiz.SetMagic(RookMagic[:], SquareG3, 1191624594721628161, 10)
	wiz.SetMagic(RookMagic[:], SquareG4, 4616194020396378369, 10)
	wiz.SetMagic(RookMagic[:], SquareG5, 5837228354998140929, 10)
	wiz.SetMagic(RookMagic[:], SquareG6, 145144269111297, 10)
	wiz.SetMagic(RookMagic[:], SquareG7, 2886877736381743169, 10)
	wiz.SetMagic(RookMagic[:], SquareG8, 281483717902977, 11)
	wiz.SetMagic(RookMagic[:], SquareH1, 72057881834303617, 12)
	wiz.SetMagic(RookMagic[:], SquareH2, 72761286848430209, 11)
	wiz.SetMagic(RookMagic[:], SquareH3, 2360027496821063681, 11)
	wiz.SetMagic(RookMagic[:], SquareH4, 9345008817802859393, 11)
	wiz.SetMagic(RookMagic[:], SquareH5, 450979987585, 11)
	wiz.SetMagic(RookMagic[:], SquareH6, 1189100664211406849, 11)
	wiz.SetMagic(RookMagic[:], SquareH7, 613193309965254785, 11)
	wiz.SetMagic(RookMagic[:], SquareH8, 13907896865786961985, 12)

	// Normally not needed, but just in case the magics are wrong.
	wiz.SearchMagics(RookMagic[:])
}

func initBishopMagic() {
	wiz := &wizard{
		Deltas:        bishopDeltas,
		MinShift:      5,
		MaxShift:      9,
		MaxNumEntries: 6000,
		Rand:          rand.New(rand.NewSource(1)),
	}
	wiz.SearchMagics(BishopMagic[:])
}
