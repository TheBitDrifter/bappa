package table

var Stats stats = stats{}

type stats struct{}

func (s stats) TotalElementTypes() int {
	return int(nextElementTypeID) - 1
}
