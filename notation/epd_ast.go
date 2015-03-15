// epd_ast.go interprests the ast tree parsed from an EPD line.
// *Node structures correspond to grammar nodes in epd_parser.y.

package notation

import (
	"fmt"
	"strconv"

	"bitbucket.org/brtzsnr/zurichess/engine"
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

func newLeafError(n *tokenNode, err error) error {
	return fmt.Errorf("at %d %s: %v", n.pos, n.str, err)
}

func trimQuotes(str string) string {
	l := len(str)
	switch {
	case l < 2:
		return str
	case str[0] == '"' && str[l-1] == '"':
		return str[1 : l-1]
	default:
		return str
	}
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
	epd.Position = engine.NewPosition()

	var err error
	if err = engine.ParsePiecePlacement(n.piecePlacement.str, epd.Position); err != nil {
		return newLeafError(n.piecePlacement, err)
	}
	if err := engine.ParseSideToMove(n.sideToMove.str, epd.Position); err != nil {
		return newLeafError(n.sideToMove, err)
	}
	if err := engine.ParseCastlingAbility(n.castlingAbility.str, epd.Position); err != nil {
		return newLeafError(n.castlingAbility, err)
	}
	if err := engine.ParseEnpassantSquare(n.enpassantSquare.str, epd.Position); err != nil {
		return newLeafError(n.enpassantSquare, err)
	}
	return nil
}

// handleBestMove handles "id" operator.
func handleId(epd *EPD, n *operationNode) error {
	ptr := n.arguments
	if ptr == nil || ptr.next != nil {
		return newLeafError(ptr.param, fmt.Errorf("id expects exactly one argument"))
	}
	epd.Id = trimQuotes(ptr.param.str)
	return nil
}

// handleBestMove handles "bm" operator.
func handleBestMove(epd *EPD, n *operationNode) error {
	for ptr := n.arguments; ptr != nil; ptr = ptr.next {
		if move, err := epd.Position.SANToMove(ptr.param.str); err != nil {
			return newLeafError(ptr.param, fmt.Errorf("invalid move: %v", err))
		} else {
			epd.BestMove = append(epd.BestMove, move)
		}
	}
	return nil
}

func handleFullMoveNumber(epd *EPD, n *operationNode) error {
	ptr := n.arguments
	if ptr == nil || ptr.next != nil {
		return newLeafError(ptr.param, fmt.Errorf("fmvn expects exactly one argument"))
	}

	var err error
	epd.Position.FullMoveNumber, err = strconv.Atoi(ptr.param.str)
	return err
}

func handleHalfMoveClock(epd *EPD, n *operationNode) error {
	ptr := n.arguments
	if ptr == nil || ptr.next != nil {
		return newLeafError(ptr.param, fmt.Errorf("hmvc expects exactly one argument"))
	}

	var err error
	epd.Position.HalfMoveClock, err = strconv.Atoi(ptr.param.str)
	return err
}

func handleComment(epd *EPD, n *operationNode) error {
	ptr := n.arguments
	if ptr == nil || ptr.next != nil {
		return newLeafError(ptr.param, fmt.Errorf("c# expects exactly one argument"))
	}

	epd.Comment[n.operator.str] = trimQuotes(ptr.param.str)
	return nil
}

// handleMap is a map from operator to a function handling the node.
var handleMap = map[string]func(edp *EPD, n *operationNode) error{
	"id":   handleId,
	"bm":   handleBestMove,
	"fmvn": handleFullMoveNumber,
	"hmvc": handleHalfMoveClock,
	"c0":   handleComment,
	"c1":   handleComment,
	"c2":   handleComment,
	"c3":   handleComment,
	"c4":   handleComment,
	"c5":   handleComment,
	"c6":   handleComment,
	"c7":   handleComment,
	"c8":   handleComment,
	"c9":   handleComment,
}

func handleOperationNode(epd *EPD, n *operationNode) error {
	for ; n != nil; n = n.next {
		if f, ok := handleMap[n.operator.str]; !ok {
			continue
		} else if err := f(epd, n); err != nil {
			return err
		}
	}
	return nil
}
