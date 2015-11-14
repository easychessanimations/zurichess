// material.go implements position evaluation.

package engine

import (
	"fmt"
)

const (
	KnownWinScore  int32 = 25000000       // All scores strictly greater than KnownWinScore are sure wins.
	KnownLossScore int32 = -KnownWinScore // All scores strictly lower than KnownLossScore are sure losses.
	MateScore      int32 = 30000000       // MateScore - N is mate env N plies.
	MatedScore     int32 = -MateScore     // MatedScore + N is mated env N plies.
	InfinityScore  int32 = 32000000       // Maximum possible score. -InfinityScore is the minimum possible score.
)

var (
	// All weights under one array for easy handling.
	Weights = [94]Score{
		{M: 2431, E: 7805}, {M: 3369, E: 7452}, {M: 36173, E: 34838}, {M: 39259, E: 38871}, {M: 53103, E: 69595}, {M: 123802, E: 131197}, {M: 406, E: 6357}, {M: 9031, E: 6890}, {M: 622, E: 2708}, {M: 997, E: 1066}, {M: 646, E: 814}, {M: 750, E: 609}, {M: 237, E: 751}, {M: -287, E: -853}, {M: 3195, E: 5401}, {M: 2247, E: 5584}, {M: 3203, E: 3837}, {M: 4660, E: 2442}, {M: 3604, E: 2670}, {M: 5178, E: 3481}, {M: 4141, E: 2897}, {M: 2304, E: 2939}, {M: 3280, E: 5129}, {M: 2221, E: 5092}, {M: 3185, E: 3768}, {M: 3486, E: 2789}, {M: 4399, E: 3857}, {M: 3284, E: 4260}, {M: 4549, E: 3521}, {M: 2238, E: 4602}, {M: 2992, E: 7182}, {M: 2150, E: 5677}, {M: 3619, E: 3869}, {M: 5208, E: 2978}, {M: 4125, E: 3835}, {M: 2632, E: 3505}, {M: 1939, E: 4242}, {M: 1586, E: 5098}, {M: 2978, E: 9597}, {M: 1860, E: 7826}, {M: 3636, E: 4610}, {M: 4263, E: 2804}, {M: 3136, E: 4561}, {M: 2907, E: 4046}, {M: 1798, E: 5941}, {M: 1246, E: 6815}, {M: 4855, E: 11301}, {M: 2430, E: 10423}, {M: 6460, E: 4248}, {M: 6151, E: 2033}, {M: 5478, E: 1230}, {M: 7162, E: 3938}, {M: 3342, E: 5717}, {M: 5218, E: 8633}, {M: 15343, E: 11598}, {M: 6336, E: 15023}, {M: -511, E: 12835}, {M: 7580, E: 8307}, {M: 8351, E: 8458}, {M: 10391, E: 8442}, {M: 7963, E: 11608}, {M: 4681, E: 10949}, {M: 8373, E: 822}, {M: 314, E: 2349}, {M: 471, E: 2864}, {M: -48, E: 6430}, {M: 2719, E: 9089}, {M: 4223, E: 15949}, {M: 6367, E: 13989}, {M: 4610, E: 5728}, {M: 7115, E: 2400}, {M: 5449, E: 8853}, {M: 4772, E: 10104}, {M: 710, E: 11554}, {M: -4969, E: 14093}, {M: -254, E: 14526}, {M: 19559, E: 10793}, {M: 28773, E: 3360}, {M: 3456, E: -1989}, {M: 4150, E: 4099}, {M: 2295, E: 5317}, {M: -2695, E: 7089}, {M: -286, E: 6430}, {M: -3191, E: 6920}, {M: 3866, E: 3962}, {M: 2331, E: -120}, {M: 1030, E: 227}, {M: -595, E: 798}, {M: -833, E: -923}, {M: 1580, E: 598}, {M: -1943, E: -650}, {M: 3154, E: 6134}, {M: 3220, E: 47}, {M: 607, E: 973},
	}

	// Named chunks of Weights
	wFigure     [FigureArraySize]Score
	wMobility   [FigureArraySize]Score
	wPawn       [48]Score
	wPassedPawn [8]Score
	wKingRank   [8]Score
	wKingFile   [8]Score
	wFlags      [8]Score // see flags defined below
)

const (
	fConnectedPawn = iota
	fDoublePawn
	fIsolatedPawn
	fPawnThreat
	fKingShelter
	fBishopPair
	fRookOnOpenFile
	fRookOnHalfOpenFile
)

func init() {
	initWeights()

	chunk := func(w []Score, out []Score) []Score {
		copy(out, w)
		return w[len(out):]
	}

	w := Weights[:]
	w = chunk(w, wFigure[:])
	w = chunk(w, wMobility[:])
	w = chunk(w, wPawn[:])
	w = chunk(w, wPassedPawn[:])
	w = chunk(w, wKingRank[:])
	w = chunk(w, wKingFile[:])
	w = chunk(w, wFlags[:])

	if len(w) != 0 {
		panic(fmt.Sprintf("not all weights used, left with %d out of %d", len(w), len(Weights)))
	}
}

func evaluatePawns(pos *Position, us Color, eval *Eval) {
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
	passed := ours &^ block                                   // no pawn env front and no enemy on the adjacent files

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
			eval.Add(wFlags[fConnectedPawn])
		}
		if double.Has(sq) {
			eval.Add(wFlags[fDoublePawn])
		}
		if isolated.Has(sq) {
			eval.Add(wFlags[fIsolatedPawn])
		}
	}
}

// evaluateSide evaluates position for a single side.
func evaluateSide(pos *Position, us Color, eval *Eval) {
	evaluatePawnsCached(pos, us, eval)
	all := pos.ByColor[White] | pos.ByColor[Black]
	them := us.Opposite()

	// Pawn
	mobility := Forward(us, pos.ByPiece(us, Pawn)) &^ all
	eval.AddN(wMobility[Pawn], mobility.Popcnt())
	mobility = pos.PawnThreats(us) & pos.ByColor[us.Opposite()]
	eval.AddN(wFlags[fPawnThreat], mobility.Popcnt())

	// Knight
	excl := pos.ByPiece(us, Pawn) | pos.PawnThreats(them)
	for bb := pos.ByPiece(us, Knight); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Knight])
		mobility := KnightMobility(sq) &^ excl
		eval.AddN(wMobility[Knight], mobility.Popcnt())
	}
	// Bishop
	numBishops := int32(0)
	for bb := pos.ByPiece(us, Bishop); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Bishop])
		mobility := BishopMobility(sq, all) &^ excl
		eval.AddN(wMobility[Bishop], mobility.Popcnt())
		numBishops++
	}
	eval.AddN(wFlags[fBishopPair], numBishops/2)

	// Rook
	for bb := pos.ByPiece(us, Rook); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Rook])
		mobility := RookMobility(sq, all) &^ excl
		eval.AddN(wMobility[Rook], mobility.Popcnt())

		f := FileBb(sq.File())
		if pos.ByPiece(us, Pawn)&f == 0 {
			if pos.ByPiece(them, Pawn)&f == 0 {
				eval.Add(wFlags[fRookOnOpenFile])
			} else {
				eval.Add(wFlags[fRookOnHalfOpenFile])
			}
		}
	}
	// Queen
	for bb := pos.ByPiece(us, Queen); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[Queen])
		mobility := QueenMobility(sq, all) &^ excl
		eval.AddN(wMobility[Queen], mobility.Popcnt())
	}

	// King
	for bb := pos.ByPiece(us, King); bb > 0; {
		sq := bb.Pop()
		eval.Add(wFigure[King])
		mobility := KingMobility(sq) &^ excl
		eval.AddN(wMobility[King], mobility.Popcnt())

		sq = sq.POV(us)
		eval.Add(wKingFile[sq.File()])
		eval.Add(wKingRank[sq.Rank()])
	}

	if pos.ByPiece(them, Queen) != 0 {
		pawns := pos.ByPiece(us, Pawn)
		king := pos.ByPiece(us, King)
		file := king.AsSquare().File()

		// TODO: Should we include adjacent pawns in the computation?
		if us == White {
			king = NorthSpan(king)
		} else /* if us == Black */ {
			king = SouthSpan(king)
		}

		if file > 0 && West(king)&pawns == 0 {
			eval.Add(wFlags[fKingShelter])
		}
		if king&pawns == 0 {
			eval.AddN(wFlags[fKingShelter], 2)
		}
		if file < 7 && East(king)&pawns == 0 {
			eval.Add(wFlags[fKingShelter])
		}
	}
}

// evaluatePosition evalues position.
func EvaluatePosition(pos *Position, eval *Eval) {
	eval.Make(pos)
	evaluateSide(pos, Black, eval)
	eval.Neg()
	evaluateSide(pos, White, eval)
}

// Evaluate evaluates position from White's POV.
func Evaluate(pos *Position) int32 {
	var env Eval
	EvaluatePosition(pos, &env)
	score := env.Feed()
	if KnownLossScore >= score || score >= KnownWinScore {
		panic(fmt.Sprintf("score %d should be between %d and %d",
			score, KnownLossScore, KnownWinScore))
	}
	return score
}

// phase computes the progress of the game.
// 0 is opening, 256 is late end game.
func phase(pos *Position) int32 {
	total := int32(16*0 + 4*1 + 4*1 + 4*2 + 2*4)
	curr := total
	curr -= pos.ByFigure[Pawn].Popcnt() * 0
	curr -= pos.ByFigure[Knight].Popcnt() * 1
	curr -= pos.ByFigure[Bishop].Popcnt() * 1
	curr -= pos.ByFigure[Rook].Popcnt() * 2
	curr -= pos.ByFigure[Queen].Popcnt() * 4
	return (curr*256 + total/2) / total
}
