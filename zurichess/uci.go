// Copyright 2014-2016 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// uci implements the UCI protocol which is described here http://wbec-ridderkerk.nl/html/UCIProtocol.html.
// There is a hidden command, setvalue, which can be used to set the material values.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "bitbucket.org/zurichess/board"
	. "bitbucket.org/zurichess/zurichess/engine"
)

var errQuit = errors.New("quit")

const (
	maxMultiPV       = 16
	maxHandicapLevel = 20
)

// uciLogger outputs search in uci format.
type uciLogger struct {
	start time.Time
	buf   *bytes.Buffer
}

func newUCILogger() *uciLogger {
	return &uciLogger{buf: &bytes.Buffer{}}
}

func (ul *uciLogger) BeginSearch() {
	ul.start = time.Now()
	ul.buf.Reset()
}

func (ul *uciLogger) EndSearch() {
	ul.flush()
}

func (ul *uciLogger) PrintPV(stats Stats, multiPV int, score int32, pv []Move) {
	// Write depth.
	now := time.Now()
	fmt.Fprintf(ul.buf, "info depth %d seldepth %d multipv %d ", stats.Depth, stats.SelDepth, multiPV)

	// Write score.
	if score > KnownWinScore {
		fmt.Fprintf(ul.buf, "score mate %d ", (MateScore-score+1)/2)
	} else if score < KnownLossScore {
		fmt.Fprintf(ul.buf, "score mate %d ", (MatedScore-score)/2)
	} else {
		fmt.Fprintf(ul.buf, "score cp %d ", score)
	}

	// Write stats.
	elapsed := uint64(maxDuration(now.Sub(ul.start), time.Microsecond))
	nps := stats.Nodes * uint64(time.Second) / elapsed
	millis := elapsed / uint64(time.Millisecond)
	fmt.Fprintf(ul.buf, "nodes %d time %d nps %d ", stats.Nodes, millis, nps)

	// Write principal variation.
	fmt.Fprintf(ul.buf, "pv")
	for _, m := range pv {
		fmt.Fprintf(ul.buf, " %v", m.UCI())
	}
	fmt.Fprintf(ul.buf, "\n")

	ul.flush()
}

func (ul *uciLogger) CurrMove(depth int, move Move, num int) {
	if depth > 15 && time.Now().Sub(ul.start) > 10*time.Second {
		fmt.Fprintf(ul.buf, "info depth %d currmove %v currmovenumber %d\n", depth, move.UCI(), num)
		ul.flush()
	}
}

// flush flushes the buf to stdout.
func (ul *uciLogger) flush() {
	os.Stdout.Write(ul.buf.Bytes())
	os.Stdout.Sync()
	ul.buf.Reset()
}

// maxDuration returns maximum of a and b.
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// UCI implements uci protocol.
type UCI struct {
	Engine      *Engine
	timeControl *TimeControl

	// buffer of 1, if empty then the engine is available
	idle chan struct{}
	// buffer of 1, if filled then the engine is pondering
	ponder chan struct{}
	// predicted position hash after 2 moves.
	predicted uint64
	// root moves to search; empty to search all of them.
	rootMoves []Move
}

func NewUCI() *UCI {
	options := Options{}
	return &UCI{
		Engine:      NewEngine(nil, newUCILogger(), options),
		timeControl: nil,
		idle:        make(chan struct{}, 1),
		ponder:      make(chan struct{}, 1),
	}
}

var reCmd = regexp.MustCompile(`^[[:word:]]+\b`)

func (uci *UCI) Execute(line string) error {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	cmd := reCmd.FindString(line)
	if cmd == "" {
		return fmt.Errorf("invalid command line")
	}

	// These commands do not expect the engine to be idle.
	switch cmd {
	case "isready":
		return uci.isready(line)
	case "quit":
		return errQuit
	case "stop":
		return uci.stop(line)
	case "uci":
		return uci.uci(line)
	case "ponderhit":
		return uci.ponderhit(line)
	}

	// Make sure that the engine is idle.
	uci.idle <- struct{}{}
	<-uci.idle

	// These commands expect engine to be idle.
	switch cmd {
	case "ucinewgame":
		return uci.ucinewgame(line)
	case "position":
		return uci.position(line)
	case "go":
		return uci.go_(line)
	case "setoption":
		return uci.setoption(line)
	default:
		return fmt.Errorf("unhandled command %s", cmd)
	}
}

func (uci *UCI) uci(line string) error {
	fmt.Printf("id name zurichess %v\n", buildVersion)
	fmt.Printf("id author Alexandru Moșoi\n")
	fmt.Printf("\n")
	fmt.Printf("option name Hash type spin default %v min 1 max 65536\n", DefaultHashTableSizeMB)
	fmt.Printf("option name MultiPV type spin default %d min 1 max %d\n", uci.Engine.Options.MultiPV, maxMultiPV)
	fmt.Printf("option name Ponder type check default true\n")
	fmt.Printf("option name Handicap Level type spin default %d min 0 max %d\n", uci.Engine.Options.HandicapLevel, maxHandicapLevel)
	fmt.Printf("option name UCI_AnalyseMode type check default false\n")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(line string) error {
	fmt.Println("readyok")
	return nil
}

func (uci *UCI) ucinewgame(line string) error {
	// Clear the hash at the beginning of each game.
	GlobalHashTable.Clear()
	return nil
}

func (uci *UCI) position(line string) error {
	args := strings.Fields(line)[1:]
	if len(args) == 0 {
		return fmt.Errorf("expected argument for 'position'")
	}

	var pos *Position

	i := 0
	var err error
	switch args[i] {
	case "startpos":
		pos, err = PositionFromFEN(FENStartPos)
		i++
	case "fen":
		for i < len(args) && args[i] != "moves" {
			i++
		}
		pos, err = PositionFromFEN(strings.Join(args[1:i], " "))
	default:
		err = fmt.Errorf("unknown position command: %s", args[0])
	}
	if err != nil {
		return err
	}

	uci.Engine.SetPosition(pos)

	if i < len(args) {
		if args[i] != "moves" {
			return fmt.Errorf("expected 'moves', got '%s'", args[1])
		}
		for _, m := range args[i+1:] {
			if move, err := uci.Engine.Position.UCIToMove(m); err != nil {
				return err
			} else {
				uci.Engine.DoMove(move)
			}
		}
	}

	return nil
}

var validGoCommands = map[string]bool{
	"searchmoves": true,
	"ponder":      true,
	"wtime":       true,
	"btime":       true,
	"binc":        true,
	"movestogo":   true,
	"depth":       true,
	"nodes":       true,
	"mate":        true,
	"movetime":    true,
	"infinite":    true,
}

func (uci *UCI) go_(line string) error {
	// TODO: Handle panic for `go depth`
	predicted := uci.predicted == uci.Engine.Position.Zobrist()
	uci.timeControl = NewTimeControl(uci.Engine.Position, predicted)
	uci.rootMoves = uci.rootMoves[:0]
	ponder := false

	args := strings.Fields(line)[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "searchmoves":
			for j := i + 1; j < len(args) && !validGoCommands[args[j]]; j++ {
				if m, err := uci.Engine.Position.UCIToMove(args[j]); err != nil {
					return err
				} else {
					i++
					uci.rootMoves = append(uci.rootMoves, m)
				}
			}
		case "ponder":
			ponder = true
		case "infinite":
			uci.timeControl = NewTimeControl(uci.Engine.Position, false)
		case "wtime":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.WTime = time.Duration(t) * time.Millisecond
		case "winc":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.WInc = time.Duration(t) * time.Millisecond
		case "btime":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.BTime = time.Duration(t) * time.Millisecond
		case "binc":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.BInc = time.Duration(t) * time.Millisecond
		case "movestogo":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.MovesToGo = int32(t)
		case "movetime":
			i++
			t, _ := strconv.Atoi(args[i])
			uci.timeControl.WTime = time.Duration(t) * time.Millisecond
			uci.timeControl.WInc = 0
			uci.timeControl.BTime = time.Duration(t) * time.Millisecond
			uci.timeControl.BInc = 0
			uci.timeControl.MovesToGo = 1
		case "depth":
			i++
			d, _ := strconv.Atoi(args[i])
			uci.timeControl.Depth = int32(d)
		case "nodes", "mate":
			log.Println(args[i], "not implemented. Ignoring")
			i++
		default:
			return fmt.Errorf("invalid go command %s", args[i])
		}
	}

	if ponder {
		// Ponder was requested, so fill the channel.
		// Next write to uci.ponder will block.
		uci.ponder <- struct{}{}
	}

	uci.timeControl.Start(ponder)
	uci.idle <- struct{}{}
	go uci.play()
	return nil
}

func (uci *UCI) ponderhit(line string) error {
	uci.timeControl.PonderHit()
	<-uci.ponder
	return nil
}

func (uci *UCI) stop(line string) error {
	// Stop the timer if not already stopped.
	if uci.timeControl != nil {
		uci.timeControl.Stop()
	}
	// No longer pondering.
	select {
	case <-uci.ponder:
	default:
	}
	// Waits until the engine becomes idle.
	uci.idle <- struct{}{}
	<-uci.idle

	return nil
}

// play starts the negine.
// Should run in its own separate goroutine.
func (uci *UCI) play() {
	_, moves := uci.Engine.PlayMoves(uci.timeControl, uci.rootMoves)

	if len(moves) >= 2 {
		uci.Engine.Position.DoMove(moves[0])
		uci.Engine.Position.DoMove(moves[1])
		uci.predicted = uci.Engine.Position.Zobrist()
		uci.Engine.Position.UndoMove()
		uci.Engine.Position.UndoMove()
	} else {
		uci.predicted = uci.Engine.Position.Zobrist()
	}

	// If pondering was requested it will block because the channel is full.
	uci.ponder <- struct{}{}
	<-uci.ponder

	if len(moves) == 0 {
		fmt.Printf("bestmove (none)\n")
	} else if len(moves) == 1 {
		fmt.Printf("bestmove %v\n", moves[0].UCI())
	} else {
		fmt.Printf("bestmove %v ponder %v\n", moves[0].UCI(), moves[1].UCI())
	}

	// Marks the engine as idle.
	// If the engine is made idle before best move is shown
	// then sometimes (at very high rate of commands position / go)
	// there is a race info / bestmove lines are intermixed wrongly.
	// This confuses the tuner, at least.
	<-uci.idle
}

var reOption = regexp.MustCompile(`^setoption\s+name\s+(.+?)(\s+value\s+(.*))?$`)

func (uci *UCI) setoption(line string) error {
	option := reOption.FindStringSubmatch(line)
	if option == nil {
		return fmt.Errorf("invalid setoption arguments")
	}

	// Handle buttons which don't have a value.
	if len(option) < 1 {
		return fmt.Errorf("missing setoption name")
	}
	switch option[1] {
	case "Clear Hash":
		GlobalHashTable.Clear()
		return nil
	}

	// Handle remaining values.
	if len(option) < 3 {
		return fmt.Errorf("missing setoption value")
	}
	switch option[1] {
	case "UCI_AnalyseMode":
		if mode, err := strconv.ParseBool(option[3]); err != nil {
			return err
		} else {
			uci.Engine.Options.AnalyseMode = mode
		}
		return nil
	case "Hash":
		if hashSizeMB, err := strconv.ParseInt(option[3], 10, 64); err != nil {
			return err
		} else {
			GlobalHashTable = NewHashTable(int(hashSizeMB))
		}
		return nil
	case "MultiPV":
		if multiPV, err := strconv.ParseInt(option[3], 10, 64); err != nil {
			return err
		} else if 1 <= multiPV && multiPV <= maxMultiPV {
			uci.Engine.Options.MultiPV = int(multiPV)
		} else {
			return fmt.Errorf("MultiPV must be between 1 and %d", maxMultiPV)
		}
		return nil
	case "Handicap Level":
		if handicap, err := strconv.ParseInt(option[3], 10, 64); err != nil {
			return err
		} else if 0 <= handicap && handicap <= maxHandicapLevel {
			uci.Engine.Options.HandicapLevel = int(handicap)
		} else {
			return fmt.Errorf("Handicap Level must be between 0 and %d", maxHandicapLevel)
		}
		return nil
	case "Ponder":
		return nil
	default:
		return fmt.Errorf("unhandled option %s", option[1])
	}
}
