package types

const (
	SyncableTypeBlock = "syncable_block"
)

var Types = []SyncableType{SyncableTypeBlock}

type SyncableType string

func (t SyncableType) Valid() bool {
	return t == SyncableTypeBlock
}
