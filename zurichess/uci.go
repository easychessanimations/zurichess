package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"bitbucket.org/brtzsnr/zurichess/engine"
)

var (
	ErrQuit = fmt.Errorf("quit")

	globals = map[string]interface{}{
		"MidGameMaterial": &engine.MidGameMaterial,
		"EndGameMaterial": &engine.EndGameMaterial,
	}
)

type UCI struct {
	Engine *engine.Engine
	Stop   chan struct{}
	Ready  sync.WaitGroup
}

func NewUCI() *UCI {
	options := engine.Options{AnalyseMode: false}
	return &UCI{
		Engine: engine.NewEngine(nil, options),
		Stop:   make(chan struct{}, 1),
	}
}

func (uci *UCI) Execute(line string) error {
	cmd := strings.Fields(line)
	if len(cmd) == 0 {
		return nil
	}

	fun, args := cmd[0], cmd[1:]

	var err error
	switch fun {
	case "uci":
		uci.uci(args)
	case "isready":
		uci.isready(args)
	case "ucinewgame":
		uci.ucinewgame(args)
	case "position":
		err = uci.position(args)
	case "go":
		uci.go_(args)
	case "stop":
		uci.stop(args)
	case "setoption":
		err = uci.setoption(args)
	case "setvalue":
		err = uci.setvalue(args)
	case "quit":
		err = ErrQuit
	default:
		log.Println("unhandled input: ", string(line))
	}

	return err
}

func (uci *UCI) uci(args []string) error {
	fmt.Println("id name zurichess")
	fmt.Println("id author Alexandru Moșoi")
	fmt.Println()
	fmt.Printf("option name UCI_AnalyseMode type check default false\n")
	fmt.Printf("option name Hash type spin default %v min 1 max 8192\n", engine.DefaultHashTableSizeMB)
	fmt.Printf("option name MvvLva type string\n")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(args []string) error {
	uci.Ready.Wait()
	fmt.Println("readyok")
	return nil
}

func (uci *UCI) ucinewgame(args []string) error {
	return nil
}

func (uci *UCI) position(args []string) error {
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

func (uci *UCI) go_(args []string) {
	var tc engine.TimeControl
	fdtc := &engine.FixedDepthTimeControl{}
	octc := &engine.OnClockTimeControl{NumPieces: int(uci.Engine.Position.NumPieces[engine.NoColor][engine.NoFigure])}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "infinite":
			i++
			octc.Time = 1000000 * time.Hour
		case "wtime":
			i++
			t, _ := strconv.Atoi(args[i])
			if uci.Engine.Position.SideToMove == engine.White {
				octc.Time = time.Duration(t) * time.Millisecond
			}
		case "winc":
			i++
			t, _ := strconv.Atoi(args[i])
			if uci.Engine.Position.SideToMove == engine.White {
				octc.Inc = time.Duration(t) * time.Millisecond
			}
		case "btime":
			i++
			t, _ := strconv.Atoi(args[i])
			if uci.Engine.Position.SideToMove == engine.Black {
				octc.Time = time.Duration(t) * time.Millisecond
			}
		case "binc":
			i++
			t, _ := strconv.Atoi(args[i])
			if uci.Engine.Position.SideToMove == engine.Black {
				octc.Inc = time.Duration(t) * time.Millisecond
			}
		case "movestogo":
			i++
			t, _ := strconv.Atoi(args[i])
			octc.MovesToGo = t
			tc = octc
		case "movetime":
			i++
			t, _ := strconv.Atoi(args[i])
			octc.Time, octc.Inc, octc.MovesToGo = time.Duration(t)*time.Millisecond, 0, 0
			tc = octc
		case "depth":
			i++
			d, _ := strconv.Atoi(args[i])
			fdtc.MinDepth = 0
			fdtc.MaxDepth = d
			tc = fdtc
		}
	}

	uci.Ready.Add(1)
	go func() {
		defer uci.Ready.Done()

		select {
		case <-uci.Stop: // Clear the channel if there is a stop command pending.
		default:
		}

		octc.Stop = uci.Stop
		tc.Start()

		moves := uci.Engine.Play(tc)
		hit, miss := uci.Engine.Stats.CacheHit, uci.Engine.Stats.CacheMiss
		log.Printf("hash: size %d, hit %d, miss %d, ratio %.2f%%",
			engine.GlobalHashTable.Size(), hit, miss,
			float32(hit)/float32(hit+miss)*100)
		if len(moves) == 0 {
			fmt.Printf("bestmove %v\n", "(none)")
		} else {
			fmt.Printf("bestmove %v\n", moves[0].UCI())
		}
	}()
}

func (uci *UCI) stop(args []string) error {
	select {
	case uci.Stop <- struct{}{}:
	default: // There is another stop event on the line.
	}
	return nil
}

func (uci *UCI) setoption(args []string) error {
	if args[0] != "name" {
		return fmt.Errorf("expected first field 'name', got %s", args[0])
	}
	if args[2] != "value" {
		return fmt.Errorf("expected third field 'value', got %s", args[0])
	}

	switch args[1] {
	case "UCI_AnalyseMode":
		if mode, err := strconv.ParseBool(args[3]); err != nil {
			return err
		} else {
			uci.Engine.Options.AnalyseMode = mode
		}
	case "Hash":
		if hashSizeMB, err := strconv.ParseInt(args[3], 10, 64); err != nil {
			return err
		} else {
			engine.GlobalHashTable = engine.NewHashTable(int(hashSizeMB))
		}
	case "MvvLva":
		return engine.SetMvvLva(args[3])
	default:
		return fmt.Errorf("unhandled option %s", args[2])
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

// setvalue will fatal in case of error because otherwise
// an error will make tunning useless.
func (uci *UCI) setvalue(args []string) error {
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
