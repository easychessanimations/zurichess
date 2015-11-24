// see.go implements static exchange evaluation.

package engine

// piece bonuses when calulating the see.
// The values are fixed to approximatively the figure bonus in mid game.
var seeScore = [FigureArraySize]int32{0, 55, 325, 341, 454, 1110, 20000}

// seeSign return true if see(m) < 0.
func seeSign(pos *Position, m Move) bool {
	if m.Piece().Figure() <= m.Capture().Figure() {
		// Even if m.Piece() is captured, we are still positive.
		return false
	}
	return see(pos, m) < 0
}

// see returns the static exchange evaluation for m.
//
// https://chessprogramming.wikispaces.com/Static+Exchange+Evaluation
// https://chessprogramming.wikispaces.com/SEE+-+The+Swap+Algorithm
//
// The implementation here is optimized for the common case when there
// isn't any capture following the move. The score returned is based
// on some fixed values for figures, different from the ones
// defined in material.go.
func see(pos *Position, m Move) int32 {
	us := pos.SideToMove
	them := us.Opposite()
	sq := m.To()
	bb := sq.Bitboard()
	bb27 := bb &^ (BbRank1 | BbRank8)
	bb18 := bb & (BbRank1 | BbRank8)

	// Occupancy tables as if moves are executed.
	var occ [ColorArraySize]Bitboard
	occ[White] = pos.ByColor[White] &^ bb
	occ[Black] = pos.ByColor[Black] &^ bb
	all := occ[White] | occ[Black]

	gain := make([]int32, 0, 4)
	for score := int32(0); score >= 0; {
		// m is the last move executed.
		// Adjust score for current player.
		score = -score
		score += seeScore[m.Capture().Figure()]
		if m.MoveType() == Promotion {
			score -= seeScore[Pawn]
			score += seeScore[m.Target().Figure()]
		}
		gain = append(gain, score)

		// Update occupancy tables for executing the move.
		occ[us] = occ[us] &^ m.From().Bitboard()
		all = all &^ m.From().Bitboard()

		// Switch sides.
		us, them = them, us
		ours := occ[us]

		var fig Figure                  // attacking figure
		var att Bitboard                // attackers
		var pawn, bishop, rook Bitboard // mobilies for our figures

		// Try every figure in order of value.
		mt := Normal

		// Pawn attacks.
		pawn = Backward(us, West(bb27)|East(bb27))
		fig, att = Pawn, pawn&ours&pos.ByFigure[Pawn]
		if att != 0 {
			goto makeMove
		}

		fig, att = Knight, bbKnightAttack[sq]&ours&pos.ByFigure[Knight]
		if att != 0 {
			goto makeMove
		}

		if bbSuperAttack[sq]&ours == 0 {
			// No other can attack sq so we give up early.
			break
		}

		bishop = BishopMobility(sq, all)
		fig, att = Bishop, bishop&ours&pos.ByFigure[Bishop]
		if att != 0 {
			goto makeMove
		}

		rook = RookMobility(sq, all)
		fig, att = Rook, rook&ours&pos.ByFigure[Rook]
		if att != 0 {
			goto makeMove
		}

		// Pawn promotions are considered queens minus the pawn.
		pawn = Backward(us, West(bb18)|East(bb18))
		fig, att = Queen, pawn&ours&pos.ByFigure[Pawn]
		if att != 0 {
			mt = Promotion
			goto makeMove
		}

		fig, att = Queen, (rook|bishop)&ours&pos.ByFigure[Queen]
		if att != 0 {
			goto makeMove
		}

		fig, att = King, bbKingAttack[sq]&ours&pos.ByFigure[King]
		if att != 0 {
			goto makeMove
		}

		// No attack found.
		break

	makeMove:
		if att != 0 {
			// Make a new pseudo-legal move of the smallest attacker.
			m = MakeMove(mt, att.Pop(), sq, m.Target(), ColorFigure(us, fig))
		}
	}

	for i := len(gain) - 2; i >= 0; i-- {
		if -gain[i+1] < gain[i] {
			gain[i] = -gain[i+1]
		}
	}
	return gain[0]
}
