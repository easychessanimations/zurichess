//go:generate stringer -type HashKind
// hash_table.go implements a global transposition table.
package engine

import (
	"unsafe"
)

type HashKind uint8

const (
	NoKind HashKind = iota
	Exact
	FailedLow
	FailedHigh
)

// HashEntry is a value in the transposition table.
type HashEntry struct {
	// Lock is used to handle hasing confligs.
	// Normally, Lock is the position's zobrist key.
	Lock  uint64
	Score int16    // score of the position
	Depth int16    // remaingin searching depth
	Move  Move     // best move found
	Kind  HashKind // type of hash
}

// HashTable is a transposition table.
// Engine uses such a table to cache moves so it doesn't recompute them again.
type HashTable struct {
	table     []HashEntry
	mask      uint64 // mask is used to determine the index in the table.
	Hit, Miss uint64
}

// NewHashTable builds transposition table that takes up to hashSizeMB megabytes.
func NewHashTable(hashSizeMB int) *HashTable {
	// Choose hashSize such that it is a power of two.
	hashEntrySize := uint64(unsafe.Sizeof(HashEntry{}))
	hashSize := (uint64(hashSizeMB) << 20) / hashEntrySize
	for hashSize&(hashSize-1) != 0 {
		hashSize &= hashSize - 1
	}

	return &HashTable{
		table: make([]HashEntry, hashSize),
		mask:  hashSize - 1,
	}
}

// ResetStats resets statistics.
func (ht *HashTable) ResetStats() {
	ht.Hit = 0
	ht.Miss = 0
}

// Size returns the number of entries in the table.
func (ht *HashTable) Size() int {
	return int(ht.mask + 1)
}

// Put puts a new entry in the database.
// Current strategy is to always replace.
func (ht *HashTable) Put(entry HashEntry) {
	key := entry.Lock & ht.mask
	if ht.table[key].Kind == NoKind ||
		ht.table[key].Lock == entry.Lock ||
		entry.Depth <= ht.table[key].Depth+1 {
		ht.table[key] = entry
	}
}

// Get returns an entry from the database.
// Lock of the returned entry matches lock.
func (ht *HashTable) Get(lock uint64) (HashEntry, bool) {
	key := lock & ht.mask
	if ht.table[key].Kind != NoKind && ht.table[key].Lock == lock {
		ht.Hit++
		return ht.table[key], true
	} else {
		ht.Miss++
		return HashEntry{}, false
	}
}

var (
	DefaultHashTableSizeMB = 32
	GlobalHashTable        *HashTable
)

func init() {
	GlobalHashTable = NewHashTable(DefaultHashTableSizeMB)
}