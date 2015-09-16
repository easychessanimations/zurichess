package engine

import (
	"sync"
	"time"
)

const (
	defaultMovesToGo = 30 // default number of more moves expected to play
	infinite         = 1000000000 * time.Second
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
	Depth       int32         // maximum depth search (including)
	MovesToGo   int           // number of remaining moves

	sideToMove Color
	branch     int        // branching factor
	currDepth  int32      // current depth searched
	stopped    atomicFlag // true to stop the search
	ponderhit  atomicFlag // true if ponder was successful

	searchTime     time.Duration // alocated time for this move
	searchDeadline time.Time     // don't go to the next depth after this deadline
	stopDeadline   time.Time     // abort search after this deadline
}

// NewTimeControl returns a new time control with no time limit,
// no depth limit, zero time increment and zero moves to go.
func NewTimeControl(pos *Position) *TimeControl {
	// Branch more when there are more pieces. With fewer pieces
	// there is less mobility and hash table kicks in more often.
	branch := 2
	for np := (pos.ByColor[White] | pos.ByColor[Black]).Popcnt(); np > 0; np /= 6 {
		branch++
	}

	return &TimeControl{
		WTime:      infinite,
		WInc:       0,
		BTime:      infinite,
		BInc:       0,
		Depth:      64,
		MovesToGo:  defaultMovesToGo,
		sideToMove: pos.SideToMove,
		branch:     branch,
	}
}

func NewFixedDepthTimeControl(pos *Position, depth int32) *TimeControl {
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
	tt := (t + (tmp-1)*i) / tmp

	if tt < 0 {
		return 0
	}
	if tt < t {
		return tt
	}
	return t
}

// Start starts the timer.
// Should start as soon as possible to set the correct time.
func (tc *TimeControl) Start(ponder bool) {
	var otime, oinc time.Duration // our time, inc
	if tc.sideToMove == White {
		otime, oinc = tc.WTime, tc.WInc
	} else {
		otime, oinc = tc.BTime, tc.BInc
	}

	// Increase the branchFactor a bit to be on the
	// safe side when there are only a few moves left.
	for i := 4; i > 0; i /= 2 {
		if tc.MovesToGo <= i {
			tc.branch++
		}
	}

	tc.stopped = atomicFlag{flag: false}
	tc.ponderhit = atomicFlag{flag: !ponder}

	// searchDeadline is the last moment when search can start a new iteration.
	// stopDeadline is when to abort the search in case of an explosion.
	now := time.Now()
	tc.searchTime = tc.thinkingTime(otime, oinc)
	tc.searchDeadline = now.Add(tc.searchTime / time.Duration(tc.branch))
	tc.stopDeadline = now.Add(tc.searchTime * 4)
}

// NextDepth returns true if search can start at depth.
func (tc *TimeControl) NextDepth(depth int32) bool {
	tc.currDepth = depth
	return tc.currDepth <= tc.Depth && !tc.hasStopped(tc.searchDeadline)
}

// PonderHit switch to our time control.
func (tc *TimeControl) PonderHit() {
	now := time.Now()
	tc.searchDeadline = now.Add(tc.searchTime / time.Duration(tc.branch))
	tc.stopDeadline = now.Add(tc.searchTime * 4)
	tc.ponderhit.set()
}

// Stop marks the search as stopped.
func (tc *TimeControl) Stop() {
	tc.stopped.set()
}

func (tc *TimeControl) hasStopped(deadline time.Time) bool {
	if tc.currDepth <= 2 {
		return false
	}
	if tc.stopped.get() {
		return true
	}
	if tc.ponderhit.get() && time.Now().After(deadline) {
		return true
	}
	return false
}

// Stopped returns true if the search has stopped.
func (tc *TimeControl) Stopped() bool {
	if !tc.hasStopped(tc.stopDeadline) {
		return false
	}
	tc.stopped.set()
	return true
}
