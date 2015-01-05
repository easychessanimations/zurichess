//go:generate go tool yacc epd_parser.y -o epd_parser.go
// epd.go implements parsing of chess positions in FEN and EPD notations.
package engine

// Extended Position Description
type EPD struct {
	Position *Position
}
