package then

import (
	"testing"

	"github.com/spf13/afero"
)

func WithAferoFS() (afero.Fs, afero.Afero) {
    fs := afero.NewMemMapFs()
    afs := afero.Afero{Fs: fs}

    return fs, afs
}

func FileContents(t *testing.T, afs afero.Afero, filepath, contents string) {
    t.Helper()

	bs, err := afs.ReadFile(filepath)
    if err != nil {
        t.Errorf("reading file: '%v'", filepath)
    }

    expected := string(bs)
    Equals(t, expected, contents)
}
