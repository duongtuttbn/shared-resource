package bigfloat

import (
	"golang.org/x/exp/constraints"
	"math"
	"math/big"
	"github.com/duongtuttbn/shared-resource/go-kit/decimals/bigint"
)

func FromInt(x *big.Int) *big.Float {
	return new(big.Float).SetInt(x)
}

/** region math */

func Abs(x *big.Float) *big.Float {
	return Zero().Abs(x)
}

func Add(a, b *big.Float) *big.Float {
	return Zero().Add(a, b)
}

func Sub(a, b *big.Float) *big.Float {
	return Zero().Sub(a, b)
}

func Mul(a, b *big.Float) *big.Float {
	return Zero().Mul(a, b)
}

func Quo(a, b *big.Float) *big.Float {
	return Zero().Quo(a, b)
}

/* endregion */

/** region Parse */

func Parse(str *string) (*big.Float, bool) {
	if str == nil {
		return nil, false
	}
	return new(big.Float).SetString(*str)
}

func SafeParse(str *string) *big.Float {
	f, ok := Parse(str)
	if !ok {
		return big.NewFloat(0)
	}
	return f
}

/** endregion */

func ConvertDecimals(number *big.Float, targetDecimals, currentDecimals int) *big.Float {
	if currentDecimals == targetDecimals {
		return number
	}
	diffDecimals := targetDecimals - currentDecimals

	decimalDelta, _ := bigint.Exp(bigint.New(10), bigint.New(int(math.Abs(float64(diffDecimals))))).Float64()

	if diffDecimals >= 0 {
		return Mul(number, New(decimalDelta))
	}

	return Quo(number, New(decimalDelta))
}

/** region Compare */

func New[T constraints.Float](num T) *big.Float {
	return big.NewFloat(float64(num))
}

func Zero() *big.Float {
	return big.NewFloat(0)
}

func IsZero(a *big.Float) bool {
	return IsEqual(a, big.NewFloat(0))
}

func IsEqual(a, b *big.Float) bool {
	return a.Cmp(b) == 0
}

func IsGreaterThan(a, b *big.Float) bool {
	return a.Cmp(b) == 1
}

func IsGreaterThanOrEqual(a, b *big.Float) bool {
	return a.Cmp(b) >= 0
}

func IsLessThan(a, b *big.Float) bool {
	return a.Cmp(b) == -1
}

func IsLessThanOrEqual(a, b *big.Float) bool {
	return a.Cmp(b) <= 0
}

/** endregion */

/** region Math */

func Div(a, b *big.Float) *float64 {
	if IsZero(a) || IsZero(b) {
		return nil
	}

	result, _ := new(big.Float).Quo(a, b).Float64()
	return &result
}

func DivString(a, b *string) *float64 {
	return Div(SafeParse(a), SafeParse(b))
}

/** endregion */

/** region Format */

func ToIntString(a *big.Float) string {
	bigInt, _ := a.Int(nil)
	return bigInt.String()
}

/** endregion */
