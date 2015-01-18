//go:generate stringer -type HashKind
// hash_table.go implements a global transposition table.

package engine

import (
	"unsafe"
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
	// Lock is used to handle hashing conflicts.
	// Normally, Lock is the position's zobrist key.
	Lock   uint64
	Score  int16    // score of the position
	Depth  int16    // remaining search depth
	Killer Move     // killer or best move found
	Kind   HashKind // type of hash
}

// HashTable is a transposition table.
// Engine uses a hash table to cache position scores so
// it doesn't have to recompute them again.
type HashTable struct {
	table []HashEntry // len(table) is a power of two and equals mask+1
	mask  uint64      // mask is used to determine the index in the table.
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

// Size returns the number of entries in the table.
func (ht *HashTable) Size() int {
	return int(ht.mask + 1)
}

// Put puts a new entry in the database.
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
		return ht.table[key], true
	} else {
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
