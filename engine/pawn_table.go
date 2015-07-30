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
	ours   Bitboard // ours's pawns
	theirs Bitboard // theirs's pawns
	score  Score
}

// hash table for pawn evaluation for a single color.
type pawnTable [1 << pawnTableBits]pawnEntry

// hash hashes ours and theirs pawns bitboards together.
func hash(ours, theirs Bitboard) int {
	h := (ours ^ theirs) * 4270591956663283
	return int(h >> (64 - pawnTableBits))
}

// get retrieves score from the table, if cached.
func (pt *pawnTable) get(ours, theirs Bitboard) (Score, bool) {
	entry := &pt[hash(ours, theirs)]
	return entry.score, (entry.ours == ours && entry.theirs == theirs)
}

// put stores score in the table.
func (pt *pawnTable) put(ours, theirs Bitboard, score Score) {
	entry := &pt[hash(ours, theirs)]
	entry.ours = ours
	entry.theirs = theirs
	entry.score = score
}
