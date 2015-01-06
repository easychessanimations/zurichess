//go:generate go tool yacc epd_parser.y -o epd_parser.go
// epd.go implements parsing of chess positions in FEN and EPD notations.
package engine

// Extended Position Description
type EPD struct {
	Position *Position
	Id       string
}

func parse(what int, line string) (*EPD, error) {
	lex := &epdLexer{
		what:   what,
		line:   line,
		pos:    -1,
		result: new(*epdNode),
	}
	if yyParse(lex) != 0 {
		return nil, lex.error
	}
	epd := &EPD{}
	if err := handleEPDNode(epd, *lex.result); err != nil {
		return nil, err
	}
	return epd, nil
}

// ParseFEN parses a FEN string.
func ParseFEN(line string) (*EPD, error) {
	return parse(_hiddenFEN, line)

}

// ParseEPD parses a EPD string.
func ParseEPD(line string) (*EPD, error) {
	return parse(_hiddenEPD, line)
}

// Same as ParseFEN, but returns only the position.
// Mostly useful for testing.
func PositionFromFEN(fen string) (*Position, error) {
	epd, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	return epd.Position, nil
}
