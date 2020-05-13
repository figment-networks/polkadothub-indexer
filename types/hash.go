package types

type Hash string

func (h Hash) Valid() bool {
	return h != ""
}

func (h Hash) Equal(o Hash) bool {
	return h == o
}
