package game

type Square struct {
	row, col int
	piece    *Piece
}

func (s Square) String() string {
	if s.piece == nil {
		return ""
	}
	return s.piece.String()
}

func (s *Square) setPiece(piece *Piece) {
	s.piece = piece
}

func (s Square) hasPiece() bool {
	return s.piece != nil
}

func (s Square) getPiece() *Piece {
	return s.piece
}

func (s Square) stringifySquare() string {
	if s.piece == nil {
		return "_ |"
	} else {
		return "" + s.piece.String() + " |"
	}
}
