package engine

const (
	pawnTableBits = 6
	pawnTableSize = 1 << pawnTableBits
)

type pawnEntry struct {
	white Bitboard
	black Bitboard
	score Score
}

type pawnTable [1 << pawnTableBits]pawnEntry

// hash hashes white and black pawns bitboards together.
func hash(white, black Bitboard) int {
	h := (white ^ black) * 4270591956663283
	return int(h >> (64 - pawnTableBits))
}

// get retrieves score from the table, if cached.
func (pt *pawnTable) get(white, black Bitboard) (Score, bool) {
	entry := &pt[hash(white, black)]
	return entry.score, (entry.white == white && entry.black == black)
}

// put stores score in the table.
func (pt *pawnTable) put(white, black Bitboard, score Score) {
	entry := &pt[hash(white, black)]
	entry.white = white
	entry.black = black
	entry.score = score
}
