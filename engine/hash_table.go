// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// hash_table.go implements a global transposition table.

package engine

import (
	"unsafe" // for sizeof

	. "bitbucket.org/zurichess/zurichess/board"
)

var (
	// DefaultHashTableSizeMB is the default size in MB.
	DefaultHashTableSizeMB = 64
	// GlobalHashTable is the global transposition table.
	GlobalHashTable *HashTable
)

type hashFlags uint8

const (
	exact      hashFlags = 1 << iota // exact score is known
	failedLow                        // Search failed low, upper bound.
	failedHigh                       // Search failed high, lower bound
	hasStatic                        // entry contains static evaluation
)

// isInBounds returns true if score matches range defined by α, β and flags.
func isInBounds(flags hashFlags, α, β, score int32) bool {
	if flags&exact != 0 {
		// Simply return if the score is exact.
		return true
	}
	if flags&failedLow != 0 && score <= α {
		// Previously the move failed low so the actual score is at most
		// entry.score. If that's lower than α this will also fail low.
		return true
	}
	if flags&failedHigh != 0 && score >= β {
		// Previously the move failed high so the actual score is at least
		// entry.score. If that's higher than β this will also fail high.
		return true
	}
	return false
}

// getBound returns the bound for score relative to α and β.
func getBound(α, β, score int32) hashFlags {
	if score <= α {
		return failedLow
	}
	if score >= β {
		return failedHigh
	}
	return exact
}

// hashEntry is a value in the transposition table.
type hashEntry struct {
	lock   uint32    // lock is used to handle hashing conflicts.
	move   Move      // best move
	score  int16     // score of the position. if mate, score is relative to current position.
	static int16     // static score of the position (not yet used)
	depth  int8      // remaining search depth
	kind   hashFlags // type of hash
}

// HashTable is a transposition table.
// Engine uses this table to cache position scores so
// it doesn't have to research them again.
type HashTable struct {
	table []hashEntry // len(table) is a power of two and equals mask+1
	mask  uint32      // mask is used to determine the index in the table.
}

// NewHashTable builds transposition table that takes up to hashSizeMB megabytes.
func NewHashTable(hashSizeMB int) *HashTable {
	// Choose hashSize such that it is a power of two.
	hashEntrySize := uint64(unsafe.Sizeof(hashEntry{}))
	hashSize := uint64(hashSizeMB) << 20 / hashEntrySize

	for hashSize&(hashSize-1) != 0 {
		hashSize &= hashSize - 1
	}
	return &HashTable{
		table: make([]hashEntry, hashSize),
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

// put puts a new entry in the database.
func (ht *HashTable) put(pos *Position, entry hashEntry) {
	lock, key0, key1 := split(pos.Zobrist(), ht.mask)
	entry.lock = lock

	if e := &ht.table[key0]; e.lock == lock || e.kind == 0 || e.depth+1 >= entry.depth {
		ht.table[key0] = entry
	} else {
		ht.table[key1] = entry
	}
}

// get returns the hash entry for position.
//
// Observation: due to collision errors, the hashEntry returned might be
// from a different table. However, these errors are not common because
// we use 32-bit lock + log_2(len(ht.table)) bits to avoid collisions.
func (ht *HashTable) get(pos *Position) hashEntry {
	lock, key0, key1 := split(pos.Zobrist(), ht.mask)
	if ht.table[key0].lock == lock {
		return ht.table[key0]
	}
	if ht.table[key1].lock == lock {
		return ht.table[key1]
	}
	return hashEntry{}
}

// Clear removes all entries from hash.
func (ht *HashTable) Clear() {
	for i := range ht.table {
		ht.table[i] = hashEntry{}
	}
}

func init() {
	GlobalHashTable = NewHashTable(DefaultHashTableSizeMB)
}
