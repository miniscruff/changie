package then

import (
	"os"
	"testing"
	"time"

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

type MockFileInfo struct {
	MockName  string
	MockIsDir bool
}

var _ os.FileInfo = (*MockFileInfo)(nil)

func (m *MockFileInfo) Name() string {
	return m.MockName
}

func (m *MockFileInfo) IsDir() bool {
	return m.MockIsDir
}

func (m *MockFileInfo) Size() int64 {
	return 0
}

func (m *MockFileInfo) Mode() os.FileMode {
	return os.ModeAppend
}

func (m *MockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (m *MockFileInfo) Sys() any {
	return nil
}
