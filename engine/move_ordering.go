package engine

var (
	// mvvlvaTable stores the ordering scores.
	//
	// MVV/LVA stands for "Most valuable victim, Least valuable attacker".
	// See https://chessprogramming.wikispaces.com/MVV-LVA.
	//
	// In zurichess the MVV/LVA formula is not used,
	// but the values are optimized and stored in this array.
	// Capturing the king should have a very high value
	// to prevent searching positions with other side in check.
	//
	// mvvlvaTable[attacker * FigureSize + victim]
	mvvlvaTable = [FigureArraySize * FigureArraySize]int{
		250, 254, 535, 757, 919, 1283, 20000, // Promotion
		250, 863, 1380, 1779, 2307, 2814, 20000, // Pawn
		250, 781, 1322, 1654, 1766, 2414, 20000, // Knight
		250, 409, 810, 1411, 2170, 3000, 20000, // Bishop
		250, 393, 1062, 1199, 2117, 2988, 20000, // Rook
		250, 349, 948, 1355, 1631, 2314, 20000, // Queen
		250, 928, 1088, 1349, 1593, 2417, 20000, // King
	}

	// Give bonus to move found in the hash table.
	HashMoveBonus int16 = 4096
	// Give bonus to killer move.
	KillerMoveBonus int16 = 1024
)

// SetMvvLva sets the MVV/LVA table.
func SetMvvLva(str string) error {
	return SetMaterialValue("MvvLva", mvvlvaTable[:], str)
}

// mvvlva computes Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
func mvvlva(m Move) int16 {
	a := int(m.Piece().Figure())
	v := int(m.Capture().Figure())
	p := int(m.Promotion().Figure())
	return int16(mvvlvaTable[a*FigureArraySize+v] + mvvlvaTable[p])
}

// heapSort sorts moves by coresponding value in order.
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

// movesStack is a double stack to store moves.
type moveStack struct {
	moves []Move
	order []int16
	heads []int
}

// Stack generates and sorts by weigth the moves for current search ply.
func (ms *moveStack) Stack(generate func(*[]Move), weight func(Move) int16) {
	if len(ms.heads) == 0 {
		ms.heads = append(ms.heads, 0)
	}
	start := len(ms.moves)
	generate(&ms.moves)
	for _, m := range ms.moves[start:] {
		ms.order = append(ms.order, weight(m))
	}
	(&heapSort{ms.moves[start:], ms.order[start:]}).sort()
	ms.heads = append(ms.heads, len(ms.moves))
}

// PopMove pops a single move from the stack.
// Returns true if such move exists at current ply.
// If current ply has no moves remainig pops the ply too.
func (ms *moveStack) PopMove(move *Move) bool {
	last := len(ms.heads) - 1
	if ms.heads[last] == ms.heads[last-1] {
		ms.PopAll()
		return false
	}

	ms.heads[last]--
	head := ms.heads[last]
	*move = ms.moves[head]
	ms.moves = ms.moves[:head]
	ms.order = ms.order[:head]
	return true
}

// PopAll pops current ply.
func (ms *moveStack) PopAll() {
	last := len(ms.heads) - 1
	head := ms.heads[last-1]
	ms.moves = ms.moves[:head]
	ms.order = ms.order[:head]
	ms.heads = ms.heads[:last]
}
