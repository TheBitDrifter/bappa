package table

var _ Schema = &quickSchema{}

type quickSchema struct {
	registered int
	rowIDs     []uint32 // Indexed by ElementTypeID - 1, so that the presence of 0 indicates absence of a elementType.
}

func newQuickSchema() *quickSchema {
	return &quickSchema{}
}

func (s *quickSchema) Register(elementTypes ...ElementType) {
	newIndexes := make([]int, 0)
	minRequiredIndex := 0

	for _, elementType := range elementTypes {
		if s.ContainsAll(elementType) {
			continue
		}
		index := int(elementType.ID() - 1)
		if index > minRequiredIndex {
			minRequiredIndex = index
		}
		newIndexes = append(newIndexes, index)
	}
	s.ensureCapacity(minRequiredIndex + 1)

	for _, i := range newIndexes {
		s.registered++
		s.rowIDs[i] = uint32(s.registered)
	}
}

func (s *quickSchema) Registered() int {
	return s.registered
}

func (s *quickSchema) RowIndexFor(elementType ElementType) uint32 {
	return s.rowIDs[elementType.ID()-1] - 1
}

func (s *quickSchema) RowIndexForID(id ElementTypeID) uint32 {
	return s.rowIDs[id-1] - 1
}

func (s *quickSchema) Contains(elementType ElementType) bool {
	index := int(elementType.ID()) - 1
	if index >= len(s.rowIDs) {
		return false
	}
	return s.rowIDs[index] != 0
}

func (s *quickSchema) ContainsAll(elementTypes ...ElementType) bool {
	for _, elementType := range elementTypes {
		if !s.Contains(elementType) {
			return false
		}
	}
	return true
}

func (s *quickSchema) ensureCapacity(required int) {
	if len(s.rowIDs) > required {
		return
	}
	rowIDs := make([]uint32, required)
	copy(rowIDs, s.rowIDs)
	s.rowIDs = rowIDs
}
