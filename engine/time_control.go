package engine

import (
	"time"
)

const (
	defaultMovesToGo    = 30 // default number of more moves expected to play
	defaultbranchFactor = 10 // default branching factor
)

type TimeControl interface {
	// Starts starts the watch.
	Start()
	// NextDepth returns next depth to run. 0 means stop.
	NextDepth() int
}

// FixedDepthTimeControl searches all depths from MinDepth to MaxDepth.
type FixedDepthTimeControl struct {
	MinDepth int
	MaxDepth int

	currDepth int
}

func (tc *FixedDepthTimeControl) Start() {
	tc.currDepth = tc.MinDepth - 1
}

func (tc *FixedDepthTimeControl) NextDepth() int {
	tc.currDepth++
	if tc.currDepth <= tc.MaxDepth {
		return tc.currDepth
	}
	return 0
}

// OnClockTimeControl is a time control that tries to split the
// remaining time over MovesToGo.
type OnClockTimeControl struct {
	// Remaining time.
	Time time.Duration
	// Time increment after each move.
	Inc time.Duration
	// Number of moves left. Recommended values
	// 0 when there is no time refresh.
	// 1 when solving puzzels.
	// n when there is a time refresh.
	MovesToGo int

	// Latest moment when to start next depth.
	timeLimit time.Time
	// Current depth.
	currDepth int
}

func (tc *OnClockTimeControl) Start() {
	movesToGo := time.Duration(defaultMovesToGo)
	if tc.MovesToGo != 0 {
		movesToGo = time.Duration(tc.MovesToGo)
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	branchFactor := time.Duration(defaultbranchFactor)
	if movesToGo < 16 {
		branchFactor += 1
	}
	if movesToGo < 8 {
		branchFactor += 2
	}
	if movesToGo < 4 {
		branchFactor += 3
	}

	// Compute how much time to think according to the formula below.
	// The formula allows engine to use more of time.Left in the begining
	// and rely more on the inc time later.
	thinkTime := (tc.Time + (movesToGo-1)*tc.Inc) / movesToGo
	tc.timeLimit = time.Now().Add(thinkTime / defaultbranchFactor)
	tc.currDepth = 0
}

func (tc *OnClockTimeControl) NextDepth() int {
	if tc.currDepth < 64 && time.Now().Before(tc.timeLimit) {
		tc.currDepth++
		return tc.currDepth
	}
	return 0
}
