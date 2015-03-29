package engine

const (
	pvTableSizeShift = 13
	pvTableSize      = 1 << pvTableSizeShift
	pvTableMask      = pvTableSize - 1
)

type pvEntry struct {
	// Lock is used to handled hash conflicts.
	// Normally set to position's Zobrist key.
	Lock uint64
	// When was the move added.
	Birth uint32
	// Move on pricipal variation for this position.
	Move Move
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
type pvTable struct {
	table []pvEntry
	timer uint32
}

// newPvTable returns a new pvTable.
func newPvTable() pvTable {
	return pvTable{
		table: make([]pvEntry, pvTableSize),
		timer: 0,
	}
}

// Put inserts a new entry.
func (pv *pvTable) Put(pos *Position, move Move) {
	// Based on pos.Zobrist() two entries are looked up.
	// If any of the two entries in the table matches
	// current position, then that one is replaced.
	// Otherwise, the older is replaced.

	entry1 := &pv.table[uint32(pos.Zobrist())&pvTableMask]
	entry2 := &pv.table[uint32(pos.Zobrist()>>32)&pvTableMask]

	var entry *pvEntry
	if entry1.Lock == pos.Zobrist() {
		entry = entry1
	} else if entry2.Lock == pos.Zobrist() {
		entry = entry2
	} else if entry1.Birth <= entry2.Birth {
		entry = entry1
	} else {
		entry = entry2
	}

	pv.timer++
	*entry = pvEntry{
		Lock:  pos.Zobrist(),
		Move:  move,
		Birth: pv.timer,
	}
}

func (pv *pvTable) get(pos *Position) *pvEntry {
	entry1 := &pv.table[uint32(pos.Zobrist())&pvTableMask]
	entry2 := &pv.table[uint32(pos.Zobrist()>>32)&pvTableMask]

	var entry *pvEntry
	if entry1.Lock == pos.Zobrist() {
		entry = entry1
	}
	if entry2.Lock == pos.Zobrist() {
		entry = entry2
	}
	if entry == nil {
		return nil
	}

	return entry
}

// Get returns the principal variation.
func (pv *pvTable) Get(pos *Position) []Move {
	seen := make(map[uint64]bool)
	var moves []Move

	// Extract the moves by following the position.
	entry := pv.get(pos)
	for entry != nil && entry.Move.MoveType != NoMove && !seen[pos.Zobrist()] {
		seen[pos.Zobrist()] = true
		moves = append(moves, entry.Move)
		pos.DoMove(entry.Move)
		entry = pv.get(pos)
	}

	// Undo all moves, so we get back to the initial state.
	for i := len(moves) - 1; i >= 0; i-- {
		pos.UndoMove(moves[i])
	}
	return moves
}
