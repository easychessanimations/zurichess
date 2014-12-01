package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("info debug ")
	log.SetFlags(log.Lshortfile)
	flag.Parse()
}

func main() {
	var pos *Position
	_ = pos

	bio := bufio.NewReader(os.Stdin)
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Println("error:", err)
			break
		}

		cmd := strings.Fields(string(line))
		if len(cmd) == 0 {
			continue
		}

		switch cmd[0] {
		case "uci":
			fmt.Println("id name zurichess")
			fmt.Println("id author Alexandru Moșoi")
			fmt.Println("uciok")
		case "isready":
			fmt.Println("readyok")
		case "ucinewgame":
			continue
		case "position":
			if cmd[1] != "startpos" {
				log.Fatalln("expected 'startpos', got'", cmd[1])
			}
			pos, err = PositionFromFEN(FENStartPos)
			if err != nil {
				log.Fatalln(err)
			}
			pos.PrettyPrint()
			continue
		case "go":
			// TODO
			continue

		default:
			log.Println("unhandled input: ", string(line))
		}
	}
}
