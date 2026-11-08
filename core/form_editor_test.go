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

	// the editor records the file it was given and writes the body to it
	dir := t.TempDir()
	editedPath := filepath.Join(dir, "edited")
	editor := filepath.Join(dir, "editor.sh")
	script := "#!/bin/sh\nprintf '%s' \"$1\" > " + editedPath + "\nprintf 'from the editor\\n' > \"$1\"\n"
	then.Nil(t, os.WriteFile(editor, []byte(script), 0o700))
	t.Setenv("EDITOR", editor)

	reader, writer := then.WithReadWritePipe(t)
	then.DelayWrite(
		t, writer,
		[]byte{5}, // ctrl+e
		[]byte{13},
	)

	prompts := &Prompts{
		Config: &Config{
			VersionExt: "md",
			Body:       BodyConfig{UseBlock: true},
		},
		StdinReader: reader,
		TimeNow:     specificTimeNow,
		Enabled:     true,
	}

	changes, err := prompts.BuildChanges()
	then.Nil(t, err)
	then.Equals(t, "from the editor", changes[0].Body)

	edited, err := os.ReadFile(editedPath)
	then.Nil(t, err)
	then.Equals(t, ".md", filepath.Ext(string(edited)))
}
