//line epd_parser.y:2

// epd_parser.y implements parsing of chess positions in EPD and FEN notations.
// For EPD see https://chessprogramming.wikispaces.com/Extended+Position+Description
// For FEN see https://chessprogramming.wikispaces.com/Forsyth-Edwards+Notation
package engine

import __yyfmt__ "fmt"

//line epd_parser.y:5
import (
	"fmt"
	"log"
	"strings"
	"unicode"
)

//line epd_parser.y:16
type yySymType struct {
	yys      int
	str      string
	result   **EPD
	epd      *EPD
	position *Position
	castling Castle
	square   Square
	toMove   Color
}

const OPERAND = 57346
const HIDDEN_FEN = 57347
const HIDDEN_EPD = 57348

var yyToknames = []string{
	"OPERAND",
	"HIDDEN_FEN",
	"HIDDEN_EPD",
}
var yyStatenames = []string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line epd_parser.y:93
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

const yyLast = 23

var yyAct = []int{

	12, 21, 17, 16, 11, 10, 2, 3, 23, 20,
	18, 15, 13, 7, 5, 1, 14, 22, 9, 19,
	6, 4, 8,
}
var yyPact = []int{

	1, -1000, 9, 9, -1000, -2, -3, -1000, -1000, -8,
	8, 7, -1000, -4, -5, -1000, 6, 5, -1000, -6,
	-1000, 4, -1000, -1000,
}
var yyPgo = []int{

	0, 22, 21, 14, 20, 19, 17, 16, 15,
}
var yyR1 = []int{

	0, 8, 8, 2, 1, 3, 4, 7, 5, 6,
}
var yyR2 = []int{

	0, 2, 2, 5, 2, 7, 1, 1, 1, 1,
}
var yyChk = []int{

	-1000, -8, 5, 6, -2, -3, -4, 4, -1, -3,
	7, 7, 8, 4, -7, 4, 7, 7, 4, -5,
	4, 7, -6, 4,
}
var yyDef = []int{

	0, -2, 0, 0, 1, 0, 0, 6, 2, 0,
	0, 0, 4, 0, 0, 7, 0, 0, 3, 0,
	8, 0, 5, 9,
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
		//line epd_parser.y:40
		{
			*yyS[yypt-1].result = yyS[yypt-0].epd
		}
	case 2:
		//line epd_parser.y:42
		{
			*yyS[yypt-1].result = yyS[yypt-0].epd
		}
	case 3:
		//line epd_parser.y:47
		{
			yyVAL.epd = &EPD{
				Position: yyS[yypt-4].position,
			}
		}
	case 4:
		//line epd_parser.y:56
		{
			yyVAL.epd = &EPD{
				Position: yyS[yypt-1].position,
			}
		}
	case 5:
		//line epd_parser.y:65
		{
			yyVAL.position = yyS[yypt-6].position
			yyVAL.position.ToMove = yyS[yypt-4].toMove
			yyVAL.position.Castle = yyS[yypt-2].castling
			yyVAL.position.Enpassant = yyS[yypt-0].square
		}
	case 6:
		//line epd_parser.y:75
		{
			yyVAL.position = parsePiecePlacement(yyS[yypt-0].str)
		}
	case 7:
		//line epd_parser.y:80
		{
			yyVAL.toMove = parseSideToMove(yyS[yypt-0].str)
		}
	case 8:
		//line epd_parser.y:85
		{
			yyVAL.castling = parseCastlingAbility(yyS[yypt-0].str)
		}
	case 9:
		//line epd_parser.y:90
		{
			yyVAL.square = parseEnpassantSquare(yyS[yypt-0].str)
		}
	}
	goto yystack /* stack new state and value */
}
