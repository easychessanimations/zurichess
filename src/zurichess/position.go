package main

// Square identifies the location on the board.
type Square int

// Rank returns a number 0...7 representing the rank of the square.
func (s Square) Rank() int {
	return int(s / 8)
}

// File returns a number 0...7 representing the file of the square.
func (s Square) File() int {
	return int(s % 8)
}

// Piece represents a colorless piece
type Piece int

// Color represents a color.
type Color int

// A birboard 8x8.
type Bitboard uint64
