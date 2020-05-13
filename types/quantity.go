package types

import (
	"database/sql/driver"
	"fmt"
	"math/big"
)

type Quantity struct {
	big.Int
}

func NewQuantity(i *big.Int) Quantity {
	return Quantity{Int: *i}
}
func NewQuantityFromInt64(i int64) Quantity {
	b := big.NewInt(i)
	return Quantity{Int: *b}
}

func NewQuantityFromBytes(bytes []byte) Quantity {
	b := big.Int{}
	return Quantity{Int: *b.SetBytes(bytes)}
}

func (b *Quantity) Valid() bool {
	return b.Int.Sign() >= 0
}

func (b *Quantity) Equals(o Quantity) bool {
	return b.Int.String() == o.Int.String()
}

// Value implement sql.Scanner
func (b *Quantity) Value() (driver.Value, error) {
	if b != nil {
		return (b).String(), nil
	}
	return nil, nil
}

func (b *Quantity) Scan(value interface{}) error {
	b.Int = *new(big.Int)
	if value == nil {
		return nil
	}
	switch t := value.(type) {
	case int64:
		b.SetInt64(t)
	case []byte:
		b.SetString(string(value.([]byte)), 10)
	case string:
		b.SetString(t, 10)
	default:
		return fmt.Errorf("could not scan type %T into BigInt ", t)
	}
	return nil
}