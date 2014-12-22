package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"zurichess/engine"
)

var (
	buildVersion = "(devel)"
	buildTime    = "(just now)"
)

func init() {
	fmt.Printf("zurichess %v, build with %v at %v, running on %v\n",
		buildVersion, runtime.Version(), buildTime, runtime.GOARCH)

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
				log.Println("for line:", string(line))
				log.Println("error:", err)
			}
		}
	}
}
