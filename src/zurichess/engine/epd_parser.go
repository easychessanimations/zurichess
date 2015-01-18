//line epd_parser.y:2

// epd_parser.y defines the grammar for chess positions in EPD and FEN notations.
//
// For EPD format see https://chessprogramming.wikispaces.com/Extended+Position+Description.
// For FEN format see https://chessprogramming.wikispaces.com/Forsyth-Edwards+Notation.

package engine

import __yyfmt__ "fmt"

//line epd_parser.y:7
import (
	"fmt"
	"log"
	"unicode"
)

//line epd_parser.y:17
type yySymType struct {
	yys       int
	result    **epdNode
	epd       *epdNode
	position  *positionNode
	operation *operationNode
	argument  *argumentNode
	token     *tokenNode
}

const _token = 57346
const _hiddenFEN = 57347
const _hiddenEPD = 57348

var yyToknames = []string{
	"_token",
	"_hiddenFEN",
	"_hiddenEPD",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line epd_parser.y:105
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

//line yacctab:1
var yyExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 10
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 25

var yyAct = []int{

	23, 22, 21, 16, 15, 14, 10, 9, 2, 3,
	25, 24, 19, 18, 17, 13, 12, 6, 5, 1,
	20, 11, 8, 4, 7,
}
var yyPact = []int{

	3, -1000, 13, 13, -1000, 0, -1, -1000, -1000, 12,
	11, -2, -3, -4, 10, 9, 8, -1000, -1000, -5,
	-7, 7, -1000, 6, -1000, -1000,
}
var yyPgo = []int{

	0, 24, 23, 18, 21, 20, 19,
}
var yyR1 = []int{

	0, 6, 6, 2, 1, 3, 4, 4, 5, 5,
}
var yyR2 = []int{

	0, 2, 2, 5, 2, 7, 0, 5, 0, 3,
}
var yyChk = []int{

	-1000, -6, 5, 6, -2, -3, 4, -1, -3, 7,
	7, -4, 4, 4, 7, 7, 7, 4, 4, 4,
	-5, 7, 8, 7, 4, 4,
}
var yyDef = []int{

	0, -2, 0, 0, 1, 0, 0, 2, 6, 0,
	0, 4, 0, 0, 0, 0, 0, 8, 3, 0,
	0, 0, 7, 0, 5, 9,
}
var yyTok1 = []int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 7, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 8,
}
var yyTok2 = []int{

	2, 3, 4, 5, 6,
}
var yyTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

const yyFlag = -1000

func yyTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(yyToknames) {
		if yyToknames[c-4] != "" {
			return yyToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yylex1(lex yyLexer, lval *yySymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		c = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			c = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		c = yyTok3[i+0]
		if c == char {
			c = yyTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(c), uint(char))
	}
	return c
}

func yyParse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yychar), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar = yylex1(yylex, &yylval)
	}
	yyn += yychar
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yychar { /* valid shift */
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yychar {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error("syntax error")
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yychar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}
			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		//line epd_parser.y:38
		{
			*yyS[yypt-1].result = yyS[yypt-0].epd
		}
	case 2:
		//line epd_parser.y:40
		{
			*yyS[yypt-1].result = yyS[yypt-0].epd
		}
	case 3:
		//line epd_parser.y:45
		{
			yyVAL.epd = &epdNode{
				position: yyS[yypt-4].position,
			}
		}
	case 4:
		//line epd_parser.y:54
		{
			yyVAL.epd = &epdNode{
				position:   yyS[yypt-1].position,
				operations: yyS[yypt-0].operation,
			}
		}
	case 5:
		//line epd_parser.y:64
		{
			yyVAL.position = &positionNode{
				piecePlacement:  yyS[yypt-6].token,
				sideToMove:      yyS[yypt-4].token,
				castlingAbility: yyS[yypt-2].token,
				enpassantSquare: yyS[yypt-0].token,
			}
		}
	case 6:
		//line epd_parser.y:76
		{
			yyVAL.operation = nil
		}
	case 7:
		//line epd_parser.y:78
		{
			yyVAL.operation = &operationNode{
				operator:  yyS[yypt-2].token,
				arguments: yyS[yypt-1].argument,
			}
			if yyS[yypt-4].operation != nil {
				yyS[yypt-4].operation.next = yyVAL.operation
				yyVAL.operation = yyS[yypt-4].operation
			}
		}
	case 8:
		//line epd_parser.y:92
		{
			yyVAL.argument = nil
		}
	case 9:
		//line epd_parser.y:94
		{
			yyVAL.argument = &argumentNode{
				param: yyS[yypt-0].token,
			}
			if yyS[yypt-2].argument != nil {
				yyS[yypt-2].argument.next = yyVAL.argument
				yyVAL.argument = yyS[yypt-2].argument
			}
		}
	}
	goto yystack /* stack new state and value */
}
