package types

type Count int64

func (p Count) Valid() bool {
	return p >= 0
}

func (p Count) Equal(o Count) bool {
	return p == o
}
