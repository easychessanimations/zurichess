package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	buildVersion = "(devel)"
	buildTime    = "(just now)"

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	version    = flag.Bool("version", false, "only print version and exit")
)

func main() {
	fmt.Printf("zurichess %v, build with %v at %v, running on %v\n",
		buildVersion, runtime.Version(), buildTime, runtime.GOARCH)

	flag.Parse()
	if *version {
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.SetOutput(os.Stdout)
	log.SetPrefix("info string ")
	log.SetFlags(log.Lshortfile)

	bio := bufio.NewReader(os.Stdin)
	uci := NewUCI()
	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			log.Println("error:", err)
			break
		}
		if err := uci.Execute(string(line)); err != nil {
			if err != ErrQuit {
				log.Println("for line:", string(line))
				log.Println("error:", err)
			}
		}
	}
}
