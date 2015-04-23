// Package notation implements parsing of chess positions.
//
// Current supported formats are FEN and EPD notations.
package notation

import (
	"bitbucket.org/brtzsnr/zurichess/engine"
)

// Extended Position Description
type EPD struct {
	Position *engine.Position
	Id       string
	BestMove []engine.Move
	Comment  map[string]string
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
	epd := &EPD{
		Comment: make(map[string]string),
	}
	if err := handleEPDNode(epd, *lex.result); err != nil {
		return nil, err
	}
	return epd, nil
}

// ParseFEN parses a FEN string and returns an EPD.
func ParseFEN(line string) (*EPD, error) {
	return parse(_hiddenFEN, line)

}

// ParseEPD parses a EPD string and returns an EPD.
func ParseEPD(line string) (*EPD, error) {
	return parse(_hiddenEPD, line)
}

func (e *EPD) String() string {
	s := engine.FormatPiecePlacement(e.Position)
	s += " " + engine.FormatSideToMove(e.Position)
	s += " " + engine.FormatCastlingAbility(e.Position)
	s += " " + engine.FormatEnpassantSquare(e.Position)

	for _, bm := range e.BestMove {
		s += " bm " + bm.LAN() + ";"
	}
	if e.Id != "" {
		s += " id \"" + e.Id + "\";"
	}
	for k, v := range e.Comment {
		s += " " + k + " \"" + v + "\";"
	}
	return s
}
