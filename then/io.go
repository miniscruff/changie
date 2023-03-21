package then

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/miniscruff/changie/shared"
)

type Saver interface {
	Save(shared.WriteFiler) error
}

// WithTempDir creates a temporary directory and moves our working directory to it.
// At the end of our test it will remove the directory and revert our working directory
// to the original value.
func WithTempDir(t *testing.T) {
	t.Helper()

	startDir, err := os.Getwd()
	Nil(t, err)

	tempDir := t.TempDir()

	err = os.Chdir(tempDir)
	Nil(t, err)

	t.Cleanup(func() {
		err := os.Chdir(startDir)
		Nil(t, err)
	})
}

func WithTempDirConfig(t *testing.T, cfg Saver, existingPath ...string) {
	t.Helper()
	WithTempDir(t)

	var err error
	if len(existingPath) > 0 {
		err = os.MkdirAll(filepath.Join(existingPath...), 0755)
		Nil(t, err)
	}

	err = cfg.Save(os.WriteFile)
	Nil(t, err)
}

func CreateFile(t *testing.T, paths ...string) {
    t.Helper()
	fullPath := filepath.Join(paths...)
	f, err := os.Create(fullPath)
	Nil(t, err)

    defer f.Close()
}

// FileExists checks whether a file exists at the path defined.
// Paths will be joined together using `filepath.Join`.
func FileExists(t *testing.T, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	_, err := os.Stat(fullPath)
	if err == nil {
		return
	}

	switch {
	case os.IsNotExist(err):
		t.Logf("expected file '%v' to exist", fullPath)
		t.FailNow()
	default:
		t.Logf("error getting info for '%v': %v", fullPath, err)
		t.FailNow()
	}
}

// FileContentsOS will check the contents of a file.
// NOTE: Remove the OS when Afero is removed
func FileContents(t *testing.T, contents string, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	bs, err := os.ReadFile(fullPath)
	if err != nil {
		t.Logf("reading file: '%v'", fullPath)
		t.FailNow()
	}

	expected := string(bs)
	Equals(t, expected, contents)
}

func WriteFile(t *testing.T, data []byte, paths ...string) {
    t.Helper()

	fullPath := filepath.Join(paths...)
	f, err := os.Create(fullPath)
	Nil(t, err)
    defer f.Close()

	_, err = f.Write(data)
	Nil(t, err)
}

type MockDirEntry struct {
    MockName string
    MockIsDir bool
    MockInfo os.FileInfo
    MockInfoErr error
}

var _ os.DirEntry = (*MockDirEntry)(nil)

func (m *MockDirEntry) Name() string {
    return m.MockName
}

func (m *MockDirEntry) IsDir() bool {
    return m.MockIsDir
}

func (m *MockDirEntry) Type() os.FileMode {
    if m.MockIsDir {
        return os.ModeDir
    }

    return os.ModePerm
}

func (m *MockDirEntry) Info() (os.FileInfo, error) {
    if m.MockInfoErr != nil {
        return nil, m.MockInfoErr
    }

    return m.MockInfo, nil
}

// MockFileInfo is a simple struct to fake the `os.FileInfo`
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
