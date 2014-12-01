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
	log.SetPrefix("info ")
	log.SetFlags(log.Ltime | log.Lshortfile)
	flag.Parse()
}

func main() {
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
			// TODO
			continue
		case "position":
			// TODO
			continue
		case "go":
			// TODO
			continue

		default:
			log.Println("unhandled input: ", string(line))
		}
	}
}
