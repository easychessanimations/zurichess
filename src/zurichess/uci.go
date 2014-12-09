package main

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
	pos *Position
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
	uci.pos = nil
	return nil
}

func (uci *UCI) position(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected argument for 'position'")
	}

	if args[0] != "startpos" {
		return fmt.Errorf("expected 'startpos', got '%s'", args[0])
	}
	var err error
	uci.pos, err = PositionFromFEN(FENStartPos)
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
		move := uci.pos.ParseMove(m)
		uci.pos.DoMove(move)
	}

	uci.pos.PrettyPrint()
	return nil
}

func (uci *UCI) go_(args []string) {
	moves := uci.pos.GenerateMoves()

	var move Move
	for {
		move = moves[rand.Intn(len(moves))]
		uci.pos.DoMove(move)
		if !uci.pos.IsChecked(uci.pos.toMove.Other()) {
			break
		}
		uci.pos.UndoMove(move)
	}

	uci.pos.PrettyPrint()
	log.Printf("selected %+q; piece %v", move, uci.pos.Get(move.From))
	fmt.Printf("bestmove %s%s\n", move.From, move.To)
}
