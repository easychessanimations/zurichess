package engine

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	ErrQuit = fmt.Errorf("quit")
)

type UCI struct {
	UCI_AnalyseMode bool

	Engine *Engine
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
	uci.UCI_AnalyseMode = true

	fmt.Println("id name zurichess")
	fmt.Println("id author Alexandru Mosoi")
	fmt.Println()
	fmt.Println("option name UCI_AnalyseMode type check default true")
	fmt.Println("uciok")
	return nil
}

func (uci *UCI) isready(args []string) error {
	fmt.Println("readyok")
	return nil
}

func (uci *UCI) ucinewgame(args []string) error {
	uci.Engine = nil
	return nil
}
func (uci *UCI) position(args []string) error {
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
		pos, err = PositionFromFEN(strings.Join(args[1:7], " "))
		i += 7
	default:
		err = fmt.Errorf("unknown position command: %s", args[0])
	}
	if err != nil {
		return err
	}

	uci.Engine = NewEngine(pos)
	uci.Engine.AnalyseMode = uci.UCI_AnalyseMode

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
	if uci.Engine.position.ToMove == White {
		tc = &white
	} else {
		tc = &black
	}

	tc.Start()
	move, _ := uci.Engine.Play(tc)
	fmt.Printf("bestmove %v\n", move)
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
		mode, err := strconv.ParseBool(args[3])
		if err != nil {
			return err
		}
		uci.UCI_AnalyseMode = mode
	default:
		return fmt.Errorf("unhandled option %s", args[2])
	}

	return nil
}

func (uci *UCI) setvalue(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("setvalue expected two arguments, got %d", len(args))
	}
	value, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	switch args[0] {
	case "PawnBonus":
		FigureBonus[Pawn] = value
	case "KnightBonus":
		FigureBonus[Knight] = value
	case "BishopBonus":
		FigureBonus[Bishop] = value
	case "RookBonus":
		FigureBonus[Rook] = value
	case "QueenBonus":
		FigureBonus[Queen] = value
	case "BishopPairBonus":
		BishopPairBonus = value
	case "KnightPawnBonus":
		KnownWinScore = value
	case "RookPawnPenalty":
		RookPawnPenalty = value
	default:
		return fmt.Errorf("unknown setvalue argument %s", args[0])
	}
	return nil
}
