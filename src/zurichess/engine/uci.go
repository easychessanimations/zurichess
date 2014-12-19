package engine

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	ErrQuit = fmt.Errorf("quit")
)

type UCI struct {
	Position *Position
	Engine   *Engine
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
	case "quit":
		err = ErrQuit
	default:
		log.Println("unhandled input: ", string(line))
	}

	return err
}

func (uci *UCI) uci(args []string) error {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("id name zurichess")
	fmt.Println("id author Alexandru Moșoi")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(args []string) error {
	fmt.Println("readyok")
	return nil
}

func (uci *UCI) ucinewgame(args []string) error {
	uci.Position = nil
	uci.Engine = nil
	return nil
}
func (uci *UCI) position(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected argument for 'position'")
	}

	i := 0
	var err error
	switch args[i] {
	case "startpos":
		uci.Position, err = PositionFromFEN(FENStartPos)
		i++
	case "fen":
		uci.Position, err = PositionFromFEN(strings.Join(args[1:7], " "))
		i += 7
	default:
		err = fmt.Errorf("unknown position command: %s", args[0])
	}
	if err != nil {
		return err
	}

	uci.Engine = NewEngine(uci.Position)
	if i < len(args) {
		if args[i] != "moves" {
			return fmt.Errorf("expected 'moves', got '%s'", args[1])
		}
		for _, m := range args[i+1:] {
			move := uci.Engine.ParseMove(m)
			uci.Engine.DoMove(move)
		}
	}

	return nil
}

func (uci *UCI) go_(args []string) {
	var white, black OnClockTimeControl
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
		}
	}

	var tc TimeControl
	if uci.Position.ToMove == White {
		tc = &white
	} else {
		tc = &black
	}

	tc.Start()
	move, _ := uci.Engine.Play(tc)
	log.Printf("selected %q (%v); piece %v", move, move, uci.Position.Get(move.From))
	fmt.Printf("bestmove %v\n", move)
}
