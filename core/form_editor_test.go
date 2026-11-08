//go:build !race

package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/miniscruff/changie/then"
)

// Opening the editor needs to cancel reading input, which is not possible for
// piped input in race builds, see form_input_race.go.
func TestAskPromptsBodyWithCtrlE(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as the editor")
	}

	// the editor writes the path of the file it was given as the body
	script := filepath.Join(t.TempDir(), "editor.sh")
	then.Nil(t, os.WriteFile(script, []byte("printf '%s\\n' \"$1\" > \"$1\"\n"), 0o600))
	t.Setenv("EDITOR", "sh "+script)

	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{5}, // ctrl+e
		[]byte{13},
	)

	prompts := &Prompts{
		Config: &Config{
			VersionExt: "rst",
			Body:       BodyConfig{UseBlock: true},
		},
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     true,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)

	// the body is the trimmed editor text, the temp file using our extension
	then.Equals(t, ".rst", filepath.Ext(changes[0].Body))
	then.True(t, filepath.IsAbs(changes[0].Body))
}
