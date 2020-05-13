package types

type SequenceId int64

func (id SequenceId) Valid() bool {
	return id >= 1
}
