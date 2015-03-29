//go:generate stringer -type HashKind
// hash_table.go implements a global transposition table.

package engine

import (
	"unsafe"
)

var (
	DefaultHashTableSizeMB = 96
	GlobalHashTable        *HashTable
)

type HashKind uint8

const (
	NoKind     HashKind = iota
	Exact               // Exact score is known
	FailedLow           // Search failed low, upper bound.
	FailedHigh          // Search failed high, lower bound
)

// HashEntry is a value in the transposition table.
type HashEntry struct {
	Kind   HashKind // type of hash
	Target Piece    // from favorite move
	From   Square   // from favorite move
	To     Square   // from favorite move

	// lock is used to handle hashing conflicts.
	// Normally, lock is derived from the position's Zobrist key.
	lock uint32

	Score int16 // score of the position
	Depth int16 // remaining search depth
}

// HashTable is a transposition table.
// Engine uses a hash table to cache position scores so
// it doesn't have to recompute them again.
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
func split(lock uint64, mask uint32) (uint32, uint32, uint32) {
	return uint32(lock >> 32), uint32(lock>>0) & mask, uint32(lock>>8) & mask
}

// Put puts a new entry in the database.
func (ht *HashTable) Put(pos *Position, entry HashEntry) {
	lock, key0, key1 := split(pos.Zobrist(), ht.mask)
	entry.lock = lock

	e0 := &ht.table[key0]
	if e0.Kind == NoKind || e0.lock == lock || e0.Depth+1 >= entry.Depth {
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
	if ht.table[key0].Kind != NoKind && ht.table[key0].lock == lock {
		return ht.table[key0], true
	}
	if ht.table[key1].Kind != NoKind && ht.table[key1].lock == lock {
		return ht.table[key1], true
	}
	return HashEntry{}, false
}

func init() {
	GlobalHashTable = NewHashTable(DefaultHashTableSizeMB)
}
