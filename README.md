# zurichess: a chess engine

zurichess is a chess engine and a chess library written in [Go](http://golang).

zurichess supports [UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html).
zurichess plays automatically on [FICS](http://www.fics.org) under unregistered
user zurichess.

## Usage

```
#!bash

$ git clone -b release.aargau https://bitbucket.org/brtzsnr/zurichess.git
$ cd zurichess/src/zurichess
$ export GOPATH=`pwd`
$ go version
go version go1.4 linux/amd64
$ ./release.sh
Built zc-master-c6b5253 at 2015-01-16 15:31:29
```

## Perft

A [perft](https://chessprogramming.wikispaces.com/Perft) tool is included.

```
#!bash
$ go run perft/perft.go 
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

## Disclaimer

This project is not associated with my employer.
