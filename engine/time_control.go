package engine

import (
	"time"
)

const (
	defaultMovesToGo    = 30 // default number of more moves expected to play
	defaultbranchFactor = 2  // default branching factor
)

type TimeControl interface {
	// Starts starts the watch.
	Start()
	// NextDepth returns next depth to run. -1 means stop.
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
	return -1
}

// OnClockTimeControl is a time control that tries to split the
// remaining time over MovesToGo.
type OnClockTimeControl struct {
	// Number of remaining pieces on the board.
	NumPieces int
	// Remaining time.
	Time time.Duration
	// Time increment after each move.
	Inc time.Duration
	// Number of moves left. Recommended values
	// 0 when there is no time refresh.
	// 1 when solving puzzles.
	// n when there is a time refresh.
	MovesToGo int
	// When time is up, Stop should be closed.
	Stop <-chan struct{}

	// Latest moment when to start next depth.
	timeLimit time.Time
	// Current depth.
	currDepth int
}

func (tc *OnClockTimeControl) Start() {
	movesToGo := defaultMovesToGo
	if tc.MovesToGo != 0 {
		movesToGo = tc.MovesToGo
	}

	// Branch more when there are more pieces.
	// With fewer pieces, hash table kicks in.
	branchFactor := defaultbranchFactor
	for np := tc.NumPieces; np > 0; np /= 4 {
		branchFactor++
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	for i := 4; i > 0; i /= 2 {
		if movesToGo <= i {
			branchFactor++
		}
	}

	// Compute how much time to think according to the formula below.
	// The formula allows engine to use more of time.Left in the begining
	// and rely more on the inc time later.
	thinkTime := (tc.Time + time.Duration(movesToGo-1)*tc.Inc) / time.Duration(movesToGo)
	if thinkTime > tc.Time {
		// Do not allocate more than we have.
		thinkTime = tc.Time
	}

	tc.timeLimit = time.Now().Add(thinkTime / time.Duration(branchFactor))
	tc.currDepth = 0
}

func (tc *OnClockTimeControl) NextDepth() int {
	select {
	case <-tc.Stop: // Stop was requested.
		return -1
	default:
	}

	if tc.currDepth < 64 && (tc.currDepth < 1 || time.Now().Before(tc.timeLimit)) {
		tc.currDepth++
		return tc.currDepth
	}
	return -1
}
