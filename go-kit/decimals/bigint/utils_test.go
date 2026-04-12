package bigint

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test cases
	testCases := []struct {
		input     string
		expected  *big.Int
		expectErr bool
	}{
		{input: "123", expected: big.NewInt(123), expectErr: false},
		{input: "-123", expected: big.NewInt(-123), expectErr: false},
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
		expected *big.Int
	}{
		{input: "123", expected: big.NewInt(123)},
		{input: "-123", expected: big.NewInt(-123)},
		{input: "not a number", expected: big.NewInt(0)},
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
	require.Equal(t, big.NewInt(0).Cmp(Zero()), 0)
}

func TestIsZero(t *testing.T) {
	// Test cases
	testCases := []struct {
		input    *big.Int
		expected bool
	}{
		{input: big.NewInt(0), expected: true},
		{input: big.NewInt(1), expected: false},
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
		input0   *big.Int
		input1   *big.Int
		expected bool
	}{
		{input0: big.NewInt(0), input1: big.NewInt(0), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(1), expected: true},
		{input0: big.NewInt(-1), input1: big.NewInt(-1), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(-1), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(0), expected: false},
		{input0: big.NewInt(-1), input1: big.NewInt(0), expected: false},
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
		input0   *big.Int
		input1   *big.Int
		expected bool
	}{
		{input0: big.NewInt(0), input1: big.NewInt(0), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(1), expected: false},
		{input0: big.NewInt(-1), input1: big.NewInt(-1), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(-1), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(0), expected: true},
		{input0: big.NewInt(-1), input1: big.NewInt(0), expected: false},
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
		input0   *big.Int
		input1   *big.Int
		expected bool
	}{
		{input0: big.NewInt(0), input1: big.NewInt(0), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(1), expected: true},
		{input0: big.NewInt(-1), input1: big.NewInt(-1), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(-1), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(0), expected: true},
		{input0: big.NewInt(-1), input1: big.NewInt(0), expected: false},
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
		input0   *big.Int
		input1   *big.Int
		expected bool
	}{
		{input0: big.NewInt(0), input1: big.NewInt(0), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(1), expected: false},
		{input0: big.NewInt(-1), input1: big.NewInt(-1), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(-1), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(0), expected: false},
		{input0: big.NewInt(-1), input1: big.NewInt(0), expected: true},
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
		input0   *big.Int
		input1   *big.Int
		expected bool
	}{
		{input0: big.NewInt(0), input1: big.NewInt(0), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(1), expected: true},
		{input0: big.NewInt(-1), input1: big.NewInt(-1), expected: true},
		{input0: big.NewInt(1), input1: big.NewInt(-1), expected: false},
		{input0: big.NewInt(1), input1: big.NewInt(0), expected: false},
		{input0: big.NewInt(-1), input1: big.NewInt(0), expected: true},
	}

	// Run the tests
	for _, tc := range testCases {
		result := IsLessThanOrEqual(tc.input0, tc.input1)
		if result != tc.expected {
			t.Errorf("IsLessThanOrEqual(%v, %v) = %v, want %v", tc.input0, tc.input1, result, tc.expected)
		}
	}
}
