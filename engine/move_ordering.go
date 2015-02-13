package engine

import "sort"

var (
	// mvvLva table stores the ordering cores.
	//
	// MVV/LVA stands for "Most valuable victim, Least valuable attacker".
	// See https://chessprogramming.wikispaces.com/MVV-LVA.
	//
	// In zurichess the MVV/LVA formula is not used,
	// but the values are optimized and stored in this array.
	//
	// mvvLva[attacker * FigureSize + victim]
	mvvLva = [FigureArraySize * FigureArraySize]int{
		0, 369, 902, 1432, 2102, 2534, 20000, // Promotion
		0, 1017, 1151, 1735, 2093, 2146, 20000, // Pawn
		0, 454, 1213, 1602, 2410, 2973, 20000, // Knight
		0, 447, 641, 1340, 1906, 2740, 20000, // Bishop
		0, 24, 599, 1174, 1737, 2565, 20000, // Rook
		0, 81, 521, 1074, 1604, 1972, 20000, // Queen
		0, 981, 1815, 1839, 2673, 3391, 20000, // King
	}
)

// SetMvvLva sets the MVV/LVA table.
func SetMvvLva(str string) error {
	return SetMaterialValue("MvvLva", mvvLva[:], str)
}

// sorterByMvvLva implements sort.Interface.
// Compares moves by Most Valuable Victim / Least Valuable Aggressor
// https://chessprogramming.wikispaces.com/MVV-LVA
type sorterByMvvLva []Move

func ml(a, v Figure) int {
	return mvvLva[FigureArraySize*int(a)+int(v)]
}

func score(m *Move) int {
	s := ml(m.Piece().Figure(), m.Capture.Figure())
	s += ml(NoFigure, m.Promotion().Figure())
	s += int(m.Data) << 5
	return s
}

func (c sorterByMvvLva) Len() int {
	return len(c)
}

func (c sorterByMvvLva) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c sorterByMvvLva) Less(i, j int) bool {
	return score(&c[i]) < score(&c[j])
}

func sortMoves(moves []Move) {
	sort.Sort(sorterByMvvLva(moves))
}
