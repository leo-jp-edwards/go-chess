package game

type Move struct {
	piece                *Piece
	squareFrom, squareTo *Square
}

func containsMove(moves []string, move string) bool {
	for _, element := range moves {
		if move == element {
			return true
		}
	}
	return false
}

func getCoordinatePosition(row, col int) string {
	return (string)('a'+col) + (string)('0'-row+boardSize)
}

func getSquarePosition(square Square) string {
	return getCoordinatePosition(square.row, square.col)
}
