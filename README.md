# zurichess: a chess engine

zurichess is a chess engine and a chess library written in [Go](http://golang).

zurichess supports [UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html).


## Building

First you need to get the latest version of Go (currently 1.4).
For instructions how to download and install Go for your OS see
[documentation](https://golang.org/doc/install).

After Go 1.4 is installed, a workspace needs to be created:

```
#!bash

$ mkdir gows ; cd gows
$ export GOPATH=`pwd`
```

## Compiling

After the workspace is created downloading and compiling zurichess is easy:

```
#!bash
$ go get -u bitbucket.org/brtzsnr/zurichess/zurichess
$ $GOPATH/bin/zurichess --version
zurichess (devel), build with go1.4 at (just now), running on amd64
```

zurichess was successfully tested under Linux amd64 and Linux arm. Other people have tested zurichess under Windows amd64. Precompiled binaries for a number of platforms and architectures can be found on the [downloads](https://bitbucket.org/brtzsnr/zurichess/downloads) page.

## Perft

A [perft](https://chessprogramming.wikispaces.com/Perft) tool is included. The tool supports any starting position and can do splits up to several levels which is very helpful for debugging a move generator. You can find other positions and results [here](http://www.10x8.net/chess/PerfT.html).

```
#!bash
$ go get -u bitbucket.org/brtzsnr/zurichess/zurichess
$ $GOPATH/bin/perft --fen "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
Searching FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
depth        nodes   captures enpassant castles eval   KNps elapsed
-----+------------+----------+---------+-------+----+------+-------
    1           20          0         0       0 good    153 130.414µs
    2          400          0         0       0 good    165 2.430465ms
    3         8902         34         0       0 good    281 31.688094ms
    4       197281       1576         0       0 good   4950 39.852835ms
    5      4865609      82719       258       0 good  13651 356.422314ms
    6    119060324    2812008      5248       0 good  26850 4.434284988s
    7   3195901860  108329926    319617  883453       45723 1m9.897481069s
```

## External links

Here is list of zurichess related links:

* [Chess Programming WIKI](http://chessprogramming.wikispaces.com/Zurichess)
* [CCRL 40/4](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details&eng=Zurichess%20150116)

Other sites, pages and articles with a lot of useful information:

* [Chess Programming WIKI](http://chessprogramming.wikispaces.com)
* [Computer Chess Club Forum](http://talkchess.com/forum/index.php)
* [The effect of hash collisions in a Computer Chess program](https://cis.uab.edu/hyatt/collisions.html)
* [Computer Chess Programming Theory](http://www.frayn.net/beowulf/theory.html)
* [Chess Programming Part V: Advanced Search](http://www.gamedev.net/page/resources/_/technical/artificial-intelligence/chess-programming-part-v-advanced-search-r1197)

## Disclaimer

This project is not associated with my employer.