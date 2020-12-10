package sp_arithmetic_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/figment-networks/polkadothub-indexer/types/sp_arithmetic"
)

func PerbillFromParts(t *testing.T) {
	tests := []struct {
		input  int64
		expect string
	}{
		{0, "0"},
		{10, "10"},
		{20000000, "20000000"},
		{200000000, "200000000"},
		{987654321, "987654321"},
		{999999999, "999999999"},

		{1000000000, "1000000000"},
		{1000000001, "1000000000"},
		{1111111111, "1000000000"},
		{2000000000, "1000000000"},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			res := sp_arithmetic.PerbillFromParts(tt.input)

			if res.String() != tt.expect {
				t.Errorf("unexpected result, want %v; got %v", tt.expect, res.String())
				return
			}
		})
	}
}

func Test_FromRationalApproximation(t *testing.T) {
	tests := []struct {
		p      string
		q      string
		expect string
	}{
		{"0", "0", "0"},
		{"0", "3", "0"},
		{"3", "0", "1000000000"},

		{"1", "3", "333333333"},
		{"2", "3", "666666666"},
		{"3", "3", "1000000000"},
		{"4", "3", "1000000000"},

		{"1", "4", "250000000"},
		{"2", "4", "500000000"},
		{"3", "4", "750000000"},
		{"5", "4", "1000000000"},
		{"8", "4", "1000000000"},

		{"1", "300", "3333333"},
		{"1", "30000", "33333"},
		{"1", "300000000", "3"},
		{"1", "3000000000", "0"},

		{"100000000000", "300000000000", "333333333"},
		{"100000000000", "200000000000", "500000000"},
		{"200000000000", "100000000000", "1000000000"},
		{"11237891723000000", "38129311000000000", "294731046"},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			p := new(big.Int)
			p, _ = p.SetString(tt.p, 10)

			q := new(big.Int)
			q, _ = q.SetString(tt.q, 10)

			res := sp_arithmetic.PerbillFromRationalApproximation(*p, *q)

			if res.String() != tt.expect {
				t.Errorf("unexpected result, want %v; got %v", tt.expect, res.String())
				return
			}

			if p.String() != tt.p || q.String() != tt.q {
				t.Errorf("params should not mutate, want p=%v, q=%v; got p=%v q=%v", tt.p, tt.q, p, q)
				return
			}
		})
	}
}

func Test_Mul(t *testing.T) {
	tests := []struct {
		p      int64
		b      string
		expect string
	}{
		{1000000000, "0", "0"},
		{1000000000, "1", "1"},
		{1000000000, "987654321987654321", "987654321987654321"},

		{500000000, "1", "0"},
		{500000000, "2", "1"},
		{500000000, "5", "2"},
		{500000000, "500", "250"},

		{333333333, "5", "2"},
		{333333333, "50", "17"},
		{333333333, "500", "167"},
		{333333333, "5000000", "1666667"},
		{333333333, "500000000000000", "166666666500000"},

		{333333333, "7", "2"},
		{333333333, "70", "23"},
		{333333333, "700000000000", "233333333100"},

		{333333333, "2", "1"},
		{333333333, "20", "7"},
		{333333333, "200", "67"},
		{333333333, "200000000000", "66666666600"},

		{333333333, "1", "0"},
		{333333333, "10", "3"},
		{333333333, "100", "33"},
		{333333333, "100000000000", "33333333300"},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			p := sp_arithmetic.PerbillFromParts(tt.p)

			b := new(big.Int)
			b, _ = b.SetString(tt.b, 10)

			res := p.Mul(*b)
			if res.String() != tt.expect {
				t.Errorf("unexpected result, want %v; got %v", tt.expect, res.String())
				return
			}

			if p.String() != strconv.FormatInt(tt.p, 10) || b.String() != tt.b {
				t.Errorf("params should not mutate, want p=%v, b=%v; got p=%v b=%v", tt.p, tt.b, p, b)
				return
			}
		})
	}
}
