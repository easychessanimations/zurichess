# zurichess: a chess engine

[Website](https://bitbucket.org/brtzsnr/zurichess) |
[CCRL](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details&each_game=1&eng=Zurichess%20150116#Zurichess_150116) |
[Documentation](https://godoc.org/bitbucket.org/brtzsnr/zurichess) |
[Wiki](http://chessprogramming.wikispaces.com/Zurichess)

zurichess is a chess engine and a chess library written in [Go](http://golang).

zurichess partially implements [UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html), but the available commands are enough for most purposes.

zurichess was successfully tested under Linux AMD64 and Linux ARM and other people have tested zurichess under Windows AMD64.
Precompiled binaries for a number of platforms and architectures can be found on the [downloads](https://bitbucket.org/brtzsnr/zurichess/downloads) page.


## Build and Compile

First you need to get the latest version of Go (currently 1.4). For instructions how to download and install Go for your OS see
[documentation](https://golang.org/doc/install). After Go 1.4 is installed, a workspace needs to be created:

```
#!bash
$ mkdir gows ; cd gows
$ export GOPATH=`pwd`
```

After the workspace is created cloning and compiling zurichess is easy:

```
#!bash
$ go get -u bitbucket.org/brtzsnr/zurichess/zurichess
$ $GOPATH/bin/zurichess --version
zurichess (devel), build with go1.4 at (just now), running on amd64
```

## Perft

A [perft](https://chessprogramming.wikispaces.com/Perft) tool is included.
The tool supports any starting position and can do splits up to several levels which is very helpful for debugging a move generator.
You can find more positions, results and external links on the [documentation](https://godoc.org/bitbucket.org/brtzsnr/zurichess/perft) page.

```
#!bash
$ go get -u bitbucket.org/brtzsnr/zurichess/perft
$ $GOPATH/bin/perft --fen "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
Searching FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
depth        nodes   captures enpassant castles   promotions eval  KNps   elapsed
-----+------------+----------+---------+---------+----------+-----+------+-------
    1           20          0         0         0          0 good    154 129.948µs
    2          400          0         0         0          0 good    158 2.531444ms
    3         8902         34         0         0          0 good    266 33.494604ms
    4       197281       1576         0         0          0 good   3454 57.114844ms
    5      4865609      82719       258         0          0 good  12141 400.762477ms
    6    119060324    2812008      5248         0          0 good  24027 4.955285846s
    7   3195901860  108329926    319617    883453          0 good  40040 1m19.817376124s
```

## Testing

[zuritest](https://bitbucket.org/brtzsnr/zuritest) is the framework used to test zurichess.

## History

### zurichess - appenzeller (in development)

This release's theme is improving search.

* Cleaned code and improved documentation.
* Implemented aspiration window search with gradual widening.
* Improved replacement strategy in transposition table.
* Doubled the number of entries in the transposition table.
* Developed [zuritest](https://bitbucket.org/brtzsnr/zuritest), testing infrastructure for zurichess.
* Fail-softly in more situations.
* Implemented UCI commands _go movetime_ and _stop_.
* Added a separate table for principal variation.
* Added killer heuristic to improve move ordering.
* Extended search when current position is in check.
* Improved time-control. In particular zurichess uses more time when there are fewer pieces on the board.

### zurichess - aargau (stable)

This is the first public release. ELO is about 1727 according to CCRL40/4.

* Core search function is a mini-max with alpha-beta pruning on top of a negamax framework.
* Sliding attacks are implemented using fancy magic bitboards.
* Search is sped up with transposition table with Zobrist hashing.
* Move ordering inside alpha-beta is done using table move & Most Valuable Victim / Least Valuable Victim.
* Quiescence search is used to reduce search instability and horizon effect.
* [Simplified evaluation function](https://chessprogramming.wikispaces.com/Simplified+evaluation+function) with tapered eval.

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
