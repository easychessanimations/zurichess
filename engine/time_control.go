package engine

import (
	"math"
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

// TimeControl is a time control that tries to split the
// remaining time over MovesToGo.
type TimeControl struct {
	WTime, WInc time.Duration // time and increment for white.
	BTime, BInc time.Duration // time and increment for black
	Depth       int           // maximum depth search (including)
	MovesToGo   int           // number of remaining moves

	stopped   atomicFlag // When time is up, stop should be closed.
	timeLimit time.Time
}

// NewTimeControl returns a new time control with no time limit,
// no depth limit, zero time increment and one move to go.
func NewTimeControl() *TimeControl {
	inf := time.Duration(math.MaxInt64)
	return &TimeControl{
		WTime:     inf,
		WInc:      0,
		BTime:     inf,
		BInc:      0,
		Depth:     64,
		MovesToGo: 1,
	}
}

func NewFixedDepthTimeControl(depth int) *TimeControl {
	tc := NewTimeControl()
	tc.Depth = depth
	return tc
}

func NewDeadlineTimeControl(deadline time.Duration) *TimeControl {
	tc := NewTimeControl()
	tc.WTime = deadline
	tc.BTime = deadline
	return tc
}

// Start starts the timer.
func (tc *TimeControl) Start(pos *Position) {
	// Branch more when there are more pieces.
	// With fewer pieces there is less mobility
	// and hash table kicks in more often.
	branchFactor := defaultbranchFactor
	for np := pos.NumPieces[NoColor][NoFigure]; np > 0; np /= 5 {
		branchFactor++
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	for i := 4; i > 0; i /= 2 {
		if tc.MovesToGo <= i {
			branchFactor++
		}
	}

	time_, inc := tc.WTime, tc.WInc
	if pos.SideToMove == Black {
		time_, inc = tc.BTime, tc.BInc
	}

	// Compute how much time to think according to the formula below.
	// The formula allows engine to use more of time.Left in the begining
	// and rely more on the inc time later.
	thinkTime := (time_ + time.Duration(tc.MovesToGo-1)*inc) / time.Duration(tc.MovesToGo)
	if thinkTime > time_ {
		// Do not allocate more time than available.
		thinkTime = time_
	}

	tc.stopped = atomicFlag{}
	tc.timeLimit = time.Now().Add(thinkTime / time.Duration(branchFactor))
}

// NextDepth returns true if search can start at depth.
func (tc *TimeControl) NextDepth(depth int) bool {
	// If maximum search is not reached then at least one ply is searched.
	// This avoid an issue when, under the clock, engine doesn't return any move
	// because it stops at depth 0.
	return depth <= tc.Depth && (depth <= 1 || !tc.IsStopped())
}

// Stop marks the search as stopped.
func (tc *TimeControl) Stop() {
	tc.stopped.set()
}

// IsStopped returns true if the search has stopped.
func (tc *TimeControl) IsStopped() bool {
	return tc.stopped.get() || time.Now().After(tc.timeLimit)
}
