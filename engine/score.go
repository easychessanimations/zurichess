// +build !coach

package engine

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
}

// Eval is a sum of scores.
type Eval struct {
	M, E  int32 // mid game, end game
	Phase int32
}

func (e *Eval) Make(pos *Position) {
	e.M, e.E = 0, 0
	e.Phase = phase(pos)
}

func (e *Eval) Feed() int32 {
	return (e.M*(256-e.Phase) + e.E*e.Phase) / 256
}

func (e *Eval) Add(s Score) {
	e.M += s.M
	e.E += s.E
}

func (e *Eval) AddN(s Score, n int32) {
	e.M += s.M * n
	e.E += s.E * n
}

func (e *Eval) Neg() {
	e.M = -e.M
	e.E = -e.E
}

var pawnsCache [ColorArraySize]pawnTable

func evaluatePawnsCached(pos *Position, us Color, eval *Eval) {
	// Use a cached value if available.
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)
	if e, ok := pawnsCache[us].get(ours, theirs); ok {
		eval.M += e.M
		eval.E += e.E
		return
	}

	var e Eval
	evaluatePawns(pos, us, &e)
	pawnsCache[us].put(ours, theirs, e)
	eval.M += e.M
	eval.E += e.E
}

func initWeights() {
}
