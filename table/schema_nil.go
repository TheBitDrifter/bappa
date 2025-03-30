package table

var _ Schema = &nilSchema{}

type nilSchema struct{}

func newNilSchema() *nilSchema {
	return &nilSchema{}
}

func (s *nilSchema) Register(elementTypes ...ElementType) {}

func (s *nilSchema) Registered() int {
	return int(nextElementTypeID) - 1
}

func (s *nilSchema) RowIndexFor(elementType ElementType) uint32 {
	return uint32(elementType.ID() - 1)
}

func (s *nilSchema) RowIndexForID(id ElementTypeID) uint32 {
	return uint32(id - 1)
}

func (s *nilSchema) Contains(elementType ElementType) bool {
	return int(elementType.ID()) <= Config.MaxElementCount()
}

func (s *nilSchema) ContainsAll(elementTypes ...ElementType) bool {
	for _, elementType := range elementTypes {
		if !s.Contains(elementType) {
			return false
		}
	}
	return true
}
