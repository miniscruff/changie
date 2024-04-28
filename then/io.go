package then

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type Saver interface {
	Save() error
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

func WithTempDirConfig(t *testing.T, cfg Saver) {
	t.Helper()
	WithTempDir(t)

	err := cfg.Save()
	Nil(t, err)
}

func CreateFile(t *testing.T, paths ...string) {
	WriteFile(t, nil, paths...)
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

// FileNotExists checks whether a file does not exist at the path defined.
// Paths will be joined together using `filepath.Join`.
func FileNotExists(t *testing.T, paths ...string) {
	t.Helper()

	fullPath := filepath.Join(paths...)

	_, err := os.Stat(fullPath)
	if err == nil {
		t.Logf("expected file '%v' not to exist", fullPath)
		t.FailNow()
	}

	if !os.IsNotExist(err) {
		t.Logf("expected to get is note exist error at '%v': %v", fullPath, err)
		t.FailNow()
	}
}

// FileContents will check the contents of a file.
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
	err := os.MkdirAll(filepath.Dir(fullPath), 0777)
	Nil(t, err)

	f, err := os.Create(fullPath)
	Nil(t, err)

	defer f.Close()

	if len(data) > 0 {
		_, err = f.Write(data)
		Nil(t, err)
	}
}

func WriteFileTo(t *testing.T, w io.WriterTo, paths ...string) {
	t.Helper()

	cd, _ := os.Getwd()
	t.Log("cd", cd)

	fullPath := filepath.Join(paths...)
	err := os.MkdirAll(filepath.Dir(fullPath), 0777)
	Nil(t, err)

	f, err := os.Create(fullPath)
	Nil(t, err)

	defer f.Close()

	_, err = w.WriteTo(f)
	Nil(t, err)
}

func DirectoryFileCount(t *testing.T, count int, paths ...string) {
	t.Helper()

	infos, err := os.ReadDir(filepath.Join(paths...))
	Nil(t, err)

	if count != len(infos) {
		t.Logf("expected %v files but found %v", count, len(infos))

		for _, fi := range infos {
			t.Log("file:", fi.Name())
		}

		t.FailNow()
	}
}

type MockDirEntry struct {
	MockName    string
	MockIsDir   bool
	MockInfo    os.FileInfo
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
