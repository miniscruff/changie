//go:build !race

package core

import "io"

// formInput returns the input for the huh form, see form_input_race.go.
func formInput(input io.Reader) io.Reader {
	return input
}
