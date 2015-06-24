package engine

import (
	"sync"
	"time"
)

const (
	defaultMovesToGo    = 30 // default number of more moves expected to play
	defaultbranchFactor = 2  // default branching factor
)

// atomicFlag is an atomic bool that can only be set.
type atomicFlag struct {
	lock sync.Mutex
	flag bool
}

func (af *atomicFlag) set() {
	af.lock.Lock()
	af.flag = true
	af.lock.Unlock()
}

func (af *atomicFlag) get() bool {
	af.lock.Lock()
	tmp := af.flag
	af.lock.Unlock()
	return tmp
}

// TimeControl is an interface to control the clock.
type TimeControl interface {
	// Start starts the watch.
	Start()
	// NextDepth returns next depth to run. -1 means stop.
	NextDepth() int
	// Stop stops the search as soon as possible.
	Stop()
	// IsStopped returns true if Stop was called.
	IsStopped() bool
}

// FixedDepthTimeControl searches all depths up to MaxDepth.
type FixedDepthTimeControl struct {
	MaxDepth  int // maximum depth to search to
	currDepth int // current depth
	stopped   atomicFlag
}

func (tc *FixedDepthTimeControl) Start() {
	tc.stopped = atomicFlag{}
	tc.currDepth = -1
}

func (tc *FixedDepthTimeControl) NextDepth() int {
	if tc.IsStopped() {
		return -1
	}
	tc.currDepth++
	if tc.currDepth <= tc.MaxDepth {
		return tc.currDepth
	}
	return -1
}

func (tc *FixedDepthTimeControl) Stop() {
	tc.stopped.set()
}

func (tc *FixedDepthTimeControl) IsStopped() bool {
	return tc.stopped.get()
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

	stopped   atomicFlag // When time is up, stop should be closed.
	currDepth int        // Current depth.
	timeLimit time.Time
}

func (tc *OnClockTimeControl) Start() {
	movesToGo := defaultMovesToGo
	if tc.MovesToGo != 0 {
		movesToGo = tc.MovesToGo
	}

	// Branch more when there are more pieces.
	// With fewer pieces there is less mobility
	// and hash table kicks in more often.
	branchFactor := defaultbranchFactor
	for np := tc.NumPieces; np > 0; np /= 5 {
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

	tc.stopped = atomicFlag{}
	tc.currDepth = 0
	tc.timeLimit = time.Now().Add(thinkTime / time.Duration(branchFactor))
}

func (tc *OnClockTimeControl) NextDepth() int {
	if tc.IsStopped() && tc.currDepth > 0 {
		return -1
	}
	if tc.currDepth < 64 {
		tc.currDepth++
		return tc.currDepth
	}
	return -1
}

func (tc *OnClockTimeControl) Stop() {
	tc.stopped.set()
}

func (tc *OnClockTimeControl) IsStopped() bool {
	return tc.stopped.get() || time.Now().After(tc.timeLimit)
}
