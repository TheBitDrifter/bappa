package table

type DefaultTableEvents struct{}

func (e *DefaultTableEvents) OnBeforeEntriesCreated(count int) error     { return nil }
func (e *DefaultTableEvents) OnAfterEntriesCreated(entries []Entry)      {}
func (e *DefaultTableEvents) OnBeforeEntriesDeleted(indices []int) error { return nil }
func (e *DefaultTableEvents) OnAfterEntriesDeleted(ids []EntryID)        {}
