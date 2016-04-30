# History

Versions are named after [Swiss Cantons](http://en.wikipedia.org/wiki/Cantons_of_Switzerland)
in alphabetical order.

## zurichess [graubuenden](https://en.wikipedia.org/wiki/Graub%C3%BCnden) (development)

## zurichess [glarus](https://en.wikipedia.org/wiki/Canton_of_Glarus) (stable)
17.Apr.2016

The theme of this release is king safety and leaf pruning.

* Improve futility conditions. Geneva's futility is a bit too aggressive and causes lots of tactical mistakes.
* Add History Leaf Pruning similar to https://chessprogramming.wikispaces.com/History+Leaf+Pruning.
* Improve pawn evaluation caching. Also cache shelter evaluation.
* Improve king safety using number of simultaneous attackers.
* Improve time control. Timeouts should be extremely rare now.
* Small tunining of LMR and NMP conditions.
* Micro-optimize the code for the future Go compiler. Next version will see big speed up.
* Move Position.SANToMove to https://bitbucket.org/zurichess/notation
* Move Polyglot hashing to https://bitbucket.org/zurichess/hashing
* Usual code clean ups, speed ups and bug fixes.

## zurichess [geneva](https://en.wikipedia.org/wiki/Canton_of_Geneva)
04.Dec.2015

The theme of this release is improving evaluation.
ELO is about 2475 on CCRL 40/40 and 2320 on CEGT 40/4.

* Implement fifty-move draw rule. Add HasLegalMoves and InsufficientMaterial methods.
* Improve move ordering: add killer phase; remove sorting.
* Improve time control: add more time when the move is predicted.
* Add basic futility pruning.
* Switch tuning to using [TensorFlow](http://tensorflow.org/) framework. txt is now deprecated.
* Evaluate rooks on open and half-open files.
* Improve mobility calculation.
* Tweak null-move conditions: allow double null-moves.
* Usual code clean ups, speed ups and bug fixes.

## zurichess [fribourg](https://en.wikipedia.org/wiki/Canton_of_Fribourg)
04.Sep.2015

The theme of this release is tuning the evaluation, search and move generation.
ELO is about 2442 on CCRL 40/40.

* Move to the new page http://bitbucket.org/zurichess/zurichess.
* Evaluate passed, connected and isolated pawns. Tuning was done
using Texel's tuning method implemented by
[txt](https://bitbucket.org/zurichess/txt).
* Add Static Exchange Evaluation (SEE).
* Ignore bad captures (SEE < 0) in quiescence search.
* Late move reduce (LMR) of all quiet non-critical moves. Aggressively reduce
bad quiet (SEE < 0) moves at higher depths.
* Adjust LMR conditions. Reduce more at high depths (near root) and high move count.
* Increase number of killers to 4. Helps with more aggressive LMR.
* Improve move generation speed. Add phased move generation: hash,
captures, and quiets. Phased move generation allows the engine to skip
generation or sorting of the moves in many cases.
* Implement `setoption Clear Hash`.
* Implement pondering. Should give some ELO boost for online competitions.
* Improve move generation order. Picked the best among 20 random orders.
* Prune two-fold repetitions at non-root nodes. This pruning cuts huge parts
of the search tree without affecting search quality. >30ELO improvement
in self play.
* Small time control adjustment. Still too little time used in the mid
game. Abort search if it takes much more time than alloted.
* Usual code clean ups, speed ups and bug fixes.

## zurichess [bern](http://en.wikipedia.org/wiki/Canton_of_Bern)
25.Jun.2015

This release's theme is pruning the search. ELO is about 2234 on CCRL 40/4.

* Implement Principal Variation Search (PVS).
* Reduce late quiet moves (LMR).
* Optimize move ordering. Penalize moves threatened by pawns in quiescence search.
* Optimize check extension. Do not extend many bad checks.
* Change Zobrist key to be equal to polyglot key. No book support, but better hashing.
* Add some integration tests such as mate in one and mate in two.
* Usual code clean ups, speed ups and bug fixes.

## zurichess [basel](http://en.wikipedia.org/wiki/Basel-Stadt)
28.Apr.2015

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

## zurichess [appenzeller](http://en.wikipedia.org/wiki/Appenzeller_cheese)
23.Feb.2015

This release's theme is improving search. ELO is about 1823 on CCRL 40/4.

* Clean code and improved documentation.
* Implement aspiration window search with gradual widening.
* Improve replacement strategy in transposition table.
* Double the number of entries in the transposition table.
* Develop [zuritest](https://bitbucket.org/zurichess/zuritest), testing infrastructure for zurichess.
* Fail-softly in more situations.
* Implement UCI commands `go movetime` and `stop`.
* Add a separate table for principal variation.
* Add killer heuristic to improve move ordering.
* Extend search when current position is in check.
* Improve time-control. In particular zurichess uses more time when there are fewer pieces on the board.

## zurichess [aargau](http://en.wikipedia.org/wiki/Aargau)
22.Jan.2015

This is the first public release. ELO is about 1727 on CCRL 40/4.

* Core search function is a mini-max with alpha-beta pruning on top of a negamax framework.
* Sliding attacks are implemented using fancy magic bitboards.
* Search is sped up with transposition table with Zobrist hashing.
* Move ordering inside alpha-beta is done using table move & Most Valuable Victim / Least Valuable Victim.
* Quiescence search is used to reduce search instability and horizon effect.
* [Simplified evaluation function](https://chessprogramming.wikispaces.com/Simplified+evaluation+function) with tapered eval.

