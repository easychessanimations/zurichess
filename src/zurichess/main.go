package main

import (
	"bufio"
	"flag"
	"log"
	"math/rand"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("info string ")
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	rand.Seed(5)
	initBbKnightAttack()
	initRookMagic()
	initBishopMagic()
}

func main() {
	var pos *Position
	_ = pos

	bio := bufio.NewReader(os.Stdin)
	uci := &UCI{}

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Println("error:", err)
			break
		}
		if err := uci.Execute(string(line)); err != nil {
			if err != ErrQuit {
				log.Println("error:", err)
			}
			break
		}
	}
}
