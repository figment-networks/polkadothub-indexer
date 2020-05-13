package types

type Height int64

func (h Height) Valid() bool {
	return h >= 0
}

func (h Height) Add(o Height) Height {
	return Height(int64(h) + int64(o))
}

func (h Height) Equal(o Height) bool {
	return h == o
}

func (h Height) Larger(o Height) bool {
	return h > o
}

func (h Height) Int64() int64 {
	return int64(h)
}
