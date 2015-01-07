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
	Engine *Engine
}

func NewUCI() *UCI {
	return &UCI{
		Engine: NewEngine(nil, EngineOptions{}),
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
	fmt.Println("id author Alexandru Mosoi")
	fmt.Println()
	fmt.Printf("option name UCI_AnalyseMode type check default %v\n", uci.Engine.Options.AnalyseMode)
	fmt.Printf("option name Hash type spin default %v min 1 max 8192\n", DefaultHashTableSizeMB)
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

	uci.Engine.SetPosition(pos)

	if i < len(args) {
		if args[i] != "moves" {
			return fmt.Errorf("expected 'moves', got '%s'", args[1])
		}
		for _, m := range args[i+1:] {
			move := uci.Engine.UCIToMove(m)
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
	if uci.Engine.Position.ToMove == White {
		tc = &white
	} else {
		tc = &black
	}
	tc.Start()

	GlobalHashTable.ResetStats()
	move, _ := uci.Engine.Play(tc)
	if uci.Engine.Options.AnalyseMode {
		hit, miss := GlobalHashTable.Hit, GlobalHashTable.Miss
		log.Printf("hash: size %d, hit %d, miss %d, ratio %.2f%%",
			GlobalHashTable.Size(), hit, miss,
			float32(hit)/float32(hit+miss)*100)
	}

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
		analyseMode, err := strconv.ParseBool(args[3])
		if err != nil {
			return err
		}
		uci.Engine.Options.AnalyseMode = analyseMode
	case "Hash":
		hashSizeMB, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return err
		}
		GlobalHashTable = NewHashTable(int(hashSizeMB))

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
	case "PawnBonusMg":
		FigureBonus[Pawn][MidGame] = value
	case "KnightBonusMg":
		FigureBonus[Knight][MidGame] = value
	case "BishopBonusMg":
		FigureBonus[Bishop][MidGame] = value
	case "RookBonusMg":
		FigureBonus[Rook][MidGame] = value
	case "QueenBonusMg":
		FigureBonus[Queen][MidGame] = value

	case "PawnBonusEg":
		FigureBonus[Pawn][EndGame] = value
	case "KnightBonusEg":
		FigureBonus[Knight][EndGame] = value
	case "BishopBonusEg":
		FigureBonus[Bishop][EndGame] = value
	case "RookBonusEg":
		FigureBonus[Rook][EndGame] = value
	case "QueenBonusEg":
		FigureBonus[Queen][EndGame] = value

	case "BishopPairBonus":
		BishopPairBonus = value
	case "KnightPawnBonus":
		KnightPawnBonus = value
	case "RookPawnPenalty":
		RookPawnPenalty = value
	default:
		return fmt.Errorf("unknown setvalue argument %s", args[0])
	}
	return nil
}
