// pawn_table.go caches pawn structure evaluation.
//
// Pawn evaluation is slow since there are many pawns times
// many features. However, pawns have restricted mobility so
// the evaluation doesn't change much.
//
// TODO: Cache one side only. Current way makes the code a bit
// ugly which is not worth it considering that the table is very
// small

package engine

const (
	pawnTableBits = 6
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
