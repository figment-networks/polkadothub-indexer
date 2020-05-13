package types

type Nonce uint64

func (h Nonce) Valid() bool {
	return h >= 0
}

func (h Nonce) Equal(o Nonce) bool {
	return h == o
}
