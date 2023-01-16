package then

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/miniscruff/changie/shared"
	"github.com/miniscruff/changie/testutils"
)

func WithAferoFS() (afero.Fs, afero.Afero) {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}

	return fs, afs
}

type Saver interface {
	Save(shared.WriteFiler) error
}

func WithAferoFSConfig(t *testing.T, cfg Saver) (*testutils.MockFS, afero.Afero) {
	fs := testutils.NewMockFS()
	afs := afero.Afero{Fs: fs}

	err := cfg.Save(afs.WriteFile)
	Nil(t, err)

	return fs, afs
}

func CreateFile(t *testing.T, afs afero.Afero, paths ...string) {
	fullPath := filepath.Join(paths...)
	_, err := afs.Create(fullPath)
	Nil(t, err)
}

func WriteFile(t *testing.T, afs afero.Afero, data []byte, paths ...string) {
	fullPath := filepath.Join(paths...)
	f, err := afs.Create(fullPath)
	Nil(t, err)

	_, err = f.Write(data)
	Nil(t, err)
}

func FileContents(t *testing.T, afs afero.Afero, contents string, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	bs, err := afs.ReadFile(fullPath)
	if err != nil {
		t.Errorf("reading file: '%v'", fullPath)
	}

	expected := string(bs)
	Equals(t, expected, contents)
}
