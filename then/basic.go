package then

import (
	"errors"
	"strings"
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

func Panic(t *testing.T, f func()) {
    defer func() {
        recover()
    }()

    f()
    t.Error("expected func to panic")
}

func True(t *testing.T, value bool) {
	t.Helper()

	if !value {
		t.Error("expected value to be true")
	}
}

func False(t *testing.T, value bool) {
	t.Helper()

	if value {
		t.Error("expected value to be false")
	}
}

func Contains(t *testing.T, sub, full string) {
	t.Helper()

	if !strings.Contains(full, sub) {
		t.Errorf("expected '%v' to be in '%v'", sub, full)
	}
}
