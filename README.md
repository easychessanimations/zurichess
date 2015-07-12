# zurichess: a chess engine

[Website](https://bitbucket.org/zurichess/zurichess) |
[CCRL](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details+%28text%29&eng=Zurichess%20Appenzeller%2064-bit) |
[Wiki](http://chessprogramming.wikispaces.com/Zurichess) |
[![Reference](https://godoc.org/bitbucket.org/zurichess/zurichess?status.svg)](https://godoc.org/bitbucket.org/zurichess/zurichess)
[![Build Status](https://drone.io/bitbucket.org/zurichess/zurichess/status.png)](https://drone.io/bitbucket.org/zurichess/zurichess/latest)

zurichess is a chess engine and a chess library written in
[Go](http://www.golang.org). Its main goals are: to be a relatively
strong chess engine and to enable chess tools writing. See
the library reference.

zurichess partially implements [UCI
protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html), but
the available commands are enough for most purposes.  zurichess was
successfully tested under Linux AMD64 and Linux ARM and other people
have tested zurichess under Windows AMD64.

zurichess plays on [FICS](http://freechess.org) under handle
[zurichess](http://ficsgames.org/cgi-bin/search.cgi?player=zurichess&action=Statistics).
Usually it runs code at tip (master) which is a bit stronger
than the latest stable version.

## Build and Compile

First you need to get the latest version of Go (currently 1.4). For
instructions how to download and install Go for your OS see
[documentation](https://golang.org/doc/install). After Go 1.4 is
installed, a workspace needs to be created:

```
#!bash
$ mkdir gows ; cd gows
$ export GOPATH=`pwd`
```

After the workspace is created cloning and compiling zurichess is easy:

```
#!bash
$ go get -u bitbucket.org/zurichess/zurichess/zurichess
$ $GOPATH/bin/zurichess --version
zurichess (devel), build with go1.4 at (just now), running on amd64
```

## Download

Precompiled binaries for several platforms and architectures can be found
on the [downloads](https://bitbucket.org/zurichess/zurichess/downloads)
page.

Latest Linux AMD64 binaries can be downloaded from
[drone.io](https://drone.io/bitbucket.org/zurichess/zurichess/files). These
binaries should be stable for any kind of testing.


## Testing

[zuritest](https://bitbucket.org/zurichess/zuritest) is the framework used to test zurichess.

## History

Versions are named after [Swiss Cantons](http://en.wikipedia.org/wiki/Cantons_of_Switzerland)
in alphabetical order.

### zurichess - [fribourg](https://en.wikipedia.org/wiki/Canton_of_Fribourg) (development)

* Moved to new page http://bitbucket/zurichess/zurichess.
* Evaluate passed, connected and isolated pawns. Tuning was done using Texel's tuning method
implemented using [txt](https://bitbucket.org/brtzsnr/txt).
* Add Static Exchange Evalution (SEE).
* Ignore bad captures (SEE < 0) in quiescence search.
* Late move reduce of all quiet moves. Aggresively reduce bad quiet (SEE < 0)
moves at higher depths.
* Increase number of killers to 4.
* Improve move generation speed. Add phased move generation: hash, captures, quiet
allows the engine to skip generation or sorting of the moves in many cases.
* Implement `setoption Clear Hash`.
* Implement pondering.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [bern](http://en.wikipedia.org/wiki/Canton_of_Bern) (stable)

This release's theme is pruning the search. ELO is about 2234 on CCRL 40/4.

* Implement Principal Variation Search (PVS).
* Reduce late quiet moves (LMR).
* Optimize move ordering. Penalize moves threatened by pawns in quiescence search.
* Optimize check extension. Do not extend many bad checks.
* Change zobrist key to be equal to polyglot key. No book support, but better hashing.
* Add some integration tests such as mate in one and mate in two.
* Usual code clean ups, speed ups and bug fixes.

### zurichess - [basel](http://en.wikipedia.org/wiki/Basel-Stadt)

This release's theme is improving evaluation function.

* Speed up move ordering considerably.
* Implement null move pruning.
* Clean up and micro optimize the code.
* Tune check extensions and move ordering.
* Award mobility and add new piece square tables.
* Handle three fold repetition.
* Cache pawn structure evaluation.
* Fix transposition table bug causing a search explosion around mates.
* Prune based on mate score.

### zurichess - [appenzeller](http://en.wikipedia.org/wiki/Appenzeller_cheese)

This release's theme is improving search. ELO is about 1823 on CCRL 40/4.

* Clean code and improved documentation.
* Implement aspiration window search with gradual widening.
* Improve replacement strategy in transposition table.
* Double the number of entries in the transposition table.
* Develop [zuritest](https://bitbucket.org/brtzsnr/zuritest), testing infrastructure for zurichess.
* Fail-softly in more situations.
* Implement UCI commands `go movetime` and `stop`.
* Add a separate table for principal variation.
* Add killer heuristic to improve move ordering.
* Extend search when current position is in check.
* Improve time-control. In particular zurichess uses more time when there are fewer pieces on the board.

### zurichess - [aargau](http://en.wikipedia.org/wiki/Aargau)

This is the first public release. ELO is about 1727 on CCRL 40/4.

* Core search function is a mini-max with alpha-beta pruning on top of a negamax framework.
* Sliding attacks are implemented using fancy magic bitboards.
* Search is sped up with transposition table with Zobrist hashing.
* Move ordering inside alpha-beta is done using table move & Most Valuable Victim / Least Valuable Victim.
* Quiescence search is used to reduce search instability and horizon effect.
* [Simplified evaluation function](https://chessprogramming.wikispaces.com/Simplified+evaluation+function) with tapered eval.

## External links

A list of zurichess related links:

* [Chess Programming WIKI](http://chessprogramming.wikispaces.com/Zurichess)
* [CCRL 40/4](http://www.computerchess.org.uk/ccrl/404/cgi/engine_details.cgi?print=Details+%28text%29&eng=Zurichess%20Appenzeller%2064-bit)
* [FICS Games](http://ficsgames.org/cgi-bin/search.cgi?player=zurichess&action=Statistics)

Other sites, pages and articles with a lot of useful information:

* [Chess Programming Part V: Advanced Search](http://www.gamedev.net/page/resources/_/technical/artificial-intelligence/chess-programming-part-v-advanced-search-r1197)
* [Chess Programming WIKI](http://chessprogramming.wikispaces.com)
* [Computer Chess Club Forum](http://talkchess.com/forum/index.php)
* [Computer Chess Programming](http://verhelst.home.xs4all.nl/chess/search.html)
* [Computer Chess Programming Theory](http://www.frayn.net/beowulf/theory.html)
* [How Stockfish Works](http://rin.io/chess-engine/)
* [Little Chess Evaluation Compendium](https://chessprogramming.wikispaces.com/file/view/LittleChessEvaluationCompendium.pdf)
* [The effect of hash collisions in a Computer Chess program](https://cis.uab.edu/hyatt/collisions.html)
* [The UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html)

## Disclaimer

This project is not associated with my employer.
