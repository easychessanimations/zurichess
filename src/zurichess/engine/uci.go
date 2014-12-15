package engine

import (
	"fmt"
	"log"
	"math/rand"
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

func (uci *UCI) positionStartpos(args []string) error {
	var err error
	uci.Position, err = PositionFromFEN(FENStartPos)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return nil
	}
	if args[1] != "moves" {
		return fmt.Errorf("expected 'moves', got '%s'", args[1])
	}

	for _, m := range args[2:] {
		move := uci.Position.ParseMove(m)
		uci.Position.DoMove(move)
	}
	return nil
}

func (uci *UCI) positionFEN(args []string) error {
	if len(args) != 7 {
		return fmt.Errorf("expected 6 fields in FEN position, got %d",
			len(args)-1)
	}

	var err error
	uci.Position, err = PositionFromFEN(strings.Join(args[1:], " "))
	return err
}

func (uci *UCI) position(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected argument for 'position'")
	}

	switch args[0] {
	case "startpos":
		if err := uci.positionStartpos(args); err != nil {
			return nil
		}
	case "fen":
		if err := uci.positionFEN(args); err != nil {
			return nil
		}
	default:
		return fmt.Errorf("unknown position command: %s", args[0])
	}

	uci.Engine = NewEngine(uci.Position)
	return nil
}

func (uci *UCI) go_(args []string) {
	move, _ := uci.Engine.Play()
	log.Printf("selected %q (%v); piece %v", move, move, uci.Position.Get(move.From))
	fmt.Printf("bestmove %v\n", move)
}
