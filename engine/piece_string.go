// generated by stringer -type Piece; DO NOT EDIT

package engine

import "fmt"

const (
	_Piece_name_0 = "WhitePawnBlackPawn"
	_Piece_name_1 = "WhiteKnightBlackKnight"
	_Piece_name_2 = "WhiteBishopBlackBishop"
	_Piece_name_3 = "WhiteRookBlackRook"
	_Piece_name_4 = "WhiteQueenBlackQueen"
	_Piece_name_5 = "WhiteKingBlackKing"
)

var (
	_Piece_index_0 = [...]uint8{9, 18}
	_Piece_index_1 = [...]uint8{11, 22}
	_Piece_index_2 = [...]uint8{11, 22}
	_Piece_index_3 = [...]uint8{9, 18}
	_Piece_index_4 = [...]uint8{10, 20}
	_Piece_index_5 = [...]uint8{9, 18}
)

func (i Piece) String() string {
	switch {
	case 5 <= i && i <= 6:
		i -= 5
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_0[i-1]
		}
		return _Piece_name_0[lo:_Piece_index_0[i]]
	case 9 <= i && i <= 10:
		i -= 9
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_1[i-1]
		}
		return _Piece_name_1[lo:_Piece_index_1[i]]
	case 13 <= i && i <= 14:
		i -= 13
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_2[i-1]
		}
		return _Piece_name_2[lo:_Piece_index_2[i]]
	case 17 <= i && i <= 18:
		i -= 17
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_3[i-1]
		}
		return _Piece_name_3[lo:_Piece_index_3[i]]
	case 21 <= i && i <= 22:
		i -= 21
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_4[i-1]
		}
		return _Piece_name_4[lo:_Piece_index_4[i]]
	case 25 <= i && i <= 26:
		i -= 25
		lo := uint8(0)
		if i > 0 {
			lo = _Piece_index_5[i-1]
		}
		return _Piece_name_5[lo:_Piece_index_5[i]]
	default:
		return fmt.Sprintf("Piece(%d)", i)
	}
}