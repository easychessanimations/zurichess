package main

// LSB returns the least significant bit of n.
func LSB(n uint64) uint64 {
	return n & (-n)
}

var debrujin64 = [64]uint{
	0, 1, 2, 7, 3, 13, 8, 19, 4, 25, 14, 28, 9, 34, 20, 40,
	5, 17, 26, 38, 15, 46, 29, 48, 10, 31, 35, 54, 21, 50, 41, 57,
	63, 6, 12, 18, 24, 27, 33, 39, 16, 37, 45, 47, 30, 53, 49, 56,
	62, 11, 23, 32, 36, 44, 52, 55, 61, 22, 43, 51, 60, 42, 59, 58,
}

// LogN returns the logarithm of n, where n is a power of two.
func LogN(n uint64) uint {
	return debrujin64[(n*0x218A392CD3D5DBF)>>58&0x3F]
}

// Popcnt counts number of bits set in n.
func Popcnt(n uint64) (c uint) {
	for ; n > 0; c++ {
		n -= LSB(n)
	}
	return c
}
