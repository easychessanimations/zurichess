package engine

import (
	"fmt"
	"log"
	"strings"
)

var (
	symbolToCastle = map[rune]Castle{
		'K': WhiteOO,
		'Q': WhiteOOO,
		'k': BlackOO,
		'q': BlackOOO,
	}

	symbolToColor = map[string]Color{
		"w": White,
		"b": Black,
	}
)

type epdNode struct {
	position   *positionNode
	operations *operationNode
}

type positionNode struct {
	piecePlacement  *tokenNode
	sideToMove      *tokenNode
	castlingAbility *tokenNode
	enpassantSquare *tokenNode
}

type operationNode struct {
	operator  *tokenNode
	arguments *argumentNode
	next      *operationNode
}

type argumentNode struct {
	param *tokenNode
	next  *argumentNode
}

type tokenNode struct {
	pos int
	str string
}

func parsePiecePlacement(str string) (*Position, error) {
	pos := &Position{}

	ranks := strings.Split(str, "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("expected 8 ranks, got %d", len(ranks))
	}
	for r := range ranks {
		f := 0
		for _, p := range ranks[r] {
			pi := symbolToPiece[p]
			if pi == NoPiece {
				if '1' <= p && p <= '8' {
					f += int(p) - int('0') - 1
				} else {
					return nil, fmt.Errorf("expected rank or number, got %s", string(p))
				}
			}
			if f >= 8 {
				return nil, fmt.Errorf("rank %d too long (%d cells)", 8-r, f)
			}
			// 7-r because FEN describes the table from 8th rank.
			pos.Put(RankFile(7-r, f), pi)
			f++
		}
		if f < 8 {
			return nil, fmt.Errorf("rank %d too short (%d cells)", r+1, f)
		}
	}

	return pos, nil
}

func parseSideToMove(str string) (Color, error) {
	if col, ok := symbolToColor[str]; ok {
		return col, nil
	}
	return NoColor, fmt.Errorf("invalid color %s", str)
}

func parseCastlingAbility(str string) (Castle, error) {
	if str == "-" {
		return NoCastle, nil
	}
	ability := NoCastle
	for _, p := range str {
		if mask, ok := symbolToCastle[p]; !ok {
			return NoCastle, fmt.Errorf("invalid castling ability %s", str)
		} else {
			ability |= mask
		}
	}
	return ability, nil
}

func parseEnpassantSquare(str string) (Square, error) {
	if str[:1] == "-" {
		return SquareA1, nil
	}
	// TODO: handle error
	return SquareFromString(str), nil
}

func newLeafError(n *tokenNode, err error) error {
	return fmt.Errorf("at %d %s: %v", n.pos, n.str, err)
}

func handleEPDNode(epd *EPD, n *epdNode) error {
	if err := handlePositionNode(epd, n.position); err != nil {
		return err
	}
	if err := handleOperationNode(epd, n.operations); err != nil {
		return err
	}
	return nil
}

func handlePositionNode(epd *EPD, n *positionNode) error {
	var err error
	if epd.Position, err = parsePiecePlacement(n.piecePlacement.str); err != nil {
		return newLeafError(n.piecePlacement, err)
	}
	if sideToMove, err := parseSideToMove(n.sideToMove.str); err != nil {
		return newLeafError(n.sideToMove, err)
	} else {
		epd.Position.SetSideToMove(sideToMove)
	}
	if castlingAbility, err := parseCastlingAbility(n.castlingAbility.str); err != nil {
		return newLeafError(n.castlingAbility, err)
	} else {
		epd.Position.SetCastlingAbility(castlingAbility)
	}
	if enpassantSquare, err := parseEnpassantSquare(n.enpassantSquare.str); err != nil {
		return newLeafError(n.enpassantSquare, err)
	} else {
		epd.Position.SetEnpassantSquare(enpassantSquare)
	}
	return nil
}

func handleId(epd *EPD, n *operationNode) error {
	if n.arguments == nil {
		return newLeafError(n.operator, fmt.Errorf("id missing argument"))
	}
	if n.arguments.next != nil {
		return newLeafError(n.operator, fmt.Errorf("id has to many arguments"))
	}
	epd.Id = n.arguments.param.str
	return nil
}

var handleMap = map[string]func(edp *EPD, n *operationNode) error{
	"id": handleId,
}

func handleOperationNode(epd *EPD, n *operationNode) error {
	for ; n != nil; n = n.next {
		if f, ok := handleMap[n.operator.str]; !ok {
			log.Println("unhandled operation", n.operator.str)
			continue
		} else if err := f(epd, n); err != nil {
			return err
		}
	}
	return nil
}