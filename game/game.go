package game

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

const (
	MovesLimitCount      = 400
	InitialBoardFileName = "./playbook/initialBoard.txt"
)

func New() ChessGame {
	game := ChessGame{NewBoard(), 0, undecided, *bufio.NewReader(os.Stdin)}
	return game
}

type ChessGame struct {
	board       *Board
	movesCount  int
	curColour   Colour
	inputReader bufio.Reader
}

func (game *ChessGame) StartInteractiveMode() {
	game.setupBoard(InitialBoardFileName)
	game.printGameStatus()

	for {
		game.changeTurn(true)
		game.printAvailableMovesInCheck()
		input := game.promptInput(game.inputReader)
		gameEnd := game.execute(input)
		if gameEnd {
			return
		}
	}
}

func (game *ChessGame) StartFileMode(path string) {
	fmt.Println("Entered file path: ", path)
}

func (game *ChessGame) setupBoard(path string) TestCase {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error: Unable to setup board from file: ", path)
		}
	}()

	testCase := ParseTestCase(path)
	game.board.setup(testCase)
	return testCase
}

func (game *ChessGame) execute(command string) bool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			game.changeTurn(false)
		}
	}()

	checkmate := game.board.execute(command, game.curColour)

	if checkmate {
		game.endGameWithWinner(getColour(game.curColour), "Checkmate", command)
	}

	if game.isDraw() {
		game.endGameByDraw(command)
		return true
	}

	game.printAction(command)
	game.printGameStatus()
	return false
}

func (game ChessGame) promptInput(reader bufio.Reader) string {
	fmt.Print(getColour(game.curColour), "> ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimRight(input, "\n") //remove "\n" from input

	return input
}

func (game *ChessGame) changeTurn(goesNext bool) {

	if goesNext {
		game.movesCount++
	} else {
		game.movesCount--
	}

	//Switch team
	switch game.curColour {
	case undecided:
		game.curColour = white
	case black:
		game.curColour = white
	case white:
		game.curColour = black
	}
}

func (game ChessGame) printAvailableMovesInCheck() {
	curTeam := game.curColour

	if !game.board.inCheck(curTeam) {
		return
	}

	fmt.Println(getColour(game.curColour) + " is in check!")
	fmt.Println("Available moves:")
	availableMoves := game.board.getAvailableMovesInCheck(curTeam)
	for _, move := range availableMoves {
		fmt.Println(move)
	}
	fmt.Println()
}

func (game ChessGame) isDraw() bool {
	return game.movesCount >= MovesLimitCount
}

func (game ChessGame) endGameWithWinner(winnerPlayer string, reason interface{}, lastCommand string) {
	game.printAction(lastCommand)
	game.printGameStatus()
	fmt.Println()
	fmt.Println(winnerPlayer, "player wins. ", reason)
}

func (game ChessGame) endGameByDraw(lastCommand string) {
	game.printAction(lastCommand)
	game.printGameStatus()
	fmt.Println("Tie game.  Too many moves.")
}

func (game ChessGame) printGameStatus() {
	fmt.Println(game.board.String())
}

func (game ChessGame) printAction(action string) {
	fmt.Println(getColour(game.curColour), " player action: ", action)
}

func ParseTestCase(path string) TestCase {
	file, err := os.Open(path)

	defer file.Close()

	if err != nil {
		panic(err)
	}

	var line string
	reader := bufio.NewReader(file)
	line, err = reader.ReadString('\n')
	line = strings.TrimSpace(line)

	var initialPositions []InitialPosition

	for line != "" {
		lineParts := strings.Split(line, " ")
		initialPositions = append(initialPositions, InitialPosition{lineParts[0], lineParts[1]})
		line, _ = reader.ReadString('\n')
		line = strings.TrimSpace(line)
	}

	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	whiteCaptures := strings.Split(line[1:len(line)-1], " ")

	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)
	blackCaptures := strings.Split(line[1:len(line)-1], " ")

	line, _ = reader.ReadString('\n')
	line = strings.TrimSpace(line)

	var moves []string
	for line != "" {
		line = strings.TrimSpace(line)
		moves = append(moves, line)
		line, _ = reader.ReadString('\n')
	}

	return TestCase{initialPositions, whiteCaptures, blackCaptures, moves}
}

type InitialPosition struct {
	Sign     string
	Position string
}

func (ip InitialPosition) String() string {
	return ip.Sign + " " + ip.Position
}

type TestCase struct {
	InitialPositions             []InitialPosition
	WhiteCaptures, BlackCaptures []string
	Moves                        []string
}

func (tc TestCase) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("initialPieces: [\n")
	for _, piece := range tc.InitialPositions {
		buffer.WriteString(piece.String())
		buffer.WriteString("\n")
	}
	buffer.WriteString("]\n")

	buffer.WriteString("moves: [\n")
	for _, move := range tc.Moves {
		buffer.WriteString(move)
		buffer.WriteString("\n")
	}
	buffer.WriteString("]")

	return buffer.String()
}
