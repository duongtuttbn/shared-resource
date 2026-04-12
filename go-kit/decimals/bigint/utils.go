package bigint

import (
	"math/big"
	"strings"

	"golang.org/x/exp/constraints"
)

func FromFloat(x *big.Float) *big.Int {
	v, _ := x.Int(nil)
	return v
}

/** region math */

func Abs(x *big.Int) *big.Int {
	return Zero().Abs(x)
}

func Add(a, b *big.Int) *big.Int {
	return Zero().Add(a, b)
}

func Sub(a, b *big.Int) *big.Int {
	return Zero().Sub(a, b)
}

func Mul(a, b *big.Int) *big.Int {
	return Zero().Mul(a, b)
}

func Div(a, b *big.Int) *big.Int {
	return Zero().Div(a, b)
}

func Quo(a, b *big.Int) *big.Int {
	return Zero().Quo(a, b)
}

func Exp(a, b *big.Int) *big.Int {
	return Zero().Exp(a, b, nil)
}

/* endregion */

/** region Parse */

func Parse(str *string) (*big.Int, bool) {
	if str == nil {
		return nil, false
	}
	return new(big.Int).SetString(standardizeBigIntInput(*str), 10)
}

func SafeParse(str *string) *big.Int {
	value, ok := Parse(str)
	if !ok {
		return big.NewInt(0)
	}
	return value
}

func standardizeBigIntInput(input string) string {
	paths := strings.Split(input, ".")
	return paths[0]
}

/** endregion */

/** region Compare */

func New[T constraints.Integer](num T) *big.Int {
	return big.NewInt(int64(num))
}

func Zero() *big.Int {
	return big.NewInt(0)
}

func IsZero(a *big.Int) bool {
	return IsEqual(a, Zero())
}

func IsEqual(a, b *big.Int) bool {
	return a.Cmp(b) == 0
}

func IsGreaterThan(a, b *big.Int) bool {
	return a.Cmp(b) == 1
}

func IsGreaterThanOrEqual(a, b *big.Int) bool {
	return a.Cmp(b) >= 0
}

func IsLessThan(a, b *big.Int) bool {
	return a.Cmp(b) == -1
}

func IsLessThanOrEqual(a, b *big.Int) bool {
	return a.Cmp(b) <= 0
}

/** endregion */
