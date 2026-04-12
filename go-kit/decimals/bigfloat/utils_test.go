package bigfloat

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test cases
	testCases := []struct {
		input     string
		expected  *big.Float
		expectErr bool
	}{
		{input: "123", expected: big.NewFloat(123.0), expectErr: false},
		{input: "123.5", expected: big.NewFloat(123.5), expectErr: false},
		{input: "not a number", expected: nil, expectErr: true},
	}

	// Run the tests
	for _, tc := range testCases {
		result, ok := Parse(&tc.input)
		if (!ok) != tc.expectErr {
			t.Errorf("Parse(%q) error = %v, expectErr = %v", tc.input, ok, tc.expectErr)
			continue
		}
		if tc.expected == nil {
			require.Nilf(t, result, "Parse(%q)", tc.input)
			continue
		}

		require.NotNil(t, result)
		if result.Cmp(tc.expected) != 0 {
			t.Errorf("Parse(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestSafeParse(t *testing.T) {
	// Test cases
	testCases := []struct {
		input    string
		expected *big.Float
	}{
		{input: "123", expected: big.NewFloat(123)},
		{input: "123.5", expected: big.NewFloat(123.5)},
		{input: "not a number", expected: big.NewFloat(0)},
	}

	// Run the tests
	for _, tc := range testCases {
		result := SafeParse(&tc.input)
		if result.Cmp(tc.expected) != 0 {
			t.Errorf("SafeParse(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestZero(t *testing.T) {
	require.Equal(t, big.NewFloat(0).Cmp(Zero()), 0)
}

func TestIsZero(t *testing.T) {
	// Test cases
	testCases := []struct {
		input    *big.Float
		expected bool
	}{
		{input: big.NewFloat(0), expected: true},
		{input: big.NewFloat(1), expected: false},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsZero(tc.input)
		if result != tc.expected {
			t.Errorf("IsZero(%v) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsEqual(t *testing.T) {
	// Test cases
	testCases := []struct {
		input0   *big.Float
		input1   *big.Float
		expected bool
	}{
		{input0: big.NewFloat(0), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(1), expected: true},
		{input0: big.NewFloat(-1), input1: big.NewFloat(-1), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(-1), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(0), expected: false},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsEqual(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsEqual(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}

func TestIsGreaterThan(t *testing.T) {
	// Test cases
	testCases := []struct {
		input0   *big.Float
		input1   *big.Float
		expected bool
	}{
		{input0: big.NewFloat(0), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(1), expected: false},
		{input0: big.NewFloat(-1), input1: big.NewFloat(-1), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(-1), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(0), expected: false},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsGreaterThan(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsGreaterThan(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}

func TestIsGreaterThanOrEqual(t *testing.T) {
	// Test cases
	testCases := []struct {
		input0   *big.Float
		input1   *big.Float
		expected bool
	}{
		{input0: big.NewFloat(0), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(1), expected: true},
		{input0: big.NewFloat(-1), input1: big.NewFloat(-1), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(-1), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: true},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsGreaterThanOrEqual(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsGreaterThanOrEqual(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}

func TestIsLessThan(t *testing.T) {
	// Test cases
	testCases := []struct {
		input0   *big.Float
		input1   *big.Float
		expected bool
	}{
		{input0: big.NewFloat(0), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(1), expected: false},
		{input0: big.NewFloat(-1), input1: big.NewFloat(-1), expected: false},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(-1), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: false},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsLessThan(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsLessThan(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}

func TestIsLessThanOrEqual(t *testing.T) {
	// Test cases
	testCases := []struct {
		input0   *big.Float
		input1   *big.Float
		expected bool
	}{
		{input0: big.NewFloat(0), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(1), expected: true},
		{input0: big.NewFloat(-1), input1: big.NewFloat(-1), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: true},
		{input0: big.NewFloat(1), input1: big.NewFloat(-1), expected: false},
		{input0: big.NewFloat(1), input1: big.NewFloat(0), expected: false},
		{input0: big.NewFloat(-1), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(0), expected: true},
		{input0: big.NewFloat(-1.2), input1: big.NewFloat(-1.2), expected: true},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsLessThanOrEqual(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsLessThanOrEqual(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}

func TestConvertDecimal(t *testing.T) {
	val, _ := ConvertDecimals(New(1000000.0), 0, 6).Float64()
	require.Equal(t, 1.0, val)
}
