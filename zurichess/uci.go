// uci implements the UCI protocol which is described here http://wbec-ridderkerk.nl/html/UCIProtocol.html.
// There is a hidden command, setvalue, which can be used to set the material values.

package main

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/brtzsnr/zurichess/engine"
)

var (
	errQuit = fmt.Errorf("quit")

	globals = map[string]interface{}{
		"Material": &engine.GlobalMaterial,
	}
)

// UCI implements uci protocol.
type UCI struct {
	Engine      *engine.Engine
	timeControl *engine.TimeControl

	// buffer of 1, if empty then the engine is available
	ready chan struct{}
	// buffer of 1, if filled then the engine is pondering
	ponder chan struct{}
}

func NewUCI() *UCI {
	options := engine.Options{AnalyseMode: true}
	return &UCI{
		Engine:      engine.NewEngine(nil, options),
		timeControl: nil,
		ready:       make(chan struct{}, 1),
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

	// These commands do not expect the engine to be ready.
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

	// Make sure that the engine is ready.
	select {
	case uci.ready <- struct{}{}:
		<-uci.ready
	default:
		return fmt.Errorf("not ready for %s", line)
	}

	// These commands expect engine to be ready.
	switch cmd {
	case "ucinewgame":
		return uci.ucinewgame(line)
	case "position":
		return uci.position(line)
	case "go":
		return uci.go_(line)
	case "setoption":
		return uci.setoption(line)
	case "setvalue":
		return uci.setvalue(line)
	default:
		return fmt.Errorf("unhandled command %s", cmd)
	}
}

func (uci *UCI) uci(line string) error {
	fmt.Printf("id name zurichess %v\n", buildVersion)
	fmt.Printf("id author Alexandru Moșoi\n")
	fmt.Printf("\n")
	fmt.Printf("option name UCI_AnalyseMode type check default false\n")
	fmt.Printf("option name Hash type spin default %v min 1 max 8192\n", engine.DefaultHashTableSizeMB)
	fmt.Printf("option name Ponder type check default true\n")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(line string) error {
	uci.ready <- struct{}{}
	<-uci.ready
	fmt.Println("readyok")
	return nil
}

func (uci *UCI) ucinewgame(line string) error {
	// Clear the hash at the beginning of each game.
	engine.GlobalHashTable.Clear()
	return nil
}

func (uci *UCI) position(line string) error {
	args := strings.Fields(line)[1:]
	if len(args) == 0 {
		return fmt.Errorf("expected argument for 'position'")
	}

	var pos *engine.Position

	i := 0
	var err error
	switch args[i] {
	case "startpos":
		pos, err = engine.PositionFromFEN(engine.FENStartPos)
		i++
	case "fen":
		pos, err = engine.PositionFromFEN(strings.Join(args[1:7], " "))
		i += 7
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
			move := uci.Engine.Position.UCIToMove(m)
			uci.Engine.DoMove(move)
		}
	}

	return nil
}

func (uci *UCI) go_(line string) error {
	// TODO: Handle panic for `go depth`
	args := strings.Fields(line)[1:]
	uci.timeControl = engine.NewTimeControl(uci.Engine.Position)
	uci.timeControl.MovesToGo = 30 // in case there is not time refresh
	ponder := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "ponder":
			ponder = true
		case "infinite":
			uci.timeControl = engine.NewTimeControl(uci.Engine.Position)
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
			uci.timeControl.MovesToGo = t
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
			uci.timeControl.Depth = d
		}
	}

	if ponder {
		// Ponder was requested, so fill the channel.
		// Next write to uci.ponder will block.
		uci.ponder <- struct{}{}
	}

	uci.timeControl.Start(ponder)
	uci.ready <- struct{}{}
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
	// Waits until the engine becomes ready.
	uci.ready <- struct{}{}
	<-uci.ready

	return nil
}

// play starts the engine.
// Should run in its own separate goroutine.
func (uci *UCI) play() {
	moves := uci.Engine.Play(uci.timeControl)
	hit, miss := uci.Engine.Stats.CacheHit, uci.Engine.Stats.CacheMiss
	log.Printf("hash: size %d, hit %d, miss %d, ratio %.2f%%",
		engine.GlobalHashTable.Size(), hit, miss,
		float32(hit)/float32(hit+miss)*100)

	// If pondering was requested it will block because the channel is full.
	uci.ponder <- struct{}{}
	<-uci.ponder

	// Marks engine ready before showing the best move,
	// otherwise there is a race if GUI sees bestmove but engine
	// is not ready.
	<-uci.ready

	if len(moves) == 0 {
		fmt.Printf("bestmove (none)\n")
	} else if len(moves) == 1 {
		fmt.Printf("bestmove %v\n", moves[0].UCI())
	} else {
		fmt.Printf("bestmove %v ponder %v\n", moves[0].UCI(), moves[1].UCI())
	}
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
		engine.GlobalHashTable.Clear()
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
			engine.GlobalHashTable = engine.NewHashTable(int(hashSizeMB))
		}
		return nil
	default:
		return fmt.Errorf("unhandled option %s", option[1])
	}
}

// setvalue will fatal in case of error because otherwise
// an error will make tunning useless.
func (uci *UCI) setvalue(line string) error {
	args := strings.Fields(line)[1:]
	if len(args) != 2 {
		log.Fatalf("expected 2 arguments, got %d", len(args))
	}
	n, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = setvalueHelper(reflect.ValueOf(globals), args[0], n)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func setvalueHelper(v reflect.Value, fields string, n int) error {
	switch v.Kind() {
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int:
		if fields != "" {
			return fmt.Errorf("expected struct or slice")
		}
		if !v.CanSet() {
			return fmt.Errorf("cannot set value")
		}
		v.SetInt(int64(n))
	case reflect.Ptr:
		return setvalueHelper(v.Elem(), fields, n)
	case reflect.Struct:
		split := strings.SplitN(fields, ".", 2)
		if split[0] == "" {
			return fmt.Errorf("missing field for type %s", v.Type().Name())
		}
		field := v.FieldByName(split[0])
		if !field.IsValid() {
			return fmt.Errorf("no such field %s for type %s", split[0], v.Type().Name())
		}
		if len(split) == 1 {
			return setvalueHelper(field, "", n)
		}
		return setvalueHelper(field, split[1], n)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		split := strings.SplitN(fields, ".", 2)
		if split[0] == "" {
			return fmt.Errorf("missing index for type %s", v.Type().Name())
		}
		index, err := strconv.Atoi(split[0])
		if err != nil {
			return err
		}
		if index >= v.Len() {
			return fmt.Errorf("out of bounds")
		}
		if len(split) == 1 {
			return setvalueHelper(v.Index(index), "", n)
		}
		return setvalueHelper(v.Index(index), split[1], n)
	case reflect.Map:
		split := strings.SplitN(fields, ".", 2)
		if split[0] == "" {
			return fmt.Errorf("missing index for type %s", v.Type().Name())
		}
		if len(split) == 1 {
			return fmt.Errorf("expected more fields for type %s", v.Type().Name())
		}
		index := v.MapIndex(reflect.ValueOf(split[0]))
		if !index.IsValid() {
			return fmt.Errorf("no such map key %v", split[0])
		}
		return setvalueHelper(index, split[1], n)
	case reflect.Interface:
		return setvalueHelper(reflect.ValueOf(v.Interface()), fields, n)
	default:
		fmt.Println("unhandled v.Kind() ==", v.Kind())
	}

	return nil
}
