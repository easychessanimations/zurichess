%{
// epd_parser.y defines the grammar for chess positions in EPD and FEN notations.
// For EPD see https://chessprogramming.wikispaces.com/Extended+Position+Description
// For FEN see https://chessprogramming.wikispaces.com/Forsyth-Edwards+Notation
package engine

import (
        "fmt"
        "unicode"
        "log"
)

%}

%union{
        result    **epdNode
        epd       *epdNode
        position  *positionNode
        operation *operationNode
        token     *tokenNode
}

%type <epd> epd fen
%type <position> position
%type <token> piecePlacement castlingAbility enpassantSquare sideToMove

%token <token> _token
%token <result> _hiddenFEN _hiddenEPD

%%

top
        : _hiddenFEN fen
        { *$1 = $2 }
        | _hiddenEPD epd
        { *$1 = $2 }
        ;

fen
        : position ' ' _token ' ' _token
        {
                $$ = &epdNode{
                        position: $1,
                }
        }
        ;

epd
        : position ';'
        {
                $$ = &epdNode{
                        position: $1,
                }
        }
        ;

position
        : piecePlacement ' ' sideToMove ' ' castlingAbility ' ' enpassantSquare
        {
                $$ = &positionNode{
                        piecePlacement:  $1,
                        sideToMove:      $3,
                        castlingAbility: $5,
                        enpassantSquare: $7,
                }
        }
        ;

piecePlacement
        : _token
        { $$ = $1 }
        ;

sideToMove
        : _token
        { $$ = $1 }
        ;

castlingAbility
        : _token
        { $$ = $1 }
        ;

enpassantSquare
        : _token
        { $$ = $1 }
        ;

%%

const eof = 0

// epdLexer is a tokenizer.
type epdLexer struct {
	what      int // FEN or EPD
	line      string
	prev, pos int
	result    **epdNode
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

	lval.token = &tokenNode{
		pos: start,
		str: lex.line[start:lex.pos],
	}
	return _token
}

func (lex *epdLexer) Error(s string) {
	lex.error = fmt.Errorf("error at %d [%s]%s: %s\n",
		lex.prev, lex.line[:lex.prev], lex.line[lex.prev:], s)
	lex.pos = len(lex.line) // next call Lex() will return eof.
	log.Println(lex.error)
}

