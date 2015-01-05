%{
// epd_parser.y implements parsing of chess positions in EPD and FEN notations.
// For EPD see https://chessprogramming.wikispaces.com/Extended+Position+Description
// For FEN see https://chessprogramming.wikispaces.com/Forsyth-Edwards+Notation
package engine

import (
	"fmt"
	"log"
	"strings"
	"unicode"
)

%}

%union{
        str      string
        result   **EPD
        epd      *EPD
        position *Position
        castling Castle
        square   Square
        toMove   Color
}

%type <epd> epd fen
%type <position> position
%type <position> piecePlacement
%type <castling> castlingAbility
%type <square> enpassantSquare
%type <toMove> sideToMove

%token <str> OPERAND
%token <result> HIDDEN_FEN HIDDEN_EPD

%%

top
        : HIDDEN_FEN fen
        { *$1 = $2 }
        | HIDDEN_EPD epd
        { *$1 = $2 }
        ;

fen
        : position ' ' OPERAND ' ' OPERAND
        {
                $$ = &EPD{
                        Position: $1,
                }
        }
        ;

epd
        : position ';'
        {
                $$ = &EPD{
                        Position: $1,
                }
        }
        ;

position
        : piecePlacement ' ' sideToMove ' ' castlingAbility ' ' enpassantSquare
        {
                $$ = $1
                $$.ToMove = $3
                $$.Castle = $5
                $$.Enpassant = $7
        }
        ;

piecePlacement
        : OPERAND
        { $$ = parsePiecePlacement($1) }
        ;

sideToMove
        : OPERAND
        { $$ = parseSideToMove($1) }
        ;

castlingAbility
        : OPERAND
        { $$ = parseCastlingAbility($1) }
        ;

enpassantSquare
        : OPERAND
        { $$ = parseEnpassantSquare($1) }
        ;

%%

const eof = 0

var (
	symbolToPiece = map[rune]Piece{
		'p': BlackPawn,
		'n': BlackKnight,
		'b': BlackBishop,
		'r': BlackRook,
		'q': BlackQueen,
		'k': BlackKing,

		'P': WhitePawn,
		'N': WhiteKnight,
		'B': WhiteBishop,
		'R': WhiteRook,
		'Q': WhiteQueen,
		'K': WhiteKing,
	}

	symbolToCastle = map[rune]Castle{
		'K': WhiteOO,
		'Q': WhiteOOO,
		'k': BlackOO,
		'q': BlackOOO,
	}
)

// Parse position.
func parsePiecePlacement(str string) *Position {
	pos := &Position{}

	ranks := strings.Split(str, "/")
	if len(ranks) != 8 {
		// TODO: handle error
		return nil
	}
	for r := range ranks {
		sq := RankFile(7-r, 0) // FEN describes the table from 8th rank.
		for _, p := range ranks[r] {
			pi := symbolToPiece[p]
			if pi == NoPiece {
				if '1' <= p && p <= '8' {
					sq = sq.Relative(0, int(p)-int('0')-1)
				} else {
					// TODO: handle error
					return nil
				}
			}
			pos.Put(sq, pi)
			sq = sq.Relative(0, 1)
		}
	}
	return pos
}

func parseSideToMove(str string) Color {
	if str == "w" {
		return White
	}
	if str == "b" {
		return Black
	}
	return NoColor
}

func parseCastlingAbility(str string) Castle {
	if str == "-" {
		return NoCastle
	}

	ability := NoCastle
	for _, p := range str {
		// If p is an invalid symbol, this does nothing.
		ability |= symbolToCastle[p]
	}
	return ability
}

// TODO: handle error
func parseEnpassantSquare(str string) Square {
	if str[:1] == "-" {
		return SquareA1
	}
	return SquareFromString(str)
}

type epdLexer struct {
	what      int // FEN or EPD
	result    **EPD
	line      string
	prev, pos int
	error     error
}

func (lex *epdLexer) Lex(lval *yySymType) int {
	lex.prev = lex.pos
	if lex.pos == -1 {
		// First we return a token to identify what to parse:
		// a FEN or an EPD.
		lex.pos++
                lval.result = lex.result
		return lex.what
	}

	size := len(lex.line)
	if lex.pos == size {
		return eof
	}

	// Compress spaces to a single ' '.
	c := rune(lex.line[lex.pos])
	if unicode.IsSpace(c) {
		for lex.pos < size && unicode.IsSpace(c) {
			lex.pos++
			c = rune(lex.line[lex.pos])
		}
		return ' '
	}

	// Handle ';'.
	if c == ';' {
		lex.pos++
		return ';'
	}

	// Handle tokens.
	start := lex.pos
	for lex.pos < size && !unicode.IsSpace(c) && c != ';' {
		lex.pos++
		if lex.pos < size {
			c = rune(lex.line[lex.pos])
		}
	}

	lval.str = lex.line[start:lex.pos]
	return OPERAND
}

func (lex *epdLexer) Error(s string) {
	lex.error = fmt.Errorf("error at %d [%s]%s: %s\n",
		lex.prev, lex.line[:lex.prev], lex.line[lex.prev:], s)
	lex.pos = len(lex.line) // next call Lex() will return eof.
	log.Println(lex.error)
}

// ParseFEN parses a FEN string.
func ParseFEN(line string) (*EPD, error) {
	lex := &epdLexer{
		what:   HIDDEN_FEN,
		result: new(*EPD),
		line:   line,
		pos:    -1,
	}

	yyParse(lex)
	return *lex.result, lex.error
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

// ParseEPD parses a EPD string.
func ParseEPD(line string) (*EPD, error) {
	lex := &epdLexer{
		what:   HIDDEN_EPD,
		result: new(*EPD),
		line:   line,
		pos:    -1,
	}

	yyParse(lex)
	return *lex.result, lex.error
}

