package then

import (
	"errors"
	"testing"
)

func Equals[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected '%v' to equal '%v'", expected, actual)
	}
}

func NotEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if expected == actual {
		t.Errorf("expected '%v' not to equal '%v'", expected, actual)
	}
}

func MapEquals[K, V comparable](t *testing.T, expected, actual map[K]V) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("length of expected does not equal actual: %v != %v", len(expected), len(actual))
	}

	for k := range expected {
		if expected[k] != actual[k] {
			t.Errorf(
				"expected value does not equal actual of key '%v': expected '%v' != '%v'",
				k,
				expected[k],
				actual[k],
			)
		}
	}
}

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

func Nil(t *testing.T, value any) {
	t.Helper()

	if value != nil {
		t.Errorf("expected '%v' to be nil", value)
	}
}

func NotNil(t *testing.T, value any) {
	t.Helper()

	if value == nil {
		t.Errorf("expected '%v' not to be nil", value)
	}
}

func Err(t *testing.T, expected, actual error) {
	t.Helper()

	if !errors.Is(actual, expected) {
		t.Errorf("expected '%v' to be '%v'", expected, actual)
	}
}

func True(t *testing.T, value bool) {
	t.Helper()

	if !value {
		t.Errorf("expected value to be true")
	}
}

func False(t *testing.T, value bool) {
	t.Helper()

	if value {
		t.Errorf("expected value to be false")
	}
}
