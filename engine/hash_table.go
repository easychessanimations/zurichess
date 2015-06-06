//go:generate stringer -type HashKind
// hash_table.go implements a global transposition table.

package engine

import (
	"unsafe"
)

var (
	DefaultHashTableSizeMB = 128      // default size in MB.
	GlobalHashTable        *HashTable // global transposition table.
)

type HashKind uint8

const (
	NoKind     HashKind = iota // No entry
	Exact                      // Exact score is known
	FailedLow                  // Search failed low, upper bound.
	FailedHigh                 // Search failed high, lower bound
)

// HashEntry is a value in the transposition table.
//
// TODO: store full move and age.
type HashEntry struct {
	// lock is used to handle hashing conflicts.
	// Normally, lock is derived from the position's Zobrist key.
	lock uint32

	Kind HashKind // type of hash
	Move Move     // best move

	Score int16 // score of the position. if mate, score is relative to current position.
	Depth int16 // remaining search depth
}

// HashTable is a transposition table.
// Engine uses this table to cache position scores so
// it doesn't have to research them again.
type HashTable struct {
	table []HashEntry // len(table) is a power of two and equals mask+1
	mask  uint32      // mask is used to determine the index in the table.
}

// NewHashTable builds transposition table that takes up to hashSizeMB megabytes.
func NewHashTable(hashSizeMB int) *HashTable {
	// Choose hashSize such that it is a power of two.
	hashEntrySize := uint64(unsafe.Sizeof(HashEntry{}))
	hashSize := uint64(hashSizeMB) << 20 / hashEntrySize

	for hashSize&(hashSize-1) != 0 {
		hashSize &= hashSize - 1
	}
	return &HashTable{
		table: make([]HashEntry, hashSize),
		mask:  uint32(hashSize - 1),
	}
}

// Size returns the number of entries in the table.
func (ht *HashTable) Size() int {
	return int(ht.mask + 1)
}

// split splits lock into a lock and two hash table indexes.
// expects mask to be at least 3 bits.
func split(lock uint64, mask uint32) (uint32, uint32, uint32) {
	hi := uint32(lock >> 32)
	lo := uint32(lock)
	h0 := lo & mask
	h1 := h0 ^ (lo >> 29)
	return hi, h0, h1
}

// Put puts a new entry in the database.
func (ht *HashTable) Put(pos *Position, entry HashEntry) {
	lock, key0, key1 := split(pos.Zobrist(), ht.mask)
	entry.lock = lock

	if e := &ht.table[key0]; e.lock == lock || e.Kind == NoKind || e.Depth+1 >= entry.Depth {
		ht.table[key0] = entry
	} else {
		ht.table[key1] = entry
	}
}

// Get returns the hash entry for position.
//
// Observation: due to collision errors, the HashEntry returned might be
// from a different table. However, these errors are not common because
// we use 32-bit lock + log_2(len(ht.table)) bits to avoid collisions.
func (ht *HashTable) Get(pos *Position) (HashEntry, bool) {
	lock, key0, key1 := split(pos.Zobrist(), ht.mask)
	if ht.table[key0].lock == lock && ht.table[key0].Kind != NoKind {
		return ht.table[key0], true
	}
	if ht.table[key1].lock == lock && ht.table[key1].Kind != NoKind {
		return ht.table[key1], true
	}
	return HashEntry{}, false
}

func init() {
	GlobalHashTable = NewHashTable(DefaultHashTableSizeMB)
}
