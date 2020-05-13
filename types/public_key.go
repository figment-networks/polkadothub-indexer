package types

type PublicKey string

func (p PublicKey) Valid() bool {
	return p != ""
}

func (p PublicKey) Equal(o PublicKey) bool {
	return p == o
}
