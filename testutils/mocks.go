package testutils

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

// MockFS is a wrapper around an in memory FS with mockable overrides
type MockFS struct {
	MockCreate   func(string) (afero.File, error)
	MockMkdirAll func(string, os.FileMode) error
	MockOpen     func(string) (afero.File, error)
	MockOpenFile func(string, int, os.FileMode) (afero.File, error)
	MockRemove   func(string) error
	MockRename   func(string, string) error
	MockChmod    func(string, os.FileMode) error
	MemFS        afero.Fs
}

func NewMockFS() *MockFS {
	return &MockFS{
		MemFS: afero.NewMemMapFs(),
	}
}

func (m *MockFS) Create(name string) (afero.File, error) {
	if m.MockCreate != nil {
		return m.MockCreate(name)
	}

	return m.MemFS.Create(name)
}

func (m *MockFS) Mkdir(name string, perm os.FileMode) error {
	panic("not implemented")
}

func (m *MockFS) MkdirAll(path string, perm os.FileMode) error {
	if m.MockMkdirAll != nil {
		return m.MockMkdirAll(path, perm)
	}

	return m.MemFS.MkdirAll(path, perm)
}

func (m *MockFS) Open(name string) (afero.File, error) {
	if m.MockOpen != nil {
		return m.MockOpen(name)
	}

	return m.MemFS.Open(name)
}

func (m *MockFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if m.MockOpenFile != nil {
		return m.MockOpenFile(name, flag, perm)
	}

	return m.MemFS.OpenFile(name, flag, perm)
}

func (m *MockFS) Remove(name string) error {
	if m.MockRemove != nil {
		return m.MockRemove(name)
	}

	return m.MemFS.Remove(name)
}

func (m *MockFS) RemoveAll(path string) error {
	panic("not implemented")
}

func (m *MockFS) Rename(before, after string) error {
	if m.MockRename != nil {
		return m.MockRename(before, after)
	}

	return m.MemFS.Rename(before, after)
}

func (m *MockFS) Stat(name string) (os.FileInfo, error) {
	return m.MemFS.Stat(name)
}

func (m *MockFS) Name() string {
	panic("not implemented")
}

func (m *MockFS) Chown(name string, uid, gid int) error {
	panic("not implemented")
}

func (m *MockFS) Chmod(name string, mode os.FileMode) error {
	if m.MockChmod != nil {
		return m.MockChmod(name, mode)
	}

	return m.MemFS.Chmod(name, mode)
}

func (m *MockFS) Chtimes(name string, atime, mtime time.Time) error {
	panic("not implemented")
}

// MockFile is a wrapper around an in memory file with mockable overrides
type MockFile struct {
	MockRead        func([]byte) (int, error)
	MockClose       func() error
	MockWrite       func([]byte) (int, error)
	MockWriteString func(string) (int, error)
	MemFile         afero.File
	contents        []byte
}

func NewMockFile(fs afero.Fs, filename string) *MockFile {
	f, _ := fs.Create(filename)

	return &MockFile{
		MemFile:  f,
		contents: []byte{},
	}
}

func (m *MockFile) Close() error {
	// always close our mock file
	m.MemFile.Close()

	if m.MockClose != nil {
		return m.MockClose()
	}

	return nil
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.MockRead != nil {
		return m.MockRead(p)
	}

	return m.MemFile.Read(p)
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	panic("not implemented")
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	panic("not implemented")
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	m.contents = append(m.contents, p...)
	if m.MockWrite != nil {
		return m.MockWrite(p)
	}

	return m.MemFile.Write(p)
}

func (m *MockFile) WriteAt(p []byte, off int64) (n int, err error) {
	panic("not implemented")
}

func (m *MockFile) Name() string {
	panic("not implemented")
}

func (m *MockFile) Readdir(count int) ([]os.FileInfo, error) {
	return m.MemFile.Readdir(count)
}

func (m *MockFile) Readdirnames(n int) ([]string, error) {
	panic("not implemented")
}

func (m *MockFile) Stat() (os.FileInfo, error) {
	return m.MemFile.Stat()
}

func (m *MockFile) Sync() error {
	panic("not implemented")
}

func (m *MockFile) Truncate(size int64) error {
	panic("not implemented")
}

func (m *MockFile) WriteString(s string) (ret int, err error) {
	m.contents = append(m.contents, []byte(s)...)
	if m.MockWriteString != nil {
		return m.MockWriteString(s)
	}

	return m.MemFile.WriteString(s)
}

func (m *MockFile) Contents() []byte {
	return m.contents
}

func (m *MockFile) String() string {
	return string(m.contents)
}
