package game

type Colour int

const (
	undecided Colour = iota
	white     Colour = iota
	black     Colour = iota
)

func getColour(colour Colour) string {
	switch colour {
	case white:
		return "WHITE Player"
	case black:
		return "BLACK Player"
	default:
		return "Unknown Player"
	}
}

func getOpponentColourName(colour Colour) string {
	var opponent Colour
	if colour == white {
		opponent = black
	} else {
		opponent = white
	}
	return getColour(opponent)
}

func getOpponentColour(curColour Colour) Colour {
	if curColour == black {
		return white
	} else {
		return black
	}
}
