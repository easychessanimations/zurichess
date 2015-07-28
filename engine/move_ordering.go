// move_ordering generates and orders moves for an engine.
// Generation is done in several phases and many times
// actual generation or sorting can be eliminated.

package engine

var (
	// MVVLVATable stores the ordering scores.
	//
	// MVV/LVA stands for "Most valuable victim, Least valuable attacker".
	// See https://chessprogramming.wikispaces.com/MVV-LVA.
	//
	// In zurichess the MVV/LVA formula is not used,
	// but the values are optimized and stored in this array.
	// Capturing the king should have a very high value
	// to prevent searching positions with other side in check.
	//
	// MVVLVATable[attacker * FigureSize + victim]
	MVVLVATable = [FigureArraySize * FigureArraySize]int{
		250, 254, 535, 757, 919, 1283, 20000, // Promotion
		250, 863, 1380, 1779, 2307, 2814, 20000, // Pawn
		250, 781, 1322, 1654, 1766, 2414, 20000, // Knight
		250, 409, 810, 1411, 2170, 3000, 20000, // Bishop
		250, 393, 1062, 1199, 2117, 2988, 20000, // Rook
		250, 349, 948, 1355, 1631, 2314, 20000, // Queen
		250, 928, 1088, 1349, 1593, 2417, 20000, // King
	}
)

const (
	// Give bonus to killer move.
	KillerMoveBonus int16 = 1024
)

const (
	// Move generation states.

	msHash = iota // return hash move

	// Generate violent and return one by one in order.
	msGenViolent    // generate moves
	msReturnViolent // return best moves in order

	// Generate tactical&quiet, return best, sort, return one by one.
	msGenRest    // generate moves
	msBestRest   // return best
	msSortRest   // sort
	msReturnRest // return in order

	msDone // all moves returned
)

// mvvlva computes Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
func mvvlva(m Move) int16 {
	a := int(m.Piece().Figure())
	v := int(m.Capture().Figure())
	p := int(m.Promotion().Figure())
	return int16(MVVLVATable[a*FigureArraySize+v] + MVVLVATable[p])
}

// heapSort sorts moves by coresponding value in order.
// heapSort is much faster than the library sort because it
// avoids interface calls.
type heapSort struct {
	moves []Move
	order []int16
}

func (hs *heapSort) swap(i, j int) {
	hs.moves[i], hs.moves[j] = hs.moves[j], hs.moves[i]
	hs.order[i], hs.order[j] = hs.order[j], hs.order[i]
}

func (hs *heapSort) sort() {
	hs.heapify()
	for end := len(hs.moves) - 1; end > 0; {
		hs.swap(end, 0)
		end--
		hs.siftDown(0, end)
	}
}

func (hs *heapSort) heapify() {
	count := len(hs.moves)
	for start := (count - 2) / 2; start >= 0; start-- {
		hs.siftDown(start, count-1)
	}
}

func (hs *heapSort) siftDown(start, end int) {
	for root := start; root*2+1 <= end; {
		swap, child := root, root*2+1
		if hs.order[swap] < hs.order[child] {
			swap = child
		}
		if child+1 <= end && hs.order[swap] < hs.order[child+1] {
			swap = child + 1
		}
		if swap == root {
			return
		}
		hs.swap(root, swap)
		root = swap
	}
}

// movesStack is a stack of moves.
type moveStack struct {
	moves []Move
	order []int16

	kind   int     // violent or all
	state  int     // current generation state
	hash   Move    // hash move
	killer [4]Move // killer moves
}

// stack is a stack of plies (movesStack).
type stack struct {
	position *Position
	moves    []moveStack
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
		st.moves = append(st.moves, moveStack{})
	}
	return &st.moves[st.position.Ply]
}

// generateMoves generates all moves.
func (st *stack) GenerateMoves(kind int, hash Move) {
	ms := st.get()
	ms.moves = ms.moves[:0]
	ms.order = ms.order[:0]
	ms.kind = kind
	ms.state = msHash
	ms.hash = hash
	// ms.killer = ms.killer // keep killers
}

// generateMoves generates all moves.
// Called from main search tree which has hash and killer moves available.
func (st *stack) generateMoves(kind int) {
	ms := &st.moves[st.position.Ply]
	if len(ms.moves) != 0 || len(ms.order) != 0 {
		panic("expected no moves")
	}
	if ms.kind&kind == 0 {
		return
	}

	// Awards bonus for hash and killer moves.
	st.position.GenerateMoves(ms.kind&kind, &ms.moves)
	for _, m := range ms.moves {
		var weight int16
		if m == ms.killer[0] {
			weight = KillerMoveBonus - 0
		} else if m == ms.killer[1] {
			weight = KillerMoveBonus - 1
		} else if m == ms.killer[2] {
			weight = KillerMoveBonus - 2
		} else if m == ms.killer[3] {
			weight = KillerMoveBonus - 3
		} else {
			weight = mvvlva(m)
		}
		ms.order = append(ms.order, weight)
	}
}

// moveBest moves best move to front.
func (st *stack) moveBest() {
	ms := &st.moves[st.position.Ply]
	bi := -1
	for i, m := range ms.moves {
		if m != ms.hash && (bi == -1 || ms.order[i] > ms.order[bi]) {
			bi = i
		}
	}

	if bi != -1 {
		last := len(ms.moves) - 1
		ms.moves[bi], ms.moves[last] = ms.moves[last], ms.moves[bi]
		ms.order[bi], ms.order[last] = ms.order[last], ms.order[bi]
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

	if move == ms.hash {
		// If the front move is the hash move, then the try next move.
		return st.popFront()
	}
	return move
}

// Pop pops a new move.
// Returns NullMove if there are no moves.
// Moves are generated in several phases:
//	first the hash move,
//      then the violent moves,
//      then the tactical and quiet moves.
func (st *stack) PopMove() Move {
	ms := &st.moves[st.position.Ply]
	for {
		switch ms.state {
		// Return the hash move.
		case msHash:
			// Return the hash move directly without generating the pseudo legal moves.
			ms.state = msGenViolent
			if ms.hash != NullMove {
				// TODO verify integrity
				return ms.hash
			}

		// Return the violent moves.
		case msGenViolent:
			ms.state = msReturnViolent
			st.generateMoves(Violent)

		case msReturnViolent:
			// Most positions have only very violent moves so
			// it doesn't make sense to sort given that captures have a high
			// chance to fail high. We just pop the moves in order of score.
			st.moveBest()
			if m := st.popFront(); m != NullMove {
				return m
			}
			if ms.kind&(Tactical|Quiet) == 0 {
				// Optimization: skip remaining steps if no Tactical or Quiet moves
				// were requested (e.g. in quiescence search).
				ms.state = msDone
			} else {
				ms.state = msGenRest
			}

		// Return the quiet and tactical moves.
		case msGenRest:
			ms.state = msBestRest
			st.generateMoves(Tactical | Quiet)

		case msBestRest:
			ms.state = msSortRest
			st.moveBest()
			if m := st.popFront(); m != NullMove {
				return m
			}

		case msSortRest:
			ms.state = msReturnRest
			hs := &heapSort{ms.moves, ms.order}
			hs.sort()

		case msReturnRest:
			if m := st.popFront(); m != NullMove {
				return m
			}
			// Update the state only when there are no moves left.
			ms.state = msDone

		case msDone:
			// Just in case another move is requested.
			return NullMove
		}

	}
}

// HasKiller returns true if there is a killer at this ply.
func (st *stack) HasKiller() bool {
	if st.position.Ply < len(st.moves) {
		ms := &st.moves[st.position.Ply]
		return ms.killer[0] != NullMove
	}
	return false
}

// IsKiller returns true if m is a killer move for currenty ply.
func (st *stack) IsKiller(m Move) bool {
	ms := &st.moves[st.position.Ply]
	return m == ms.killer[0] || m == ms.killer[1] || m == ms.killer[2] || m == ms.killer[3]
}

// SaveKiller saves a killer move, m.
func (st *stack) SaveKiller(m Move) {
	ms := &st.moves[st.position.Ply]
	if m.Capture() == NoPiece {
		// Move the newly found killer first.
		if m == ms.killer[0] {
			// do nothing
		} else if m == ms.killer[1] {
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		} else if m == ms.killer[2] {
			ms.killer[2] = ms.killer[1]
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		} else {
			ms.killer[3] = ms.killer[2]
			ms.killer[2] = ms.killer[1]
			ms.killer[1] = ms.killer[0]
			ms.killer[0] = m
		}
	}
}
