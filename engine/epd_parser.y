%{
// epd_parser.y defines the grammar for chess positions in EPD and FEN notations.
// 
// For EPD format see https://chessprogramming.wikispaces.com/Extended+Position+Description.
// For FEN format see https://chessprogramming.wikispaces.com/Forsyth-Edwards+Notation.

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
        argument  *argumentNode
        token     *tokenNode
}

%type <epd> epd fen
%type <position> position
%type <operation> operations
%type <argument> arguments

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
        : position operations
        {
                $$ = &epdNode{
                        position:   $1,
                        operations: $2,
                }
        }
        ;

position
        : _token ' ' _token ' ' _token ' ' _token
        {
                $$ = &positionNode{
                        piecePlacement:  $1,
                        sideToMove:      $3,
                        castlingAbility: $5,
                        enpassantSquare: $7,
                }
        }
        ;

operations
        :
        { $$ = nil }
        | operations ' ' _token arguments ';'
        {
                $$ = &operationNode{
                       operator:  $3,
                       arguments: $4,
                }
                if $1 != nil {
                        $1.next = $$
                        $$ = $1
                }
        }
        ;

arguments
        :
        { $$ = nil }
        | arguments ' ' _token
        {
                $$ = &argumentNode{
                        param: $3,
                }
                if $1 != nil {
                        $1.next = $$
                        $$ = $1
                }
        }
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

// TODO: Handle spaces between quotes, e.g. "foo bar".
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

