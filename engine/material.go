// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  = 25000          // KnownWinScore is strictly greater than all evaluation scores (mate not included).
	KnownLossScore = -KnownWinScore // KnownLossScore is strictly smaller than all evaluation scores (mated not included).
	MateScore      = 30000          // MateScore - N is mate in N plies.
	MatedScore     = -MateScore     // MatedScore + N is mated in N plies.
	InfinityScore  = 32000          // InfinityScore is possible score. -InfinityScore is the minimum possible score.
)

var (
	// Weights stores all evaluation parameters under one array for easy handling.
	// All numbers are multiplied by 128.
	//
	// Zurichess' evaluation is a very simple neural network with no hidden layers,
	// and one output node y = W_m * x * (1-p) + W_e * x * p where W_m are
	// middle game weights, W_e are endgame weights, x is input, p is phase between
	// middle game and end game, and y is the score.
	// The network has |x| = len(Weights) inputs corresponding to features
	// extracted from the position. These features are symmetrical wrt to colors.
	// The network is trained using the Texel's Tuning Method
	// https://chessprogramming.wikispaces.com/Texel%27s+Tuning+Method.
	Weights = [98]Score{
		{M: -135, E: 17}, {M: 10544, E: 11687}, {M: 47875, E: 34564}, {M: 51128, E: 37621}, {M: 70354, E: 70354}, {M: 171465, E: 117599}, {M: 48, E: -327}, {M: 20, E: 17},
		{M: 888, E: 3220}, {M: 1241, E: 1456}, {M: 802, E: 1073}, {M: 741, E: 664}, {M: 190, E: 913}, {M: -375, E: -867}, {M: -558, E: 1357}, {M: -1759, E: 1096},
		{M: -2155, E: 338}, {M: 95, E: -3144}, {M: -1265, E: -616}, {M: 3341, E: -592}, {M: 1872, E: -1220}, {M: -1682, E: -1957}, {M: -391, E: 1293}, {M: -1561, E: 0},
		{M: -1082, E: -656}, {M: -907, E: -1525}, {M: 531, E: 189}, {M: 146, E: 213}, {M: 1551, E: -1304}, {M: -1058, E: -581}, {M: -397, E: 2308}, {M: -1567, E: 1148},
		{M: 423, E: -812}, {M: 1750, E: -1391}, {M: 1314, E: -239}, {M: -318, E: -581}, {M: -1474, E: 667}, {M: -2105, E: 864}, {M: 486, E: 4145}, {M: -712, E: 2383},
		{M: -175, E: 828}, {M: 1601, E: -2025}, {M: 517, E: -63}, {M: 1732, E: -963}, {M: -208, E: 1552}, {M: -2700, E: 2492}, {M: 1007, E: 5541}, {M: 122, E: 3713},
		{M: 1906, E: 2919}, {M: 3304, E: -2060}, {M: 4432, E: -5123}, {M: 7167, E: -1328}, {M: 5340, E: -104}, {M: 3641, E: 1931}, {M: 8812, E: 2226}, {M: 7298, E: 1284},
		{M: -160, E: 746}, {M: 672, E: -3497}, {M: 501, E: -6772}, {M: -6509, E: -3396}, {M: -473, E: 45}, {M: -4848, E: 99}, {M: -11, E: 47}, {M: -641, E: 2432},
		{M: -600, E: 3039}, {M: 558, E: 5455}, {M: 2796, E: 8245}, {M: 8440, E: 14164}, {M: 21462, E: 18791}, {M: 58, E: 116}, {M: -4, E: -6315}, {M: -1591, E: -52},
		{M: -2972, E: 1142}, {M: -668, E: 1466}, {M: -12, E: 2058}, {M: 11970, E: 829}, {M: 14974, E: -778}, {M: 27941, E: -4722}, {M: -51, E: -5507}, {M: 2657, E: -30},
		{M: 2127, E: 1340}, {M: -4446, E: 2665}, {M: -1074, E: 1983}, {M: -3807, E: 2885}, {M: 5046, E: -587}, {M: 2643, E: -5545}, {M: 136, E: -148}, {M: 1316, E: -306},
		{M: 4394, E: -1701}, {M: 5640, E: -2135}, {M: 1297, E: 361}, {M: -1796, E: 940}, {M: -886, E: -1392}, {M: 2752, E: 191}, {M: -2537, E: 588}, {M: 3655, E: 7497},
		{M: 4506, E: 188}, {M: 1607, E: 473},
	}

	// Named chunks of Weights
	wFigure             [FigureArraySize]Score
	wMobility           [FigureArraySize]Score
	wPawn               [48]Score
	wPassedPawn         [8]Score
	wKingRank           [8]Score
	wKingFile           [8]Score
	wKingAttack         [4]Score
	wConnectedPawn      Score
	wDoublePawn         Score
	wIsolatedPawn       Score
	wPawnThreat         Score
	wKingShelter        Score
	wBishopPair         Score
	wRookOnOpenFile     Score
	wRookOnHalfOpenFile Score

	// Evaluation caches.
	pawnsAndShelterCache *cache

	// Figure bonuses to use when computing the futility margin.
	futilityFigureBonus [FigureArraySize]int32
)

func init() {
	// Initializes weights.
	initWeights()
	slice := func(w []Score, out []Score) []Score {
		copy(out, w)
		return w[len(out):]
	}
	entry := func(w []Score, out *Score) []Score {
		*out = w[0]
		return w[1:]
	}

	w := Weights[:]
	w = slice(w, wFigure[:])
	w = slice(w, wMobility[:])
	w = slice(w, wPawn[:])
	w = slice(w, wPassedPawn[:])
	w = slice(w, wKingRank[:])
	w = slice(w, wKingFile[:])
	w = slice(w, wKingAttack[:])
	w = entry(w, &wConnectedPawn)
	w = entry(w, &wDoublePawn)
	w = entry(w, &wIsolatedPawn)
	w = entry(w, &wPawnThreat)
	w = entry(w, &wKingShelter)
	w = entry(w, &wBishopPair)
	w = entry(w, &wRookOnOpenFile)
	w = entry(w, &wRookOnHalfOpenFile)

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}

	// Initialize caches.
	pawnsAndShelterCache = newCache(9, hashPawnsAndShelter, evaluatePawnsAndShelter)

	// Initializes futility figure bonus
	for i, w := range wFigure {
		if w.M >= w.E {
			futilityFigureBonus[i] = scaleToCentipawn(w.M)
		} else {
			futilityFigureBonus[i] = scaleToCentipawn(w.E)
		}
	}
}

func hashPawnsAndShelter(pos *Position, us Color) uint64 {
	h := murmurSeed[us]
	h = murmurMix(h, uint64(pos.ByPiece(us, Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us.Opposite(), Pawn)))
	h = murmurMix(h, uint64(pos.ByPiece(us, King)))
	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		// Mixes in something to signal queen's presence.
		h = murmurMix(h, murmurSeed[NoColor])
	}
	return h
}

func evaluatePawnsAndShelter(pos *Position, us Color) Eval {
	var eval Eval
	eval.Merge(evaluatePawns(pos, us))
	eval.Merge(evaluateShelter(pos, us))
	return eval
}

func evaluatePawns(pos *Position, us Color) Eval {
	var eval Eval
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)

	// From white's POV (P - white pawn, p - black pawn).
	// block   wings
	// ....... .....
	// .....P. .....
	// .....x. .....
	// ..p..x. .....
	// .xxx.x. .xPx.
	// .xxx.x. .....
	// .xxx.x. .....
	// .xxx.x. .....
	block := East(theirs) | theirs | West(theirs)
	wings := East(ours) | West(ours)
	double := Bitboard(0)
	if us == White {
		block = SouthSpan(block) | SouthSpan(ours)
		double = ours & South(ours)
	} else /* if us == Black */ {
		block = NorthSpan(block) | NorthSpan(ours)
		double = ours & North(ours)
	}

	isolated := ours &^ Fill(wings)                           // no pawn on the adjacent files
	connected := ours & (North(wings) | wings | South(wings)) // has neighbouring pawns
	passed := passedPawns(pos, us)                            // no pawn in front and no enemy on the adjacent files

	for bb := ours; bb != 0; {
		sq := bb.Pop()
		povSq := sq.POV(us)
		rank := povSq.Rank()

		eval.Add(wFigure[Pawn])
		eval.Add(wPawn[povSq-8])

		if passed.Has(sq) {
			eval.Add(wPassedPawn[rank])
		}
		if connected.Has(sq) {
			eval.Add(wConnectedPawn)
		}
		if double.Has(sq) {
			eval.Add(wDoublePawn)
		}
		if isolated.Has(sq) {
			eval.Add(wIsolatedPawn)
		}
	}

	return eval
}

func evaluateShelter(pos *Position, us Color) Eval {
	var eval Eval
	pawns := pos.ByPiece(us, Pawn)
	king := pos.ByPiece(us, King)

	sq := king.AsSquare().POV(us)
	eval.Add(wKingFile[sq.File()])
	eval.Add(wKingRank[sq.Rank()])

	if pos.ByPiece(us.Opposite(), Queen) != 0 {
		king = ForwardSpan(us, king)
		file := sq.File()
		if file > 0 && West(king)&pawns == 0 {
			eval.Add(wKingShelter)
		}
		if king&pawns == 0 {
			eval.AddN(wKingShelter, 2)
		}
		if file < 7 && East(king)&pawns == 0 {
			eval.Add(wKingShelter)
		}
	}
	return eval
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	eval.Merge(pawnsAndShelterCache.load(pos, us))
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	theirPawns := pos.ByPiece(them, Pawn)
	theirKing := pos.ByPiece(them, King)
	theirKingArea := bbKingArea[theirKing.AsSquare()]
	numAttackers := 0
	attackStrength := int32(0)

	// Pawn forward mobility.
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	eval.AddN(wMobility[Pawn], mobility.Count())
	mobility = pos.PawnThreats(us)
	eval.AddN(wPawnThreat, (mobility & pos.ByColor[them]).Count())

	// Knight
	excl := pos.ByPiece(us, Pawn) | pos.PawnThreats(them)
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Knight])
		mobility := KnightMobility(sq)
		eval.AddN(wMobility[Knight], (mobility &^ excl).Count())

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Bishop])
		mobility := BishopMobility(sq, all)
		eval.AddN(wMobility[Bishop], (mobility &^ excl).Count())
		numBishops++

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	eval.AddN(wBishopPair, numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Rook])
		mobility := RookMobility(sq, all)
		eval.AddN(wMobility[Rook], (mobility &^ excl).Count())

		// Evaluate rook on open and semi open files.
		// https://chessprogramming.wikispaces.com/Rook+on+Open+File
		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				eval.Add(wRookOnOpenFile)
			} else {
				eval.Add(wRookOnHalfOpenFile)
			}
		}

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Queen])
		mobility := QueenMobility(sq, all) &^ excl
		eval.AddN(wMobility[Queen], (mobility &^ excl).Count())

		if a := mobility & theirKingArea &^ theirPawns; a != 0 {
			numAttackers++
			attackStrength += a.CountMax2()
		}
	}

	// King, each side has one.
	{
		sq := pos.ByPiece(us, King).AsSquare()
		mobility := KingMobility(sq) &^ excl
		eval.AddN(wMobility[King], mobility.Count())
	}

	// Evaluate attacking the king
	if numAttackers >= len(wKingAttack) {
		numAttackers = len(wKingAttack) - 1
	}
	eval.AddN(wKingAttack[numAttackers], attackStrength)
}

// evaluatePosition evalues position.
func EvaluatePosition(pos *Position) Eval {
	var eval Eval
	evaluateSide(pos, Black, &eval)
	eval.Neg()
	evaluateSide(pos, White, &eval)
	return eval
}

// Evaluate evaluates position from White's POV.
// Scores fits into a int16.
func Evaluate(pos *Position) int32 {
	eval := EvaluatePosition(pos)
	score := eval.Feed(Phase(pos))
	return scaleToCentipawn(score)
}

// Phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func Phase(pos *Position) int32 {
	total := int32(4*1 + 4*1 + 4*3 + 2*6)
	curr := total
	curr -= pos.ByFigure[Knight].Count() * 1
	curr -= pos.ByFigure[Bishop].Count() * 1
	curr -= pos.ByFigure[Rook].Count() * 3
	curr -= pos.ByFigure[Queen].Count() * 6
	return (curr*256 + total/2) / total
}

// passedPawns returns all passed pawns of us in pos.
func passedPawns(pos *Position, us Color) Bitboard {
	ours := pos.ByPiece(us, Pawn)
	theirs := pos.ByPiece(us.Opposite(), Pawn)
	theirs |= East(theirs) | West(theirs)
	block := BackwardSpan(us, theirs|ours)
	return ours &^ block
}

// scaleToCentipawn scales a score in the original scale to centipawns.
func scaleToCentipawn(score int32) int32 {
	// Divides by 128 and rounds to the nearest integer.
	return (score + score>>31<<6) >> 7
}
