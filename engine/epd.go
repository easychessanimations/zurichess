// epd.go implements parsing of chess positions in FEN and EPD notations.

package engine

// Extended Position Description
type EPD struct {
	Position *Position
	Id       string
	BestMove []Move
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

// PositionFromFEN parses a FEN string and returns the position.
// Mostly useful for testing.
func PositionFromFEN(fen string) (*Position, error) {
	epd, err := ParseFEN(fen)
	if err != nil {
		return nil, err
	}
	return epd.Position, nil
}

var (
	itoa = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
)

// formatPiecePlacement converts a position to FEN piece placement.
func formatPiecePlacement(pos *Position) string {
	s := ""
	for r := 7; r >= 0; r-- {
		space := 0
		for f := 0; f < 8; f++ {
			sq := RankFile(r, f)
			pi := pos.Get(sq)
			if pi == NoPiece {
				space++
			} else {
				if space != 0 {
					s += itoa[space]
					space = 0
				}
				s += string(pieceToSymbol[pi])
			}
		}

		if space != 0 {
			s += itoa[space]
		}
		if r != 0 {
			s += "/"
		}
	}
	return s
}

func (e *EPD) String() string {
	s := formatPiecePlacement(e.Position)
	s += " " + colorToSymbol[e.Position.SideToMove]
	s += " " + e.Position.CastlingAbility().String()
	if e.Position.EnpassantSquare() == SquareA1 {
		s += " -"
	} else {
		s += " " + e.Position.EnpassantSquare().String()
	}

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
