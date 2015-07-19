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

	numPieces  int
	sideToMove Color
	stopped    atomicFlag // true to stop the search
	ponderhit  atomicFlag // true if ponder was successful

	searchTime     time.Duration
	searchDeadline time.Time
	ponderTime     time.Duration
	ponderDeadline time.Time
}

// NewTimeControl returns a new time control with no time limit,
// no depth limit, zero time increment and zero moves to go.
func NewTimeControl(pos *Position) *TimeControl {
	inf := time.Duration(math.MaxInt64)
	return &TimeControl{
		WTime:      inf,
		WInc:       0,
		BTime:      inf,
		BInc:       0,
		Depth:      64,
		MovesToGo:  defaultMovesToGo,
		numPieces:  int((pos.ByColor[White] | pos.ByColor[Black]).Popcnt()),
		sideToMove: pos.SideToMove,
	}
}

func NewFixedDepthTimeControl(pos *Position, depth int) *TimeControl {
	tc := NewTimeControl(pos)
	tc.Depth = depth
	tc.MovesToGo = 1
	return tc
}

func NewDeadlineTimeControl(pos *Position, deadline time.Duration) *TimeControl {
	tc := NewTimeControl(pos)
	tc.WTime = deadline
	tc.BTime = deadline
	tc.MovesToGo = 1
	return tc
}

// thinkingTime calculates how much time to think this round.
// t is the remaining time, i is the increment.
func (tc *TimeControl) thinkingTime(t, i time.Duration) time.Duration {
	// The formula allows engine to use more of time in the begining
	// and rely more on the increment later.
	tmp := time.Duration(tc.MovesToGo)
	if tt := (t + (tmp-1)*i) / tmp; tt < t {
		return tt
	}
	return t
}

// Start starts the timer.
// Should start as soon as possible to set the correct time.
func (tc *TimeControl) Start(ponder bool) {
	// Branch more when there are more pieces. With fewer pieces
	// there is less mobility and hash table kicks in more often.
	branchFactor := time.Duration(defaultbranchFactor)
	for np := tc.numPieces - 2; np > 0; np /= 6 {
		branchFactor++
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	for i := 4; i > 0; i /= 2 {
		if tc.MovesToGo <= i {
			branchFactor++
		}
	}

	var otime, oinc time.Duration // our time, inc
	var ttime, tinc time.Duration // their time, inc
	if tc.sideToMove == White {
		otime, oinc = tc.WTime, tc.WInc
		ttime, tinc = tc.BTime, tc.BInc
	} else {
		otime, oinc = tc.BTime, tc.BInc
		ttime, tinc = tc.WTime, tc.WInc
	}

	tc.stopped = atomicFlag{}
	tc.ponderhit = atomicFlag{flag: !ponder}

	// Searches stops such that the last ply has enough time to finish before alloted time.
	tc.searchTime = tc.thinkingTime(otime, oinc) / branchFactor
	// Pondering stops based on other's time plus some of our time.
	tc.ponderTime = (tc.thinkingTime(ttime, tinc) + tc.searchTime/2) / branchFactor

	now := time.Now()
	tc.ponderDeadline = now.Add(tc.ponderTime)
	tc.searchDeadline = now.Add(tc.searchTime)
}

// NextDepth returns true if search can start at depth.
func (tc *TimeControl) NextDepth(depth int) bool {
	// If maximum search is not reached then at least some plies is searched.
	// This avoid an issue when under the clock engine does not return any move
	// because it stops at depth 0.
	// We also want to stop the search early for `go depth 0`.
	return depth <= tc.Depth && (depth <= 2 || !tc.Stopped())
}

// PonderHit switch to our time control.
func (tc *TimeControl) PonderHit() {
	tc.searchDeadline = time.Now().Add(tc.searchTime)
	tc.ponderhit.set()
}

// Aborted returns true if pondering was aborted.
func (tc *TimeControl) Aborted() bool {
	// tc.ponderhit.get() is true if the engine is currently thinking on its own time.
	return !tc.ponderhit.get() && tc.stopped.get()
}

// Stop marks the search as stopped.
// The result of the search is going to be used.
func (tc *TimeControl) Stop() {
	tc.stopped.set()
}

// Stopped returns true if the search has stopped.
func (tc *TimeControl) Stopped() bool {
	if tc.stopped.get() {
		return true
	}
	if tc.ponderhit.get() && time.Now().After(tc.searchDeadline) {
		tc.stopped.set()
		return true
	}
	if !tc.ponderhit.get() && time.Now().After(tc.ponderDeadline) {
		tc.stopped.set()
		return true
	}
	return false
}
