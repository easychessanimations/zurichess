package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"zurichess/engine"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("info string ")
	log.SetFlags(log.Lshortfile)
	flag.Parse()

}

func main() {
	bio := bufio.NewReader(os.Stdin)
	uci := &engine.UCI{}

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Println("error:", err)
			break
		}
		if err := uci.Execute(string(line)); err != nil {
			if err != engine.ErrQuit {
				log.Println("error:", err)
			}
			break
		}
	}
}
