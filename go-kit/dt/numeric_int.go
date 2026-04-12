package dt

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"tla-backend/pkg/go-kit/decimals/bigfloat"
	"tla-backend/pkg/go-kit/decimals/bigint"

	"golang.org/x/exp/constraints"
)

type NumericInt big.Int

func Num[T constraints.Integer](num T) *NumericInt {
	return NumFromBigInt(big.NewInt(int64(num)))
}

func NumFromBigInt(num *big.Int) *NumericInt {
	return (*NumericInt)(num)
}

func NumFromBigFloat(num *big.Float) *NumericInt {
	return (*NumericInt)(bigint.FromFloat(num))
}

func (b NumericInt) Value() (driver.Value, error) {
	return (*big.Int)(&b).String(), nil
}

func (b *NumericInt) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch t := value.(type) {
	case int64:
		(*big.Int)(b).SetInt64(t)
	case string:
		if !b.setString(t) {
			return fmt.Errorf("failed to load value from string: %v", value)
		}
	case []uint8:
		if !b.setString(string(t)) {
			return fmt.Errorf("failed to load value from []uint8: %v", value)
		}
	default:
		return fmt.Errorf("could not scan type %T into NumericInt", t)
	}

	return nil
}

func (b *NumericInt) ToBigInt() *big.Int {
	return (*big.Int)(b)
}

func (b *NumericInt) ToBigFloat() *big.Float {
	return bigfloat.FromInt(b.ToBigInt())
}

func (b *NumericInt) SetBigInt(value *big.Int) *NumericInt {
	(*big.Int)(b).Set(value)
	return b
}

func (b NumericInt) MarshalJSON() ([]byte, error) {
	return json.Marshal((*big.Int)(&b).String())
}

func (b *NumericInt) UnmarshalJSON(data []byte) error {
	input := strings.Trim(string(data), "\"")
	if !b.setString(input) {
		return fmt.Errorf("failed to load value from data: %v", data)
	}
	return nil
}

func (b *NumericInt) setString(value string) bool {
	value = cleanNumericInput(value)
	_, ok := (*big.Int)(b).SetString(value, 0)
	return ok
}

func cleanNumericInput(input string) string {
	number, decimal, found := strings.Cut(input, ".")
	if !found {
		return input
	}
	if strings.Count(decimal, "0") == len(decimal) {
		return number
	}
	return input
}
