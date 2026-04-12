package kit

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	result := make([]R, len(collection))

	for i := range collection {
		result[i] = iteratee(collection[i])
	}

	return result
}

func ForEach[T any](collection []T, iteratee func(item T)) {
	for i := range collection {
		iteratee(collection[i])
	}
}

func Filter[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	result := make(Slice, 0, len(collection))

	for i := range collection {
		if predicate(collection[i]) {
			result = append(result, collection[i])
		}
	}

	return result
}

func Reject[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	result := Slice{}

	for i := range collection {
		if !predicate(collection[i]) {
			result = append(result, collection[i])
		}
	}

	return result
}

func FilterMap[T any, R any](collection []T, callback func(item T) (R, bool)) []R {
	result := make([]R, 0, len(collection))

	for i := range collection {
		if r, ok := callback(collection[i]); ok {
			result = append(result, r)
		}
	}

	return result
}

func RejectMap[T any, R any](collection []T, callback func(item T) (R, bool)) []R {
	var result []R

	for i := range collection {
		if r, ok := callback(collection[i]); !ok {
			result = append(result, r)
		}
	}

	return result
}

func UniqMap[T any, R comparable](collection []T, iteratee func(item T) R) []R {
	result := make([]R, 0, len(collection))
	seen := make(map[R]struct{}, len(collection))

	for i := range collection {
		r := iteratee(collection[i])
		if _, ok := seen[r]; !ok {
			result = append(result, r)
			seen[r] = struct{}{}
		}
	}
	return result
}

// FloatToString converts a float64 to a string with up to 6 decimal places,
// trimming unnecessary trailing zeros and the decimal point if it's left alone.
//
// Example:
//
//	FloatToString(3.140000) // "3.14"
//	FloatToString(2.000000) // "2"
//	FloatToString(5.123456) // "5.123456"
func FloatToString(value float64) string {
	raw := fmt.Sprintf("%.6f", value)
	for strings.HasSuffix(raw, "0") {
		raw = raw[:len(raw)-1]
	}
	if strings.HasSuffix(raw, ".") {
		return raw[:len(raw)-1]
	}

	return raw
}

// ConvertType converts any data to type T using JSON marshaling.
//
// It first marshals the input to JSON, then unmarshals it into a value of type T.
// Useful for converting between compatible structs or maps.
func ConvertType[T any](data any) (T, error) {
	var result T
	bytes, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	return result, err
}

// Default returns the zero value of type T.
func Default[T any]() T {
	var t T
	return t
}

// Tap applies the given function to a pointer of the value,
// then returns the (possibly modified) value.
//
// Useful for modifying a copy inline.
func Tap[T any](value T, fn func(*T)) T {
	fn(&value)
	return value
}

// TapFunc wraps a function that takes a pointer to T,
// returning a new function that takes T by value.
//
// Useful for applying changes to a copy of the value.
func TapFunc[T any](fn func(*T)) func(T) T {
	return func(value T) T {
		fn(&value)
		return value
	}
}
