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
		250, 254, 535, 757, 919, 1283, 20000, // Promotion
		250, 863, 1380, 1779, 2307, 2814, 20000, // Pawn
		250, 781, 1322, 1654, 1766, 2414, 20000, // Knight
		250, 409, 810, 1411, 2170, 3000, 20000, // Bishop
		250, 393, 1062, 1199, 2117, 2988, 20000, // Rook
		250, 349, 948, 1355, 1631, 2314, 20000, // Queen
		250, 928, 1088, 1349, 1593, 2417, 20000, // King
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
