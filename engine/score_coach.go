// +build coach

package engine

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
	I    int   // index in Weights
}

// Eval is a sum of scores.
type Eval struct {
	M, E   int32 // mid game, end game
	Phase  int32
	Values []int32
}

func (e *Eval) Make(pos *Position) {
	e.M, e.E = 0, 0
	e.Phase = phase(pos)
	for i := range e.Values {
		e.Values[i] = 0
	}
}

func (e *Eval) Feed() int32 {
	return (e.M*(256-e.Phase) + e.E*e.Phase) / 256
}

func (e *Eval) Recompute() {
	e.M, e.E = 0, 0
	for i := range Weights {
		s, n := Weights[i], e.Values[i]
		e.M += s.M * n
		e.E += s.E * n
	}
}

func (e *Eval) Add(s Score) {
	e.M += s.M
	e.E += s.E
	if e.Values != nil {
		e.Values[s.I] += 1
	}
}

func (e *Eval) AddN(s Score, n int32) {
	e.M += s.M * n
	e.E += s.E * n
	if e.Values != nil {
		e.Values[s.I] += n
	}
}

func (e *Eval) Neg() {
	e.M = -e.M
	e.E = -e.E
	for i, v := range e.Values {
		e.Values[i] = -v
	}
}

func evaluatePawnsCached(pos *Position, us Color, eval *Eval) {
	evaluatePawns(pos, us, eval)
}

func initWeights() {
	for i := range Weights {
		Weights[i].I = i
	}
}
