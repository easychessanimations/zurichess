// +build coach

package engine

const disableCache = true

// Score represents a pair of mid and end game scores.
type Score struct {
	M, E int32 // mid game, end game
	I    int   // index in Weights
}

// Eval is a sum of scores.
type Eval struct {
	M, E   int32   // mid game, end game
	Values []int32 // input values
}

func (e *Eval) Feed(phase int32) int32 {
	return (e.M*(256-phase) + e.E*phase) / 256
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

func initWeights() {
	for i := range Weights {
		Weights[i].I = i
	}
}
