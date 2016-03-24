// attack.go generates move bitboards for all pieces.
// zurichess uses magic bitboards to generate sliding pieces moves.
// A very good description by Pradyumna Kannan can be read at:
// http://www.pradu.us/old/Nov27_2008/Buzz/research/magic/Bitboards.pdf
//
// TODO: move magic generation into an internal package.

package engine

import (
	"fmt"
	"math"
	"math/rand"
)

var (
	// bbPawnAttack contains pawn's attack tables.
	bbPawnAttack [64]Bitboard
	// bbKnightAttack contains knight's attack tables.
	bbKnightAttack [64]Bitboard
	// bbKingAttack contains king's attack tables (excluding castling).
	bbKingAttack [64]Bitboard
	bbKingArea   [64]Bitboard
	// bbSuperAttack contains queen piece's attack tables. This queen can jump.
	bbSuperAttack [64]Bitboard

	rookMagic    [64]magicInfo
	rookDeltas   = [][2]int{{-1, +0}, {+1, +0}, {+0, -1}, {+0, +1}}
	bishopMagic  [64]magicInfo
	bishopDeltas = [][2]int{{-1, +1}, {+1, +1}, {+1, -1}, {-1, -1}}
)

func init() {
	initBbPawnAttack()
	initBbKnightAttack()
	initBbKingAttack()
	initBbKingArea()
	initBbSuperAttack()
	initRookMagic()
	initBishopMagic()
}

func initJumpAttack(jump [][2]int, attack []Bitboard) {
	for r := 0; r < 8; r++ {
		for f := 0; f < 8; f++ {
			bb := Bitboard(0)
			for _, d := range jump {
				r0, f0 := r+d[0], f+d[1]
				if 0 > r0 || r0 >= 8 || 0 > f0 || f0 >= 8 {
					continue
				}
				bb |= RankFile(r0, f0).Bitboard()
			}
			attack[RankFile(r, f)] = bb
		}
	}
}

func initBbPawnAttack() {
	pawnJump := [][2]int{
		{-1, -1}, {-1, +1}, {+1, +1}, {+1, -1},
	}
	initJumpAttack(pawnJump, bbPawnAttack[:])
}

func initBbKnightAttack() {
	knightJump := [][2]int{
		{-2, -1}, {-2, +1}, {+2, -1}, {+2, +1},
		{-1, -2}, {-1, +2}, {+1, -2}, {+1, +2},
	}
	initJumpAttack(knightJump, bbKnightAttack[:])
}

func initBbKingAttack() {
	kingJump := [][2]int{
		{-1, -1}, {-1, +0}, {-1, +1}, {+0, +1},
		{+1, +1}, {+1, +0}, {+1, -1}, {+0, -1},
	}
	initJumpAttack(kingJump, bbKingAttack[:])
}

func initBbKingArea() {
	kingJump := [][2]int{
		{+1, -1}, {+1, +0}, {+1, +1},
		{+0, +1}, {+0, +0}, {+0, +1},
		{-1, -1}, {-1, +0}, {-1, +1},
	}
	initJumpAttack(kingJump, bbKingArea[:])
}

func initBbSuperAttack() {
	for sq := SquareMinValue; sq <= SquareMaxValue; sq++ {
		bbSuperAttack[sq] = slidingAttack(sq, rookDeltas, BbEmpty) | slidingAttack(sq, bishopDeltas, BbEmpty)
	}
}

func slidingAttack(sq Square, deltas [][2]int, occupancy Bitboard) Bitboard {
	r, f := sq.Rank(), sq.File()
	bb := Bitboard(0)
	for _, d := range deltas {
		r0, f0 := r, f
		for {
			r0, f0 = r0+d[0], f0+d[1]
			if 0 > r0 || r0 >= 8 || 0 > f0 || f0 >= 8 {
				// Stop when outside of the board.
				break
			}
			sq0 := RankFile(r0, f0)
			bb |= sq0.Bitboard()
			if occupancy&sq0.Bitboard() != 0 {
				// Stop when a piece was hit.
				break
			}
		}
	}
	return bb
}

// spell hashes bb using magic.
//
// magic stores in the upper 4 bits the shift.
// spell will return a number between 0 and 1<<shift that can be used
// to index in an array of size 1<<shift.
func spell(magic uint64, bb Bitboard) uint {
	shift := uint(magic >> 60)
	mul := magic * uint64(bb)
	return uint(mul >> ((64 - shift) & 63))
	// &63 lets the compiler now that shift fits 6 bits and should not generate CMPQ, SBBQ, ANDQ instructions on amd64
}

type magicInfo struct {
	mask  Bitboard   // square's mask.
	magic uint64     // magic multiplier. first 4 bits are the shift.
	store []Bitboard // attack boards of size 1<<shift
}

func (mi *magicInfo) Attack(ref Bitboard) Bitboard {
	return mi.store[spell(mi.magic, ref&mi.mask)]
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
		index := spell(magic, bb)
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
	return true
}

// randMagic returns a random magic number
func (wiz *wizard) randMagic() uint64 {
	r := uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	r &= uint64(wiz.Rand.Int63())
	return r << 1
}

// mask is the attack set on empty board minus the border.
func (wiz *wizard) mask(sq Square) Bitboard {
	// Compute border. Trick source: stockfish.
	border := (BbRank1 | BbRank8) & ^RankBb(sq.Rank())
	border |= (BbFileA | BbFileH) & ^FileBb(sq.File())
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

		if shift >= 16 {
			panic("shift too large, should fit in 4 bits")
		}

		// Pick a good magic and test whether it gives a perfect hash.
		var magic uint64
		for popcnt(uint64(mask)*magic) < 8 {
			magic = wiz.randMagic()>>4 + uint64(shift)<<60
		}
		wiz.tryMagicNumber(mi, sq, magic, shift)
	}
}

// SearchMagic finds new magics.
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
	if !wiz.tryMagicNumber(&mi[sq], sq, magic, shift) {
		panic(fmt.Sprintf("invalid magic: sq=%v magic=%d shift=%d", sq, magic, shift))
	}
}

func initRookMagic() {
	wiz := &wizard{
		Deltas:        rookDeltas,
		MinShift:      10,
		MaxShift:      13,
		MaxNumEntries: 130000,
		Rand:          rand.New(rand.NewSource(1)),
	}

	// A set of known good magics for rook.
	// Finding good rook magics is slow, so we just use some precomputed values.
	// For readability reasons, do not make an array.
	wiz.SetMagic(rookMagic[:], SquareA1, 13871104596958527489, 12)
	wiz.SetMagic(rookMagic[:], SquareA2, 13294766839654515745, 11)
	wiz.SetMagic(rookMagic[:], SquareA3, 12682176682988142722, 11)
	wiz.SetMagic(rookMagic[:], SquareA4, 12700151226211271712, 11)
	wiz.SetMagic(rookMagic[:], SquareA5, 12718166584917491776, 11)
	wiz.SetMagic(rookMagic[:], SquareA6, 12718167822132854784, 11)
	wiz.SetMagic(rookMagic[:], SquareA7, 12704936747524477440, 11)
	wiz.SetMagic(rookMagic[:], SquareA8, 14123447863346202689, 12)
	wiz.SetMagic(rookMagic[:], SquareB1, 13276629294211416066, 11)
	wiz.SetMagic(rookMagic[:], SquareB2, 11819767592120262662, 10)
	wiz.SetMagic(rookMagic[:], SquareB3, 11547229994870669634, 10)
	wiz.SetMagic(rookMagic[:], SquareB4, 11533727773575094272, 10)
	wiz.SetMagic(rookMagic[:], SquareB5, 11533754931743834120, 10)
	wiz.SetMagic(rookMagic[:], SquareB6, 11533718923258118148, 10)
	wiz.SetMagic(rookMagic[:], SquareB7, 11529285419787157760, 10)
	wiz.SetMagic(rookMagic[:], SquareB8, 13294666783274303617, 11)
	wiz.SetMagic(rookMagic[:], SquareC1, 12718211046152601728, 11)
	wiz.SetMagic(rookMagic[:], SquareC2, 11678115399045480512, 10)
	wiz.SetMagic(rookMagic[:], SquareC3, 11709430499771490308, 10)
	wiz.SetMagic(rookMagic[:], SquareC4, 11533895744381649024, 10)
	wiz.SetMagic(rookMagic[:], SquareC5, 11817621348392375872, 10)
	wiz.SetMagic(rookMagic[:], SquareC6, 11602962319172567074, 10)
	wiz.SetMagic(rookMagic[:], SquareC7, 11533721396085526656, 10)
	wiz.SetMagic(rookMagic[:], SquareC8, 12720699928691621913, 11)
	wiz.SetMagic(rookMagic[:], SquareD1, 12862289334011691013, 11)
	wiz.SetMagic(rookMagic[:], SquareD2, 11534000159346151425, 10)
	wiz.SetMagic(rookMagic[:], SquareD3, 11565282326138061056, 10)
	wiz.SetMagic(rookMagic[:], SquareD4, 11817491610299678752, 10)
	wiz.SetMagic(rookMagic[:], SquareD5, 11565389047350167552, 10)
	wiz.SetMagic(rookMagic[:], SquareD6, 11529232638388764800, 10)
	wiz.SetMagic(rookMagic[:], SquareD7, 11605918111094538368, 10)
	wiz.SetMagic(rookMagic[:], SquareD8, 12700432596172017665, 11)
	wiz.SetMagic(rookMagic[:], SquareE1, 12682418197727551552, 11)
	wiz.SetMagic(rookMagic[:], SquareE2, 11637582929600710656, 10)
	wiz.SetMagic(rookMagic[:], SquareE3, 11549622531702917120, 10)
	wiz.SetMagic(rookMagic[:], SquareE4, 11602399100508045348, 10)
	wiz.SetMagic(rookMagic[:], SquareE5, 11529496594143520768, 10)
	wiz.SetMagic(rookMagic[:], SquareE6, 11531466880309035136, 10)
	wiz.SetMagic(rookMagic[:], SquareE7, 11530342595645408384, 10)
	wiz.SetMagic(rookMagic[:], SquareE8, 13258878786679607309, 11)
	wiz.SetMagic(rookMagic[:], SquareF1, 12826295771027603457, 11)
	wiz.SetMagic(rookMagic[:], SquareF2, 11601554222457422848, 10)
	wiz.SetMagic(rookMagic[:], SquareF3, 11604650889784197248, 10)
	wiz.SetMagic(rookMagic[:], SquareF4, 11538297014261514368, 10)
	wiz.SetMagic(rookMagic[:], SquareF5, 11556238845011823616, 10)
	wiz.SetMagic(rookMagic[:], SquareF6, 11565314281088483588, 10)
	wiz.SetMagic(rookMagic[:], SquareF7, 11529287615987843200, 10)
	wiz.SetMagic(rookMagic[:], SquareF8, 12754757236671677442, 11)
	wiz.SetMagic(rookMagic[:], SquareG1, 12718167546776256768, 11)
	wiz.SetMagic(rookMagic[:], SquareG2, 12110742384527147009, 10)
	wiz.SetMagic(rookMagic[:], SquareG3, 11529219444383613480, 10)
	wiz.SetMagic(rookMagic[:], SquareG4, 11673480059784528459, 10)
	wiz.SetMagic(rookMagic[:], SquareG5, 11540476278587002945, 10)
	wiz.SetMagic(rookMagic[:], SquareG6, 12393915323072643077, 10)
	wiz.SetMagic(rookMagic[:], SquareG7, 11529778031507014144, 10)
	wiz.SetMagic(rookMagic[:], SquareG8, 13559221442951841924, 11)
	wiz.SetMagic(rookMagic[:], SquareH1, 13907117851032979712, 12)
	wiz.SetMagic(rookMagic[:], SquareH2, 12682277290311172864, 11)
	wiz.SetMagic(rookMagic[:], SquareH3, 12790506616286104708, 11)
	wiz.SetMagic(rookMagic[:], SquareH4, 12691284498155917568, 11)
	wiz.SetMagic(rookMagic[:], SquareH5, 12691214707118309444, 11)
	wiz.SetMagic(rookMagic[:], SquareH6, 12691162564172840964, 11)
	wiz.SetMagic(rookMagic[:], SquareH7, 12898328025837603328, 11)
	wiz.SetMagic(rookMagic[:], SquareH8, 13979186448239231522, 12)

	// Enable the next line to find new magics.
	// wiz.SearchMagics(rookMagic[:])
}

func initBishopMagic() {
	wiz := &wizard{
		Deltas:        bishopDeltas,
		MinShift:      5,
		MaxShift:      9,
		MaxNumEntries: 6000,
		Rand:          rand.New(rand.NewSource(1)),
	}

	// Bishop magics, unlike rook magics are easy to find.
	wiz.SearchMagics(bishopMagic[:])
}

// KnightMobility returns all squares a knight can reach from sq.
func KnightMobility(sq Square) Bitboard {
	return bbKnightAttack[sq]
}

// BishopMobility returns the squares a bishop can reach from sq given all pieces.
func BishopMobility(sq Square, all Bitboard) Bitboard {
	return bishopMagic[sq].Attack(all)
}

// RookMobility returns the squares a rook can reach from sq given all pieces.
func RookMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all)
}

// QueenMobility returns the squares a queen can reach from sq given all pieces.
func QueenMobility(sq Square, all Bitboard) Bitboard {
	return rookMagic[sq].Attack(all) | bishopMagic[sq].Attack(all)
}

// KingMobility returns all squares a king can reach from sq.
// Doesn't include castling.
func KingMobility(sq Square) Bitboard {
	return bbKingAttack[sq]
}
