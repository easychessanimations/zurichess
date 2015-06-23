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
	buildVersion = "bern"
	buildTime    = "(just now)"

	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	version    = flag.Bool("version", false, "only print version and exit")
)

func init() {
	if buildTime == "(just now)" {
		// If build time is not known assume it's the modification time of the binary
		fi, err := os.Stat(os.Args[0])
		if err != nil {
			return
		}
		buildTime = fi.ModTime().Format("2006-01-02 15:04:05")
	}
}

func main() {
	fmt.Printf("zurichess %v https://bitbucket.org/brtzsnr/zurichess\n", buildVersion)
	fmt.Printf("build with %v at %v, running on %v\n", runtime.Version(), buildTime, runtime.GOARCH)

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
			if err != errQuit {
				log.Println(err)
			} else {
				break
			}
		}
	}
}
