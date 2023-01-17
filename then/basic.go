package then

import (
	"errors"
	"strings"
	"testing"
)

// Equals compares two values, in some rare cases due to generic limitations
// you may have to use `reflect.DeepEquals` instead.
func Equals[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected '%v' to equal '%v'", expected, actual)
	}
}

// NotEquals compares two values are not equal
func NotEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected == actual {
		t.Errorf("expected '%v' not to equal '%v'", expected, actual)
	}
}

// MapEquals compares two map values for same length and same key+value pairs.
func MapEquals[M1, M2 ~map[K]V, K, V comparable](t *testing.T, expected M1, actual M2) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("length of expected does not equal actual: %v != %v", len(expected), len(actual))
	}

	for k, v1 := range expected {
		v2, ok := actual[k]
		if !ok {
			t.Errorf(
				"actual map is missing key '%v', expected value: '%v'",
				k,
				v1,
			)
		}

		if v1 != v2 {
			t.Errorf(
				"expected value does not equal actual of key '%v': expected '%v' != '%v'",
				k,
				v1,
				v2,
			)
		}
	}
}

// SliceEquals compares two slices for same length and same values at each index.
func SliceEquals[T comparable](t *testing.T, expected, actual []T) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("length of expected does not equal actual: %v != %v", len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			t.Errorf(
				"expected value does not equal actual at index %v: expected '%v' != '%v'",
				i,
				expected[i],
				actual[i],
			)
		}
	}
}

// Nil compares a value to nil, in some cases you may need to do `Equals(t, value, nil)`
func Nil(t *testing.T, value any) {
	t.Helper()

	if value != nil {
		t.Errorf("expected '%v' to be nil", value)
	}
}

// NotNil compares a value is not nil.
func NotNil(t *testing.T, value any) {
	t.Helper()

	if value == nil {
		t.Errorf("expected '%v' not to be nil", value)
	}
}

// Err checks if our actual error is the expected error or wrapped in the expected error.
func Err(t *testing.T, expected, actual error) {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Errorf("expected '%v' to be '%v'", expected, actual)
	}
}

// Panic checks if our func would panic.
func Panic(t *testing.T, f func()) {
	defer func() {
		// we don't care what the value is, only that we had to recover
		_ = recover()
	}()

	f()
	t.Error("expected func to panic")
}

// True checks if a value is true.
func True(t *testing.T, value bool) {
	t.Helper()

	if !value {
		t.Error("expected value to be true")
	}
}

// False checks if a value is false.
func False(t *testing.T, value bool) {
	t.Helper()

	if value {
		t.Error("expected value to be false")
	}
}

// Contains checks if our substring is contained in the full string
func Contains(t *testing.T, sub, full string) {
	t.Helper()

	if !strings.Contains(full, sub) {
		t.Errorf("expected '%v' to be in '%v'", sub, full)
	}
}
