package perbill

import (
	"math/big"
)

const (
	maxval = 1000000000
)

var (
	zero big.Int
	one  big.Int = *big.NewInt(1)

	max     big.Int = *big.NewInt(maxval)
	halfMax big.Int = *big.NewInt(maxval / 2)
)

// Perbill is a (incomplete) golang implementation of rusts sp_arithmetic::Perbill
type Perbill struct {
	big.Int
}

func FromParts(parts int64) Perbill {
	parts = minimum(parts, maxval)
	b := big.NewInt(parts)
	return Perbill{Int: *b}
}

func FromRationalApproximation(p, q big.Int) Perbill {
	qReduce := big.Int{}
	// q cannot be zero.
	if q.Cmp(&one) < 0 {
		qReduce.Set(&one)
	} else {
		qReduce.Set(&q)
	}

	pReduce := big.Int{}
	// p should not be bigger than q.
	if qReduce.Cmp(&p) < 1 {
		pReduce.Set(&qReduce)
	} else {
		pReduce.Set(&p)
	}

	factor := big.Int{}
	factor.Quo(&qReduce, &max)

	rem := big.Int{}
	rem.Mod(&qReduce, &max)
	if rem.CmpAbs(&zero) != 0 {
		factor.Add(&factor, &one)
	}

	if factor.Cmp(&one) < 1 {
		factor.Set(&one)
	}

	qReduce.Quo(&qReduce, &factor)
	pReduce.Quo(&pReduce, &factor)

	part := big.Int{}
	part.Mul(&pReduce, &max)
	part.Quo(&part, &qReduce)

	return Perbill{part}
}

// see https://github.com/w3f/substrate/blob/ed258da33752aa49e76ab077f750c48ad0e43fab/core/sr-primitives/src/sr_arithmetic.rs#L162
func (p *Perbill) Mul(b big.Int) big.Int {
	part := p.Int

	rem := big.Int{}
	rem.Mod(&b, &max)

	remMultipliedUpper := big.Int{}
	remMultipliedUpper.Mul(&rem, &part)

	remMultipliedDividedSized := big.Int{}
	remMultipliedDividedSized.Quo(&remMultipliedUpper, &max)

	rem.Mod(&remMultipliedUpper, &max)
	if rem.Cmp(&halfMax) == 1 {
		remMultipliedDividedSized.Add(&remMultipliedDividedSized, &one)
	}

	result := big.Int{}
	result.Quo(&b, &max)
	result.Mul(&result, &part)
	result.Add(&result, &remMultipliedDividedSized)

	return result
}

func minimum(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
