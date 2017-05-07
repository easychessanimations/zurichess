// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package engine

import (
	. "bitbucket.org/zurichess/zurichess/board"
)

const (
	pvTableSize = 1 << 13
	pvTableMask = pvTableSize - 1
)

type pvEntry struct {
	// lock is used to handled hash conflicts.
	// Normally set to position's Zobrist key.
	lock uint64
	// move on pricipal variation for this position.
	move Move
}

// pvTable is like hash table, but only to keep principal variation.
//
// The additional table to store the PV was suggested by Robert Hyatt. See
//
// * http://www.talkchess.com/forum/viewtopic.php?topic_view=threads&p=369163&t=35982
// * http://www.talkchess.com/forum/viewtopic.php?t=36099
//
// During alpha-beta search entries that are on principal variation,
// are exact nodes, i.e. their score lies exactly between alpha and beta.
type pvTable []pvEntry

// newPvTable returns a new pvTable.
func newPvTable() pvTable {
	return pvTable(make([]pvEntry, pvTableSize))
}

// Put inserts a new entry.  Ignores NullMoves.
func (pv pvTable) Put(pos *Position, move Move) {
	if move == NullMove {
		return
	}

	// Based on pos.Zobrist() two entries are looked up.
	// If any of the two entries in the table matches
	// current position, then that one is replaced.
	// Otherwise, the older is replaced.

	zobrist := pos.Zobrist()
	pv[zobrist&pvTableMask] = pvEntry{
		lock: zobrist,
		move: move,
	}
}

// TODO: Lookup move in transposition table if none is available.
func (pv pvTable) get(pos *Position) Move {
	zobrist := pos.Zobrist()
	if entry := &pv[zobrist&pvTableMask]; entry.lock == zobrist {
		return entry.move
	}
	return NullMove
}

// Get returns the principal variation.
func (pv pvTable) Get(pos *Position) []Move {
	seen := make(map[uint64]bool)
	var moves []Move

	// Extract the moves by following the position.
	next := pv.get(pos)
	for next != NullMove && !seen[pos.Zobrist()] {
		seen[pos.Zobrist()] = true
		moves = append(moves, next)
		pos.DoMove(next)
		next = pv.get(pos)
	}

	// Undo all moves, so we get back to the initial state.
	for range moves {
		pos.UndoMove()
	}
	return moves
}
