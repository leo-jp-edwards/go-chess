package game

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

const boardSize = 8

type Board struct {
	squares                      [][]Square
	whiteCaptures, blackCaptures []string
}

func NewBoard() *Board {
	board := new(Board)
	squares := make([][]Square, boardSize)
	for i := 0; i < boardSize; i++ {
		squares[i] = make([]Square, boardSize)
		for j := 0; j < boardSize; j++ {
			square := Square{i, j, nil}
			squares[i][j] = square
		}
	}
	board.squares = squares
	return board
}

func (b Board) initPiece(position string, sign string) {
	square := b.getSquare(position)
	if square == nil || square.hasPiece() {
		panic("initPiece() failed on positon: " + position)
		return
	}
	piece := createPiece(sign, square.row, square.col)
	square.setPiece(&piece)
}

func (b *Board) setup(testCase TestCase) {
	for _, ip := range testCase.InitialPositions {
		b.initPiece(ip.Position, ip.Sign)
	}

	b.whiteCaptures = testCase.WhiteCaptures
	b.blackCaptures = testCase.BlackCaptures
}

func (b *Board) execute(command string, colour Colour) bool {
	if b.inCheck(colour) {
		validMoves := b.getAvailableMovesInCheck(colour)
		if !containsMove(validMoves, command) {
			panic(illegalMove)
		}
	}

	tokens := strings.Split(command, " ")
	if len(tokens) != 2 {
		panic(illegalMove)
	}

	move, err := b.checkMove(tokens[0], tokens[1], colour)
	if err != nil {
		return false
	}
	b.movePiece(move.piece, move.squareFrom, move.squareTo)
	// Handle promotion
	return b.inCheckmate(getOpponentColour(colour))
}

func (b *Board) movePiece(piece *Piece, squareFrom, squareTo *Square) {
	capturedPiece := squareTo.piece
	if capturedPiece != nil {
		b.captured(*capturedPiece)
	}

	squareFrom.setPiece(nil)
	squareTo.setPiece(piece)

	piece.row = squareTo.row
	piece.col = squareTo.col
}

func (b *Board) captured(capturedPiece Piece) {
	colour := getOpponentColour(capturedPiece.colour)
	sign := capturedPiece.sign
	switch colour {
	case white:
		b.whiteCaptures = append(b.whiteCaptures, sign)
	case black:
		b.blackCaptures = append(b.blackCaptures, sign)
	}
}

func (b Board) inCheck(curColour Colour) bool {
	kingPosition := b.getKingPosition(curColour)
	opponentPositions := b.getReachablePositions(getOpponentColour(curColour))
	for _, position := range opponentPositions {
		if kingPosition == position {
			return true
		}
	}
	return false
}

func (b Board) inCheckmate(curColour Colour) bool {
	return b.inCheck(curColour) && len(b.getAvailableMovesInCheck(curColour)) == 0
}

// this should return an array of Moves
func (b Board) getAvailableMovesInCheck(colour Colour) []string {
	var availableMoves []string

	availableMoves = append(availableMoves, b.getAvailableMovesByMovingKing(colour)...)

	availableMoves = append(availableMoves, b.getAvailableMovesByMovingOtherPieces(colour)...)

	return availableMoves
}

func (b Board) getAvailableMovesByMovingKing(colour Colour) []string {
	var moves []string
	kingPosition := b.getKingPosition(colour)
	kingPiece := b.getSquare(kingPosition).piece

	opponentMoves := b.getReachablePositions(getOpponentColour(colour))

	for _, kingMove := range getMoves(b, *kingPiece) {
		if !containsMove(opponentMoves, kingMove) {
			moves = append(moves, kingPosition+" "+kingMove)
		}
	}
	return moves
}

func (b Board) getAvailableMovesByMovingOtherPieces(colour Colour) []string {
	var moves []string
	threatenKingPieces := b.getThreateningKingPieces(colour)
	if len(threatenKingPieces) != 1 {
		return moves //Too many threatening pieces to capture
	}

	threatenPiece := threatenKingPieces[0]
	threatenPosition := getCoordinatePosition(threatenPiece.row, threatenPiece.col)

	for _, ownPiece := range b.getAllPieces(colour) {
		if isKing(ownPiece) {
			continue //king's move was already considered in method getAvailableMovesByMovingKing()
		}
		positionFrom := getCoordinatePosition(ownPiece.row, ownPiece.col)

		for _, positionTo := range getMoves(b, ownPiece) {
			if !b.moveWillCauseSelfCheck(positionFrom, positionTo, colour) {
				moves = append(moves, positionFrom+" "+threatenPosition)
			}
		}
	}
	return moves
}

func (b Board) getThreateningKingPieces(colour Colour) []Piece {
	var threatenPieces []Piece
	kingPosition := b.getKingPosition(colour)
	for _, opponentPiece := range b.getAllPieces(getOpponentColour(colour)) {
		opponentMoves := getMoves(b, opponentPiece)
		if containsMove(opponentMoves, kingPosition) {
			threatenPieces = append(threatenPieces, opponentPiece)
		}
	}
	return threatenPieces
}

func (b Board) moveWillCauseSelfCheck(positionFrom, positionTo string, colour Colour) bool {

	squareFrom := b.getSquare(positionFrom)
	squareTo := b.getSquare(positionTo)

	pieceFrom := squareFrom.piece
	pieceTo := squareTo.piece

	//Move piece and check
	b.movePiece(pieceFrom, squareFrom, squareTo)
	selfInCheck := b.inCheck(colour)

	//Move pieces back
	b.movePiece(pieceFrom, squareTo, squareFrom)
	squareTo.setPiece(pieceTo)

	return selfInCheck
}

func (b Board) canMoveTo(position string, colour Colour) bool {

	square := b.getSquare(position)
	if square == nil {
		return false
	}
	piece := square.getPiece()
	return piece == nil || piece.colour != colour
}

func (b Board) isEmptyAt(position string) bool {
	square := b.getSquare(position)
	return square != nil && !square.hasPiece()
}

func (b Board) getKingPosition(colour Colour) string {
	var kingSymbol string
	if colour == white {
		kingSymbol = WhiteKing
	} else {
		kingSymbol = BlackKing
	}

	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			square := b.squares[i][j]
			piece := square.getPiece()
			if piece != nil && piece.String() == kingSymbol {
				return getSquarePosition(square)
			}
		}
	}

	panic("Error: Cannot find King from the board")
}

func (b Board) getReachablePositions(colour Colour) []string {
	var moves []string
	pieces := b.getAllPieces(colour)
	for _, piece := range pieces {
		moves = append(moves, getMoves(b, piece)...)
	}
	return moves
}

func (b Board) getAllPieces(colour Colour) []Piece {
	var pieces []Piece
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			piece := b.squares[i][j].getPiece()
			if piece != nil && piece.colour == colour {
				pieces = append(pieces, *piece)
			}

		}
	}
	return pieces
}

func (b Board) getSquare(position string) *Square {
	col := int(position[0] - 'a')
	if col < 0 || col >= boardSize {
		return nil
	}
	row := ((int)(position[1]-'0'))*-1 + boardSize
	if row < 0 || row >= boardSize {
		return nil
	}
	return &b.squares[row][col]
}

func (b Board) String() string {
	var buffer bytes.Buffer

	boardStr := make([][]string, boardSize)
	for i := 0; i < boardSize; i++ {
		boardStr[i] = make([]string, boardSize)
		for j := 0; j < boardSize; j++ {
			boardStr[i][j] = b.squares[i][j].String()
		}
	}
	buffer.WriteString(b.StringifyBoard())

	buffer.WriteString("White Captures: [")
	for _, capturedSign := range b.whiteCaptures {
		if capturedSign != "" {
			buffer.WriteString(getPieceSymbol(capturedSign) + " ") //cannot do getPieceSymbol("")
		}
	}
	buffer.WriteString("] \n")

	buffer.WriteString("Black Captures: [")
	for _, capturedSign := range b.blackCaptures {
		if capturedSign != "" {
			buffer.WriteString(getPieceSymbol(capturedSign) + " ") //cannot do getPieceSymbol("")
		}
	}
	buffer.WriteString("] \n")

	return buffer.String()
}

func (b Board) StringifyBoard() string {
	row := len(b.squares)
	col := len(b.squares[0])

	var buffer bytes.Buffer

	for i := 0; i < col; i++ {
		if i == 0 {
			buffer.WriteString("   ")
		}
		colLetter := (string)('a' + i)
		buffer.WriteString("" + colLetter + " ")
	}
	buffer.WriteString("\n")

	for i := row - 1; i >= 0; i-- {
		buffer.WriteString(strconv.Itoa(i + 1))
		buffer.WriteString(" |")
		for j := 0; j < col; j++ {
			buffer.WriteString(b.squares[-i+len(b.squares)-1][j].stringifySquare())
		}
		buffer.WriteString(" " + strconv.Itoa(i+1))
		buffer.WriteString("\n")
		if i != 0 {
			buffer.WriteString("\n")
		}
	}

	for i := 0; i < col; i++ {
		if i == 0 {
			buffer.WriteString("   ")
		}
		colLetter := (string)('a' + i)
		buffer.WriteString("" + colLetter + " ")
	}
	buffer.WriteString("\n")
	return buffer.String()
}

func (b Board) checkMove(origin, destination string, colour Colour) (*Move, error) {
	squareFrom := b.getSquare(origin)
	if squareFrom == nil {
		return nil, errors.New(illegalMove)
	}
	squareTo := b.getSquare(destination)
	if squareTo == nil {
		return nil, errors.New(illegalMove)
	}
	piece := squareFrom.getPiece()
	if piece == nil || piece.colour != colour {
		return nil, errors.New(illegalMove)
	}
	moves := getMoves(b, *piece)
	if !containsMove(moves, destination) {
		return nil, errors.New(illegalMove)
	}

	if b.moveWillCauseSelfCheck(origin, destination, colour) {
		return nil, errors.New(causingSelfCheck)
	}
	return &Move{piece, squareFrom, squareTo}, nil
}
