//go:build race

package core

import (
	"io"
	"os"

	"github.com/charmbracelet/x/term"
)

// formInput returns the input for the huh form when built with -race.
//
// bubbletea reads files with a cancel reader, which it can close while its
// read loop is still running when a form ends. This is a data race in
// bubbletea that -race reports for every test using a pipe as input.
// To keep the race detector useful for our own code, pipes are read as a
// plain reader in race builds. Plain readers can not be cancelled, so opening
// an editor with ctrl+e does not work with piped input in race builds.
func formInput(input io.Reader) io.Reader {
	file, ok := input.(*os.File)
	if !ok || term.IsTerminal(file.Fd()) {
		return input
	}

	return struct{ io.Reader }{file}
}
