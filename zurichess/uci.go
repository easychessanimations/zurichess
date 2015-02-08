package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/brtzsnr/zurichess/engine"
)

var (
	ErrQuit = fmt.Errorf("quit")
)

type UCI struct {
	Engine *engine.Engine
}

func NewUCI() *UCI {
	return &UCI{
		Engine: engine.NewEngine(nil, engine.EngineOptions{}),
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
	case "setoption":
		err = uci.setoption(args)
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
	fmt.Printf("option name Hash type spin default %v min 1 max 8192\n", engine.DefaultHashTableSizeMB)
	fmt.Printf("option name MvvLva type string\n")
	fmt.Printf("option name FigureBonus.MidGame type string\n")
	fmt.Printf("option name FigureBonus.EndGame type string\n")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(args []string) error {
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
	var white, black engine.OnClockTimeControl
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "infinite":
			white.Time = 1000000 * time.Hour
			black.Time = 1000000 * time.Hour
		case "wtime":
			i++
			t, _ := strconv.Atoi(args[i])
			white.Time = time.Duration(t) * time.Millisecond
		case "winc":
			i++
			t, _ := strconv.Atoi(args[i])
			white.Inc = time.Duration(t) * time.Millisecond
		case "btime":
			i++
			t, _ := strconv.Atoi(args[i])
			black.Time = time.Duration(t) * time.Millisecond
		case "binc":
			i++
			t, _ := strconv.Atoi(args[i])
			black.Inc = time.Duration(t) * time.Millisecond
		case "movestogo":
			i++
			t, _ := strconv.Atoi(args[i])
			white.MovesToGo = t
			black.MovesToGo = t
		case "movetime":
			i++
			t, _ := strconv.Atoi(args[i])
			white.Time, white.Inc, white.MovesToGo = time.Duration(t)*time.Millisecond, 0, 0
			black.Time, black.Inc, black.MovesToGo = time.Duration(t)*time.Millisecond, 0, 0
		}
	}

	var tc engine.TimeControl
	if uci.Engine.Position.ToMove == engine.White {
		tc = &white
	} else {
		tc = &black
	}
	tc.Start()

	moves := uci.Engine.Play(tc)
	hit, miss := uci.Engine.Stats.CacheHit, uci.Engine.Stats.CacheMiss
	log.Printf("hash: size %d, hit %d, miss %d, ratio %.2f%%",
		engine.GlobalHashTable.Size(), hit, miss,
		float32(hit)/float32(hit+miss)*100)
	fmt.Printf("bestmove %v\n", moves[0])
}

func (uci *UCI) setoption(args []string) error {
	if args[0] != "name" {
		return fmt.Errorf("expected first field 'name', got %s", args[0])
	}
	if args[2] != "value" {
		return fmt.Errorf("expected third field 'value', got %s", args[0])
	}

	switch args[1] {
	case "Hash":
		if hashSizeMB, err := strconv.ParseInt(args[3], 10, 64); err != nil {
			return err
		} else {
			engine.GlobalHashTable = engine.NewHashTable(int(hashSizeMB))
		}
	case "MvvLva":
		return engine.SetMvvLva(args[3])
	case "FigureBonus.MidGame":
		return engine.SetMaterialValue(args[1], engine.MidGameMaterial.FigureBonus[:], args[3])
	case "FigureBonus.EndGame":
		return engine.SetMaterialValue(args[1], engine.EndGameMaterial.FigureBonus[:], args[3])
	default:
		return fmt.Errorf("unhandled option %s", args[2])
	}

	return nil
}
