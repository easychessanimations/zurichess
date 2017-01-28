// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// move_ordering generates and orders moves for an engine.
// Generation is done in several phases and many times
// actual generation or sorting can be eliminated.

package engine

const (
	// Move generation states.

	msHash          = iota // return hash move
	msGenViolent           // generate violent moves
	msReturnViolent        // return violent moves in order
	msGenKiller            // generate killer moves
	msReturnKiller         // return killer moves  in order
	msGenRest              // generate remaining moves
	msReturnRest           // return remaining moves in order
	msDone                 // all moves returned
)

// mvvlva values based on one pawn = 10.
var mvvlvaBonus = [...]int16{0, 10, 40, 45, 68, 145, 256}

// mvvlva computes Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
func mvvlva(h *historyTable, m Move) int16 {
	if m.IsQuiet() {
		// Sort quiet moves by how well they performed.
		// Start at a very low score (-20000) so it doesn't overlap good/bad captures range.
		return int16(-20000 + h.get(m))
	}

	a := m.Target().Figure()
	v := m.Capture().Figure()
	return mvvlvaBonus[v]*64 - mvvlvaBonus[a]
}

// movesStack is a stack of moves.
type moveStack struct {
	moves []Move  // list of moves
	order []int16 // weight of each move for comparison

	kind   int     // violent or all
	state  int     // current generation state
	hash   Move    // hash move
	killer [3]Move // two killer moves and one counter move
}

// stack is a stack of plies (movesStack).
type stack struct {
	position *Position
	moves    []moveStack
	history  *historyTable
	counter  *[1 << 11]Move // counter moves table
}

// Reset clear the stack for a new position.
func (st *stack) Reset(pos *Position) {
	st.position = pos
	st.moves = st.moves[:0]
}

// get returns the moveStack for current ply.
// allocates memory if necessary.
func (st *stack) get() *moveStack {
	for len(st.moves) <= st.position.Ply {
		st.moves = append(st.moves, moveStack{
			moves: make([]Move, 0, 16),
			order: make([]int16, 0, 16),
		})
	}
	return &st.moves[st.position.Ply]
}

// GenerateMoves generates all moves of kind.
func (st *stack) GenerateMoves(kind int, hash Move) {
	ms := st.get()
	ms.moves = ms.moves[:0] // clear the array, but keep the backing memory
	ms.order = ms.order[:0]
	ms.kind = kind
	ms.state = msHash
	ms.hash = hash
	ms.killer[2] = NullMove
	// ms.killer = ms.killer // keep killers
}

// generateMoves generates all moves.
func (st *stack) generateMoves(kind int) {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) != 0 || len(ms.order) != 0 {
		panic("expected no moves")
	}
	if ms.kind&kind == 0 {
		return
	}
	st.position.GenerateMoves(ms.kind&kind, &ms.moves)
	for _, m := range ms.moves {
		ms.order = append(ms.order, mvvlva(st.history, m))
	}
	st.sort()
}

// Gaps from Best Increments for the Average Case of Shellsort, Marcin Ciura.
var shellSortGaps = [...]int{132, 57, 23, 10, 4, 1}

func (st *stack) sort() {
	ms := &st.moves[st.position.Ply]
	for _, gap := range shellSortGaps {
		for i := gap; i < len(ms.order); i++ {
			j := i
			to, tm := ms.order[j], ms.moves[j]
			for ; j >= gap && ms.order[j-gap] > to; j -= gap {
				ms.order[j] = ms.order[j-gap]
				ms.moves[j] = ms.moves[j-gap]
			}
			ms.order[j], ms.moves[j] = to, tm
		}
	}
}

// popFront pops the move from the front
func (st *stack) popFront() Move {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) == 0 {
		return NullMove
	}

	last := len(ms.moves) - 1
	move := ms.moves[last]
	ms.moves = ms.moves[:last]
	ms.order = ms.order[:last]
	return move
}

// Pop pops a new move.
// Returns NullMove if there are no moves.
// Moves are generated in several phases:
//	first the hash move,
//      then the violent moves,
//      then the killer moves,
//      then the tactical and quiet moves.
func (st *stack) PopMove() Move {
	ms := &st.moves[st.position.Ply]
	for {
		switch ms.state {
		// Return the hash move.
		case msHash:
			// Return the hash move directly without generating the pseudo legal moves.
			ms.state = msGenViolent
			if st.position.IsPseudoLegal(ms.hash) {
				return ms.hash
			}

		// Return the violent moves.
		case msGenViolent:
			ms.state = msReturnViolent
			st.generateMoves(Violent)

		case msReturnViolent:
			if m := st.popFront(); m == NullMove {
				if ms.kind&Quiet == 0 {
					// Skip killers and quiets if only violent moves are searched.
					ms.state = msDone
				} else {
					ms.state = msGenKiller
				}
			} else if m != ms.hash && m != NullMove {
				return m
			}

		// Return two killer moves and one counter move.
		case msGenKiller:
			// ms.moves is a stack so moves are pushed in the reversed order.
			ms.state = msReturnKiller
			cm := st.counter[st.counterIndex()]
			if cm != ms.killer[0] && cm != ms.killer[1] && cm != NullMove {
				ms.killer[2] = cm
				ms.moves = append(ms.moves, cm)
				ms.order = append(ms.order, -2)
			}
			if m := ms.killer[1]; m != NullMove {
				ms.moves = append(ms.moves, m)
				ms.order = append(ms.order, -1)
			}
			if m := ms.killer[0]; m != NullMove {
				ms.moves = append(ms.moves, m)
				ms.order = append(ms.order, 0)
			}

		case msReturnKiller:
			if m := st.popFront(); m == NullMove {
				ms.state = msGenRest
			} else if m != ms.hash && st.position.IsPseudoLegal(m) {
				return m
			}

		// Return remaining quiet and tactical moves.
		case msGenRest:
			ms.state = msReturnRest
			st.generateMoves(Quiet)

		case msReturnRest:
			if m := st.popFront(); m == NullMove {
				ms.state = msDone
			} else if m == ms.hash || st.IsKiller(m) {
				break
			} else {
				return m
			}

		case msDone:
			// Just in case another move is requested.
			return NullMove
		}
	}
}

// IsKiller returns true if m is a killer move for currenty ply.
func (st *stack) IsKiller(m Move) bool {
	ms := &st.moves[st.position.Ply]
	return m == ms.killer[0] || m == ms.killer[1] || m == ms.killer[2]
}

// SaveKiller saves a killer move, m.
func (st *stack) SaveKiller(m Move) {
	ms := &st.moves[st.position.Ply]
	if !m.IsViolent() {
		st.counter[st.counterIndex()] = m
		// Move the newly found killer first.
		if m != ms.killer[0] {
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		}
	}
}

// counterIndex returns the index of the counter move in the counter table.
// The hash is computed based on the last move.
func (st *stack) counterIndex() int {
	pos := st.position
	hash := murmurMix(uint64(pos.LastMove()), murmurSeed[pos.Us()])
	return int(hash % uint64(len(st.counter)))
}
